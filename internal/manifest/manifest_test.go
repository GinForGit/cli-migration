package manifest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GinForGit/cli-migration/pkg/api"
)

func TestNew(t *testing.T) {
	source := api.SourceInfo{OS: api.OSWindows, Arch: "amd64"}
	m := New(source)
	if m.Version != CurrentVersion {
		t.Fatalf("expected version %s, got %s", CurrentVersion, m.Version)
	}
	if m.Source.Arch != "amd64" {
		t.Fatalf("expected arch amd64, got %s", m.Source.Arch)
	}
	if m.GeneratedAt.IsZero() {
		t.Fatal("expected GeneratedAt to be set")
	}
}

func TestValidate(t *testing.T) {
	source := api.SourceInfo{OS: api.OSWindows}
	m := New(source)
	if err := Validate(m); err != nil {
		t.Fatalf("valid manifest failed validation: %v", err)
	}

	invalid := &api.Manifest{Version: CurrentVersion}
	if err := Validate(invalid); err == nil {
		t.Fatal("expected validation error for missing source")
	}
}

func TestRoundTripYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.yaml")

	source := api.SourceInfo{OS: api.OSLinux, Arch: "amd64"}
	m := New(source)
	m.Entries = []api.Entry{
		{Name: "git", Command: "git", Version: "2.45.0", Provider: api.ProviderApt},
	}

	if err := Save(path, m); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(loaded.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(loaded.Entries))
	}
	if loaded.Entries[0].Command != "git" {
		t.Fatalf("expected command git, got %s", loaded.Entries[0].Command)
	}
}

func TestRoundTripJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")

	source := api.SourceInfo{OS: api.OSWindows, Arch: "amd64"}
	m := New(source)
	m.Entries = []api.Entry{
		{Name: "node", Command: "node", Version: "20.0.0", Provider: api.ProviderScoop},
	}

	if err := SaveJSON(path, m); err != nil {
		t.Fatalf("save json failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty json file")
	}
}
