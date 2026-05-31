package claude

import (
	"net/http"
	"net/http/httptest"
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
	ctx.Request.Header.Set("X-Oneapi-Request-Id", "req-test")

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

	if got := header.Get("User-Agent"); got != "claude-code/2.1.158" {
		t.Fatalf("expected claude-code user-agent, got %q", got)
	}
	if got := header.Get("x-client-app"); got != "claude-code" {
		t.Fatalf("expected x-client-app header, got %q", got)
	}
	if got := header.Get("x-app"); got != "claude-code" {
		t.Fatalf("expected x-app header, got %q", got)
	}
	if got := header.Get("X-Claude-Code-Session-Id"); got != "req-test" {
		t.Fatalf("expected session id from request, got %q", got)
	}
	if got := header.Get("x-api-key"); got != "test-key" {
		t.Fatalf("expected x-api-key, got %q", got)
	}
}
