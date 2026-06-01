package controller

import (
	"reflect"
	"testing"
)

func TestBuildAutoRoutingAliases(t *testing.T) {
	aliases := buildAutoRoutingAliases([]string{
		"deepseek-chat",
		"deepseek-reasoner",
		"kimi-k2",
		"ossgpt-32b",
		"gpt-4o",
	})

	want := []string{"auto", "auto-deepseek", "auto-kimi", "auto-ossgpt", "auto-gpt"}
	if !reflect.DeepEqual(aliases, want) {
		t.Fatalf("buildAutoRoutingAliases() = %#v, want %#v", aliases, want)
	}
}
