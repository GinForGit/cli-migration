package crossplatform

import (
	"testing"

	"github.com/GinForGit/cli-migration/pkg/api"
)

func TestCanResolve(t *testing.T) {
	entry := api.Entry{
		Name:    "git",
		Command: "git",
		Provider: api.ProviderScoop,
		ProviderArgs: map[string]interface{}{
			"package": "git",
		},
	}
	if !CanResolve(entry, api.OSWindows, api.OSLinux) {
		t.Fatal("expected git to be resolvable from windows to linux")
	}
	if CanResolve(entry, api.OSWindows, api.OSWindows) {
		t.Fatal("expected same-os resolution to be false")
	}
}

func TestResolve(t *testing.T) {
	entry := api.Entry{
		Name:    "git",
		Command: "git",
		Version: "2.45.0",
		Provider: api.ProviderScoop,
		ProviderArgs: map[string]interface{}{
			"package": "git",
		},
	}
	resolver := NewDefaultResolver()
	resolved, err := resolver.Resolve(entry, api.OSWindows, api.OSLinux)
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if resolved.Provider != api.ProviderApt {
		t.Fatalf("expected apt, got %s", resolved.Provider)
	}
	pkg, _ := resolved.ProviderArgs["package"].(string)
	if pkg != "git" {
		t.Fatalf("expected package git, got %s", pkg)
	}
	if _, ok := resolved.TargetOverrides[api.OSLinux]; !ok {
		t.Fatal("expected target override for linux")
	}
}

func TestResolveNoMapping(t *testing.T) {
	entry := api.Entry{
		Name:    "unknown-tool",
		Command: "unknown-tool",
		Provider: api.ProviderScoop,
		ProviderArgs: map[string]interface{}{
			"package": "unknown-tool",
		},
	}
	resolver := NewDefaultResolver()
	_, err := resolver.Resolve(entry, api.OSWindows, api.OSLinux)
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
}
