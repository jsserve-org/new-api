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

func TestSetupRequestHeaderOpenAIRequestProfiles(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name      string
		settings  dto.ChannelOtherSettings
		wantUA    string
		wantTitle string
		wantSrc   string
	}{
		{
			name: "legacy opencode toggle",
			settings: dto.ChannelOtherSettings{
				OpenAILikeOpenCode: true,
			},
			wantUA:    "opencode",
			wantTitle: "opencode",
			wantSrc:   "opencode",
		},
		{
			name: "codex profile",
			settings: dto.ChannelOtherSettings{
				OpenAIRequestProfile: dto.OpenAIRequestProfileCodex,
			},
			wantUA:    "codex_cli_rs",
			wantTitle: "codex",
			wantSrc:   "codex",
		},
		{
			name: "hermes agent profile",
			settings: dto.ChannelOtherSettings{
				OpenAIRequestProfile: dto.OpenAIRequestProfileHermesAgent,
			},
			wantUA:    "hermes-agent",
			wantTitle: "Hermes Agent",
			wantSrc:   "hermes-agent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = httptest.NewRequest(http.MethodPost, "/v1/chat/completions", nil)
			ctx.Request.Header.Set("Content-Type", "application/json")

			header := http.Header{}
			header.Set("User-Agent", "new-api-test-client")
			header.Set("X-Title", "new-api")
			info := &relaycommon.RelayInfo{
				ChannelMeta: &relaycommon.ChannelMeta{
					ChannelType:          constant.ChannelTypeOpenAI,
					ApiKey:               "test-key",
					ChannelOtherSettings: tt.settings,
				},
			}

			adaptor := &Adaptor{}
			if err := adaptor.SetupRequestHeader(ctx, &header, info); err != nil {
				t.Fatalf("SetupRequestHeader returned error: %v", err)
			}

			if got := header.Get("User-Agent"); got != tt.wantUA {
				t.Fatalf("expected user-agent %q, got %q", tt.wantUA, got)
			}
			if got := header.Get("X-Title"); got != tt.wantTitle {
				t.Fatalf("expected x-title %q, got %q", tt.wantTitle, got)
			}
			if got := header.Get("X-Source"); got != tt.wantSrc {
				t.Fatalf("expected x-source %q, got %q", tt.wantSrc, got)
			}
			if got := header.Get("Authorization"); got != "Bearer test-key" {
				t.Fatalf("expected authorization header, got %q", got)
			}
		})
	}
}
