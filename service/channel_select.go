package service

import (
	"errors"
	"math/rand"
	"sort"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/gin-gonic/gin"
)

const AutoModelName = "auto"
const AutoModelPrefix = "auto-"

type RetryParam struct {
	Ctx           *gin.Context
	TokenGroup    string
	ModelName     string
	AllowedModels map[string]bool
	Retry         *int
	resetNextTry  bool
}

func (p *RetryParam) GetRetry() int {
	if p.Retry == nil {
		return 0
	}
	return *p.Retry
}

func (p *RetryParam) SetRetry(retry int) {
	p.Retry = &retry
}

func (p *RetryParam) IncreaseRetry() {
	if p.resetNextTry {
		p.resetNextTry = false
		return
	}
	if p.Retry == nil {
		p.Retry = new(int)
	}
	*p.Retry++
}

func (p *RetryParam) ResetRetryNextTry() {
	p.resetNextTry = true
}

func isAutoModelName(modelName string) bool {
	return modelName == AutoModelName || strings.HasPrefix(modelName, AutoModelPrefix)
}

func IsAutoRoutingModel(modelName string) bool {
	return isAutoModelName(modelName)
}

func getAutoRouteKeyword(modelName string) string {
	if !strings.HasPrefix(modelName, AutoModelPrefix) {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(strings.TrimPrefix(modelName, AutoModelPrefix)))
}

func matchesAutoRoute(candidate string, routeModel string) bool {
	keyword := getAutoRouteKeyword(routeModel)
	if keyword == "" {
		return true
	}
	candidate = strings.ToLower(strings.TrimSpace(candidate))
	formatted := strings.ToLower(strings.TrimSpace(ratio_setting.FormatMatchingModelName(candidate)))
	return strings.Contains(candidate, keyword) || strings.Contains(formatted, keyword)
}

func allowedAutoModel(candidate string, allowed map[string]bool, routeModel string) bool {
	if candidate == "" || isAutoModelName(candidate) {
		return false
	}
	if !matchesAutoRoute(candidate, routeModel) {
		return false
	}
	if len(allowed) == 0 {
		return true
	}
	if allowed[AutoModelName] {
		return true
	}
	if allowed[candidate] {
		return true
	}
	_, ok := allowed[ratio_setting.FormatMatchingModelName(candidate)]
	return ok
}

func getAutoCandidateModels(group string, allowed map[string]bool, routeModel string) []string {
	models := model.GetGroupEnabledModels(group)
	if routeModel != AutoModelName && !isAutoModelName(routeModel) {
		return []string{}
	}
	
	candidates := make([]string, 0, len(models))
	for _, modelName := range models {
		if allowedAutoModel(modelName, allowed, routeModel) && !common.StringsContains(candidates, modelName) {
			candidates = append(candidates, modelName)
		}
	}
	sort.Strings(candidates)
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})
	return candidates
}

func getRandomSatisfiedChannelForAnyModel(group string, allowed map[string]bool, routeModel string, retry int) (*model.Channel, string, error) {
	candidates := getAutoCandidateModels(group, allowed, routeModel)
	for _, candidate := range candidates {
		channel, err := model.GetRandomSatisfiedChannel(group, candidate, retry)
		if err != nil {
			return nil, candidate, err
		}
		if channel != nil {
			return channel, candidate, nil
		}
	}
	return nil, "", nil
}

