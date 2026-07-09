package discover

import (
	"testing"

	"github.com/GinForGit/cli-migration/pkg/api"
)

func TestFilterEntries(t *testing.T) {
	entries := []api.Entry{
		{Name: "Git", Command: "git", Provider: api.ProviderApt},
		{Name: "Node.js", Command: "node", Provider: api.ProviderApt},
		{Name: "CLI Tool", Command: "cli-tool", Provider: api.ProviderManual},
		{Name: "Azure CLI", Command: "az", Provider: api.ProviderManual},
	}

	filtered := FilterEntries(entries, "cli")
	if len(filtered) != 2 {
		t.Fatalf("expected 2 entries matching 'cli', got %d", len(filtered))
	}
	if filtered[0].Command != "cli-tool" {
		t.Errorf("expected cli-tool, got %s", filtered[0].Command)
	}
	if filtered[1].Command != "az" {
		t.Errorf("expected az, got %s", filtered[1].Command)
	}
}

func TestFilterEntriesEmpty(t *testing.T) {
	entries := []api.Entry{
		{Name: "Git", Command: "git", Provider: api.ProviderApt},
	}

	filtered := FilterEntries(entries, "")
	if len(filtered) != len(entries) {
		t.Fatalf("expected all entries when filter is empty, got %d", len(filtered))
	}
}

func TestFilterEntriesCaseInsensitive(t *testing.T) {
	entries := []api.Entry{
		{Name: "Azure CLI", Command: "az", Provider: api.ProviderManual},
	}

	filtered := FilterEntries(entries, "CLI")
	if len(filtered) != 1 {
		t.Fatalf("expected 1 entry matching case-insensitive 'CLI', got %d", len(filtered))
	}
}

func TestFilterEntriesNoMatch(t *testing.T) {
	entries := []api.Entry{
		{Name: "Git", Command: "git", Provider: api.ProviderApt},
	}

	filtered := FilterEntries(entries, "xyz")
	if len(filtered) != 0 {
		t.Fatalf("expected 0 entries matching 'xyz', got %d", len(filtered))
	}
}
