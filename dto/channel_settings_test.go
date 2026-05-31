package dto

import "testing"

func TestApplyDerivedProxyFromOpenVPNConfig(t *testing.T) {
	settings := ChannelSettings{
		OpenVPNConfig: `client
remote vpn.example.com 1194
http-proxy 127.0.0.1 8080
`,
	}

	settings.ApplyDerivedProxy()

	if settings.Proxy != "http://127.0.0.1:8080" {
		t.Fatalf("expected derived http proxy, got %q", settings.Proxy)
	}
}

func TestApplyDerivedProxyPreservesExplicitProxy(t *testing.T) {
	settings := ChannelSettings{
		Proxy: "socks5://127.0.0.1:1080",
		OpenVPNConfig: `socks-proxy 10.0.0.1 1081`,
	}

	settings.ApplyDerivedProxy()

	if settings.Proxy != "socks5://127.0.0.1:1080" {
		t.Fatalf("expected explicit proxy to win, got %q", settings.Proxy)
	}
}

func TestApplyDerivedProxyFromSocksProxyDirective(t *testing.T) {
	settings := ChannelSettings{
		OpenVPNConfig: `
; comment
socks-proxy 127.0.0.1 1080
`,
	}

	settings.ApplyDerivedProxy()

	if settings.Proxy != "socks5://127.0.0.1:1080" {
		t.Fatalf("expected derived socks proxy, got %q", settings.Proxy)
	}
}
