package service

import "testing"

func TestIsAutoRoutingModel(t *testing.T) {
	cases := map[string]bool{
		"auto":          true,
		"auto-deepseek": true,
		"auto-kimi":     true,
		"gpt-4o":        false,
	}

	for modelName, expected := range cases {
		if got := IsAutoRoutingModel(modelName); got != expected {
			t.Fatalf("IsAutoRoutingModel(%q) = %v, want %v", modelName, got, expected)
		}
	}
}

func TestMatchesAutoRoute(t *testing.T) {
	cases := []struct {
		candidate  string
		routeModel string
		want       bool
	}{
		{candidate: "deepseek-chat", routeModel: "auto-deepseek", want: true},
		{candidate: "kimi-k2", routeModel: "auto-kimi", want: true},
		{candidate: "ossgpt-32b", routeModel: "auto-ossgpt", want: true},
		{candidate: "gpt-4o", routeModel: "auto-deepseek", want: false},
		{candidate: "gpt-4o", routeModel: "auto", want: true},
	}

	for _, tc := range cases {
		if got := matchesAutoRoute(tc.candidate, tc.routeModel); got != tc.want {
			t.Fatalf("matchesAutoRoute(%q, %q) = %v, want %v", tc.candidate, tc.routeModel, got, tc.want)
		}
	}
}
