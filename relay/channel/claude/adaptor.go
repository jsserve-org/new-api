package claude

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/relay/channel"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/setting/model_setting"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

type Adaptor struct {
}

func (a *Adaptor) ConvertGeminiRequest(*gin.Context, *relaycommon.RelayInfo, *dto.GeminiChatRequest) (any, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.ClaudeRequest) (any, error) {
	return request, nil
}

func (a *Adaptor) ConvertAudioRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.AudioRequest) (io.Reader, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.ImageRequest) (any, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) Init(info *relaycommon.RelayInfo) {
}

func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	requestURL := fmt.Sprintf("%s/v1/messages", info.ChannelBaseUrl)
	if !shouldAppendClaudeBetaQuery(info) {
		return requestURL, nil
	}

	parsedURL, err := url.Parse(requestURL)
	if err != nil {
		return "", err
	}
	query := parsedURL.Query()
	query.Set("beta", "true")
	parsedURL.RawQuery = query.Encode()
	return parsedURL.String(), nil
}

func shouldAppendClaudeBetaQuery(info *relaycommon.RelayInfo) bool {
	if info == nil {
		return false
	}
	if info.IsClaudeBetaQuery {
		return true
	}
	if info.ChannelOtherSettings.ClaudeBetaQuery {
		return true
	}
	return false
}

func CommonClaudeHeadersOperation(c *gin.Context, req *http.Header, info *relaycommon.RelayInfo) {
	// common headers operation
	anthropicBeta := c.Request.Header.Get("anthropic-beta")
	if anthropicBeta != "" {
		req.Set("anthropic-beta", anthropicBeta)
	}
	model_setting.GetClaudeSettings().WriteHeaders(info.OriginModelName, req)
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Header, info *relaycommon.RelayInfo) error {
	channel.SetupApiRequestHeader(info, c, req)
	applyClaudeRequestProfileHeaders(c, req, info.ChannelOtherSettings.OpenAIRequestProfile)
	req.Set("x-api-key", info.ApiKey)
	anthropicVersion := c.Request.Header.Get("anthropic-version")
	if anthropicVersion == "" {
		anthropicVersion = "2023-06-01"
	}
	req.Set("anthropic-version", anthropicVersion)
	CommonClaudeHeadersOperation(c, req, info)
	return nil
}

func applyClaudeRequestProfileHeaders(c *gin.Context, req *http.Header, profile dto.OpenAIRequestProfile) {
	switch profile {
	case dto.OpenAIRequestProfileCCSwitch:
		setClaudeProfileHeader(req, "User-Agent", "claude-code/2.1.158")
		setClaudeProfileHeader(req, "x-app", "claude-code")
		setClaudeProfileHeader(req, "x-client-app", "claude-code")
		setClaudeProfileHeader(req, "X-Claude-Code-Session-Id", claudeCodeSessionID(c))
	case dto.OpenAIRequestProfileOpenCode:
		setClaudeProfileHeader(req, "User-Agent", "opencode")
		setClaudeProfileHeader(req, "HTTP-Referer", "https://opencode.ai/")
		setClaudeProfileHeader(req, "X-Title", "opencode")
		setClaudeProfileHeader(req, "X-Source", "opencode")
	case dto.OpenAIRequestProfileCodex:
		setClaudeProfileHeader(req, "User-Agent", "codex_cli_rs")
		setClaudeProfileHeader(req, "originator", "codex_cli_rs")
	case dto.OpenAIRequestProfilePi:
		setClaudeProfileHeader(req, "User-Agent", "pi-coding-agent")
		setClaudeProfileHeader(req, "originator", "pi")
	case dto.OpenAIRequestProfileOpenClaw:
		setClaudeProfileHeader(req, "User-Agent", "openclaw")
		setClaudeProfileHeader(req, "originator", "openclaw")
	case dto.OpenAIRequestProfileHermesAgent:
		setClaudeProfileHeader(req, "User-Agent", "hermes-agent")
		setClaudeProfileHeader(req, "originator", "hermes_agent")
	}
}

func claudeCodeSessionID(c *gin.Context) string {
	if c == nil {
		return "00000000-0000-4000-8000-000000000000"
	}
	requestID := c.GetString(common.RequestIdKey)
	if requestID == "" && c.Request != nil {
		requestID = c.Request.Header.Get(common.RequestIdKey)
	}
	if requestID == "" {
		return "00000000-0000-4000-8000-000000000000"
	}
	return requestID
}

func setClaudeProfileHeader(req *http.Header, key string, value string) {
	req.Set(key, value)
}

func (a *Adaptor) ConvertOpenAIRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return RequestOpenAI2ClaudeMessage(c, *request)
}

func (a *Adaptor) ConvertRerankRequest(c *gin.Context, relayMode int, request dto.RerankRequest) (any, error) {
	return nil, nil
}

func (a *Adaptor) ConvertEmbeddingRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.EmbeddingRequest) (any, error) {
	//TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) ConvertOpenAIResponsesRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.OpenAIResponsesRequest) (any, error) {
	// TODO implement me
	return nil, errors.New("not implemented")
}

func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (any, error) {
	return channel.DoApiRequest(a, c, info, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *types.NewAPIError) {
	info.FinalRequestRelayFormat = types.RelayFormatClaude
	if info.IsStream {
		return ClaudeStreamHandler(c, resp, info)
	} else {
		return ClaudeHandler(c, resp, info)
	}
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}