// CacheGetRandomSatisfiedChannel tries to get a random channel that satisfies the requirements.
// 尝试获取一个满足要求的随机渠道。
//
// For "auto" tokenGroup with cross-group Retry enabled:
// 对于启用了跨分组重试的 "auto" tokenGroup：
//
//   - Each group will exhaust all its priorities before moving to the next group.
//     每个分组会用完所有优先级后才会切换到下一个分组。
//
//   - Uses ContextKeyAutoGroupIndex to track current group index.
//     使用 ContextKeyAutoGroupIndex 跟踪当前分组索引。
//
//   - Uses ContextKeyAutoGroupRetryIndex to track the global Retry count when current group started.
//     使用 ContextKeyAutoGroupRetryIndex 跟踪当前分组开始时的全局重试次数。
//
//   - priorityRetry = Retry - startRetryIndex, represents the priority level within current group.
//     priorityRetry = Retry - startRetryIndex，表示当前分组内的优先级级别。
//
//   - When GetRandomSatisfiedChannel returns nil (priorities exhausted), moves to next group.
//     当 GetRandomSatisfiedChannel 返回 nil（优先级用完）时，切换到下一个分组。
//
// Example flow (2 groups, each with 2 priorities, RetryTimes=3):
// 示例流程（2个分组，每个有2个优先级，RetryTimes=3）：
//
//	Retry=0: GroupA, priority0 (startRetryIndex=0, priorityRetry=0)
//	         分组A, 优先级0
//
//	Retry=1: GroupA, priority1 (startRetryIndex=0, priorityRetry=1)
//	         分组A, 优先级1
//
//	Retry=2: GroupA exhausted → GroupB, priority0 (startRetryIndex=2, priorityRetry=0)
//	         分组A用完 → 分组B, 优先级0
//
//	Retry=3: GroupB, priority1 (startRetryIndex=2, priorityRetry=1)
//	         分组B, 优先级1
func CacheGetRandomSatisfiedChannel(param *RetryParam) (*model.Channel, string, error) {
	var channel *model.Channel
	var err error
	selectGroup := param.TokenGroup
	userGroup := common.GetContextKeyString(param.Ctx, constant.ContextKeyUserGroup)

	if param.TokenGroup == "auto" {
		if len(setting.GetAutoGroups()) == 0 {
			return nil, selectGroup, errors.New("auto groups is not enabled")
		}
		autoGroups := GetUserAutoGroup(userGroup)

		// startGroupIndex: the group index to start searching from
		// startGroupIndex: 开始搜索的分组索引
		startGroupIndex := 0
		crossGroupRetry := common.GetContextKeyBool(param.Ctx, constant.ContextKeyTokenCrossGroupRetry)

		if lastGroupIndex, exists := common.GetContextKey(param.Ctx, constant.ContextKeyAutoGroupIndex); exists {
			if idx, ok := lastGroupIndex.(int); ok {
				startGroupIndex = idx
			}
		}

		for i := startGroupIndex; i < len(autoGroups); i++ {
			autoGroup := autoGroups[i]
			// Calculate priorityRetry for current group
			// 计算当前分组的 priorityRetry
			priorityRetry := param.GetRetry()
			// If moved to a new group, reset priorityRetry and update startRetryIndex
			// 如果切换到新分组，重置 priorityRetry 并更新 startRetryIndex
			if i > startGroupIndex {
				priorityRetry = 0
			}
			logger.LogDebug(param.Ctx, "Auto selecting group: %s, priorityRetry: %d", autoGroup, priorityRetry)

			selectedModel := param.ModelName
			if isAutoModelName(param.ModelName) {
				channel, selectedModel, err = getRandomSatisfiedChannelForAnyModel(autoGroup, param.AllowedModels, param.ModelName, priorityRetry)
				if err != nil {
					return nil, selectGroup, err
				}
			} else {
				channel, _ = model.GetRandomSatisfiedChannel(autoGroup, param.ModelName, priorityRetry)
			}
			if channel == nil {
				// Current group has no available channel for this model, try next group
				// 当前分组没有该模型的可用渠道，尝试下一个分组
				logger.LogDebug(param.Ctx, "No available channel in group %s for model %s at priorityRetry %d, trying next group", autoGroup, param.ModelName, priorityRetry)
				// 重置状态以尝试下一个分组
				common.SetContextKey(param.Ctx, constant.ContextKeyAutoGroupIndex, i+1)
				common.SetContextKey(param.Ctx, constant.ContextKeyAutoGroupRetryIndex, 0)
				// Reset retry counter so outer loop can continue for next group
				// 重置重试计数器，以便外层循环可以为下一个分组继续
				param.SetRetry(0)
				continue
			}
			if isAutoModelName(param.ModelName) && selectedModel != "" {
				param.ModelName = selectedModel
			}
			common.SetContextKey(param.Ctx, constant.ContextKeyAutoGroup, autoGroup)
			selectGroup = autoGroup
			logger.LogDebug(param.Ctx, "Auto selected group: %s", autoGroup)

			// Prepare state for next retry
			// 为下一次重试准备状态
			if crossGroupRetry && priorityRetry >= common.RetryTimes {
				// Current group has exhausted all retries, prepare to switch to next group
				// This request still uses current group, but next retry will use next group
				// 当前分组已用完所有重试次数，准备切换到下一个分组
				// 本次请求仍使用当前分组，但下次重试将使用下一个分组
				logger.LogDebug(param.Ctx, "Current group %s retries exhausted (priorityRetry=%d >= RetryTimes=%d), preparing switch to next group for next retry", autoGroup, priorityRetry, common.RetryTimes)
				common.SetContextKey(param.Ctx, constant.ContextKeyAutoGroupIndex, i+1)
				// Reset retry counter so outer loop can continue for next group
				// 重置重试计数器，以便外层循环可以为下一个分组继续
				param.SetRetry(0)
				param.ResetRetryNextTry()
			} else {
				// Stay in current group, save current state
				// 保持在当前分组，保存当前状态
				common.SetContextKey(param.Ctx, constant.ContextKeyAutoGroupIndex, i)
			}
			break
		}
	} else {
		if isAutoModelName(param.ModelName) {
			selectedModel := ""
			channel, selectedModel, err = getRandomSatisfiedChannelForAnyModel(param.TokenGroup, param.AllowedModels, param.ModelName, param.GetRetry())
			if err != nil {
				return nil, param.TokenGroup, err
			}
			if selectedModel != "" {
				param.ModelName = selectedModel
			}
		} else {
			channel, err = model.GetRandomSatisfiedChannel(param.TokenGroup, param.ModelName, param.GetRetry())
			if err != nil {
				return nil, param.TokenGroup, err
			}
		}
	}
	return channel, selectGroup, nil
}
