package openai

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
)

func TestSetupRequestHeaderOpenAILikeOpenCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
	ctx.Request.Header.Set("Content-Type", "application/json")

	header := http.Header{}
	info := &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelType: constant.ChannelTypeOpenAI,
			ApiKey:      "test-key",
			ChannelOtherSettings: dto.ChannelOtherSettings{
				OpenAILikeOpenCode: true,
			},
		},
	}

	adaptor := &Adaptor{}
	if err := adaptor.SetupRequestHeader(ctx, &header, info); err != nil {
		t.Fatalf("SetupRequestHeader returned error: %v", err)
	}

	if got := header.Get("User-Agent"); got != "opencode" {
		t.Fatalf("expected opencode user-agent, got %q", got)
	}
	if got := header.Get("HTTP-Referer"); got != "https://opencode.ai/" {
		t.Fatalf("expected opencode referer, got %q", got)
	}
	if got := header.Get("X-Title"); got != "opencode" {
		t.Fatalf("expected opencode title, got %q", got)
	}
	if got := header.Get("X-Source"); got != "opencode" {
		t.Fatalf("expected opencode source, got %q", got)
	}
	if got := header.Get("Authorization"); got != "Bearer test-key" {
		t.Fatalf("expected authorization header, got %q", got)
	}
}
