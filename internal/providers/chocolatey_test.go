package providers

import (
	"testing"

	"github.com/GinForGit/cli-migration/pkg/api"
)

func TestParseChocoList(t *testing.T) {
	input := `chocolatey|2.2.2
git|2.45.0
nodejs|20.12.0
`
	entries := parseChocoList(input)
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].Name != "chocolatey" || entries[0].Version != "2.2.2" {
		t.Fatalf("unexpected first entry: %+v", entries[0])
	}
	if entries[1].Provider != api.ProviderChoco {
		t.Fatalf("expected provider choco, got %s", entries[1].Provider)
	}
}

func TestParseChocoListEmpty(t *testing.T) {
	entries := parseChocoList("")
	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}
