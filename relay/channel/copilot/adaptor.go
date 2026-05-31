package copilot

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/relay/channel"
	"github.com/QuantumNous/new-api/relay/channel/openai"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"

	"github.com/gin-gonic/gin"
)

const (
	copilotTokenURL  = "https://api.github.com/copilot_internal/v2/token"
	tokenCacheTTL    = 25 * time.Minute // tokens typically expire after 30 min
	tokenExchangeTTL = 15 * time.Second
)

type copilotTokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

type cachedToken struct {
	Token     string
	ExpiresAt time.Time
}

var (
	tokenCache   = make(map[string]cachedToken)
	tokenCacheMu sync.RWMutex
)

type Adaptor struct {
	openaiAdaptor openai.Adaptor
}

func (a *Adaptor) Init(info *relaycommon.RelayInfo) {
	a.openaiAdaptor.Init(info)
}

func (a *Adaptor) GetRequestURL(info *relaycommon.RelayInfo) (string, error) {
	return a.openaiAdaptor.GetRequestURL(info)
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return ChannelName
}

// exchangeGitHubToken exchanges a GitHub OAuth token for a GitHub Copilot API token.
func exchangeGitHubToken(ctx context.Context, githubToken string) (string, error) {
	token := strings.TrimSpace(githubToken)
	if token == "" {
		return "", errors.New("copilot channel: empty github token")
	}

	// Check cache first
	tokenCacheMu.RLock()
	if cached, ok := tokenCache[token]; ok && time.Now().Before(cached.ExpiresAt) {
		tokenCacheMu.RUnlock()
		return cached.Token, nil
	}
	tokenCacheMu.RUnlock()

	// Exchange via GitHub API
	exchangeCtx, cancel := context.WithTimeout(ctx, tokenExchangeTTL)
	defer cancel()

	req, err := http.NewRequestWithContext(exchangeCtx, http.MethodGet, copilotTokenURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: tokenExchangeTTL}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		if len(bodyStr) > 500 {
			bodyStr = bodyStr[:500] + "..."
		}
		return "", errors.New("copilot token exchange failed: HTTP " + resp.Status + " - " + bodyStr)
	}

	var tokenResp copilotTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	if tokenResp.Token == "" {
		return "", errors.New("copilot token exchange returned empty token")
	}

	// Calculate cache TTL based on expires_at or fallback to default
	ttl := tokenCacheTTL
	if tokenResp.ExpiresAt > 0 {
		expiry := time.Unix(tokenResp.ExpiresAt, 0)
		remaining := time.Until(expiry)
		if remaining > 0 {
			// Cache for 80% of remaining time
			ttl = time.Duration(float64(remaining) * 0.8)
		}
	}

	// Cache the token
	tokenCacheMu.Lock()
	tokenCache[token] = cachedToken{
		Token:     tokenResp.Token,
		ExpiresAt: time.Now().Add(ttl),
	}
	tokenCacheMu.Unlock()

	return tokenResp.Token, nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Header, info *relaycommon.RelayInfo) error {
	channel.SetupApiRequestHeader(info, c, req)

	key := strings.TrimSpace(info.ApiKey)
	if key == "" {
		return errors.New("copilot channel: api key (github token) is required")
	}

	// Check if the key looks like a raw Copilot token (starts with "tid=")
	// or a GitHub token (starts with "ghp_", "gho_", "github_pat_", "ghu_")
	copilotToken := key
	if isGitHubToken(key) {
		var err error
		copilotToken, err = exchangeGitHubToken(c.Request.Context(), key)
		if err != nil {
			logger.LogError(c.Request.Context(), "[Copilot] token exchange failed: "+err.Error())
			return err
		}
	}

	req.Set("Authorization", "Bearer "+copilotToken)
	req.Set("Content-Type", "application/json")

	// Copilot API requires these headers
	if req.Get("Copilot-Integration-Id") == "" {
		req.Set("Copilot-Integration-Id", "vscode-chat")
	}
	if req.Get("Editor-Version") == "" {
		req.Set("Editor-Version", "vscode/1.95.0")
	}
	if req.Get("User-Agent") == "" {
		req.Set("User-Agent", "GitHubCopilotChat/0.21.0")
	}

	return nil
}

// isGitHubToken checks if the key looks like a GitHub personal access token.
func isGitHubToken(key string) bool {
	key = strings.TrimSpace(key)
	return strings.HasPrefix(key, "ghp_") ||
		strings.HasPrefix(key, "gho_") ||
		strings.HasPrefix(key, "github_pat_") ||
		strings.HasPrefix(key, "ghu_")
}

func (a *Adaptor) ConvertOpenAIRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeneralOpenAIRequest) (any, error) {
	// Copilot doesn't support StreamOptions, clear it
	request.StreamOptions = nil
	return a.openaiAdaptor.ConvertOpenAIRequest(c, info, request)
}

func (a *Adaptor) ConvertRerankRequest(c *gin.Context, relayMode int, request dto.RerankRequest) (any, error) {
	return a.openaiAdaptor.ConvertRerankRequest(c, relayMode, request)
}

func (a *Adaptor) ConvertEmbeddingRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.EmbeddingRequest) (any, error) {
	return a.openaiAdaptor.ConvertEmbeddingRequest(c, info, request)
}

func (a *Adaptor) ConvertAudioRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.AudioRequest) (io.Reader, error) {
	return nil, errors.New("copilot channel: audio endpoint not supported")
}

func (a *Adaptor) ConvertImageRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.ImageRequest) (any, error) {
	return nil, errors.New("copilot channel: image endpoint not supported")
}

func (a *Adaptor) ConvertOpenAIResponsesRequest(c *gin.Context, info *relaycommon.RelayInfo, request dto.OpenAIResponsesRequest) (any, error) {
	return nil, errors.New("copilot channel: responses endpoint not supported")
}

func (a *Adaptor) ConvertClaudeRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.ClaudeRequest) (any, error) {
	return nil, errors.New("copilot channel: /v1/messages endpoint not supported")
}

func (a *Adaptor) ConvertGeminiRequest(c *gin.Context, info *relaycommon.RelayInfo, request *dto.GeminiChatRequest) (any, error) {
	return nil, errors.New("copilot channel: gemini endpoint not supported")
}

func (a *Adaptor) DoRequest(c *gin.Context, info *relaycommon.RelayInfo, requestBody io.Reader) (any, error) {
	return a.openaiAdaptor.DoRequest(c, info, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo) (usage any, err *types.NewAPIError) {
	if info.IsStream {
		return openai.OaiStreamHandler(c, info, resp)
	}
	return openai.OpenaiHandler(c, info, resp)
}