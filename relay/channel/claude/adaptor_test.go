package claude

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
)

func TestSetupRequestHeaderClaudeCodeProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
	ctx.Request.Header.Set("Content-Type", "application/json")

	header := http.Header{}
	info := &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelType: constant.ChannelTypeAnthropic,
			ApiKey:      "test-key",
			ChannelOtherSettings: dto.ChannelOtherSettings{
				OpenAIRequestProfile: dto.OpenAIRequestProfileCCSwitch,
			},
		},
	}

	adaptor := &Adaptor{}
	if err := adaptor.SetupRequestHeader(ctx, &header, info); err != nil {
		t.Fatalf("SetupRequestHeader returned error: %v", err)
	}

	if got := header.Get("User-Agent"); got != "claude-cli/2.1.158 (external, sdk-cli)" {
		t.Fatalf("expected claude cli user-agent, got %q", got)
	}
	if got := header.Get("x-app"); got != "cli" {
		t.Fatalf("expected x-app header, got %q", got)
	}
	if got := header.Get("X-Stainless-Lang"); got != "js" {
		t.Fatalf("expected stainless lang header, got %q", got)
	}
	if got := header.Get("anthropic-dangerous-direct-browser-access"); got != "true" {
		t.Fatalf("expected dangerous direct browser access header, got %q", got)
	}
	if got := header.Get("anthropic-beta"); !strings.Contains(got, "claude-code-20250219") {
		t.Fatalf("expected claude-code beta header, got %q", got)
	}
	if got := header.Get("X-Claude-Code-Session-Id"); got == "" {
		t.Fatal("expected generated session id")
	}
	if got := header.Get("x-api-key"); got != "test-key" {
		t.Fatalf("expected x-api-key, got %q", got)
	}
}

func TestGetRequestURLClaudeCodeProfileAppendsBetaQuery(t *testing.T) {
	info := &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelBaseUrl: "https://agentrouter.org",
			ChannelOtherSettings: dto.ChannelOtherSettings{
				OpenAIRequestProfile: dto.OpenAIRequestProfileCCSwitch,
			},
		},
	}

	got, err := (&Adaptor{}).GetRequestURL(info)
	if err != nil {
		t.Fatalf("GetRequestURL returned error: %v", err)
	}
	if got != "https://agentrouter.org/v1/messages?beta=true" {
		t.Fatalf("unexpected request URL: %q", got)
	}
}
