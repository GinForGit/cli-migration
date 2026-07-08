package bundle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GinForGit/cli-migration/internal/manifest"
	"github.com/GinForGit/cli-migration/pkg/api"
)

func TestPackAndUnpack(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.yaml")
	bundlePath := filepath.Join(dir, "test.bundle.tar.gz")
	unpackDir := filepath.Join(dir, "unpacked")

	source := api.SourceInfo{OS: api.OSLinux, Arch: "amd64"}
	m := manifest.New(source)
	m.Entries = []api.Entry{
		{Name: "git", Command: "git", Version: "2.45.0", Provider: api.ProviderApt},
	}
	if err := manifest.Save(manifestPath, m); err != nil {
		t.Fatalf("save manifest failed: %v", err)
	}

	if err := Pack(manifestPath, bundlePath, PackOptions{}); err != nil {
		t.Fatalf("pack failed: %v", err)
	}

	info, err := os.Stat(bundlePath)
	if err != nil {
		t.Fatalf("stat bundle failed: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("expected non-empty bundle")
	}

	if err := os.MkdirAll(unpackDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	extractedManifest, err := Unpack(bundlePath, unpackDir)
	if err != nil {
		t.Fatalf("unpack failed: %v", err)
	}

	loaded, err := manifest.Load(extractedManifest)
	if err != nil {
		t.Fatalf("load extracted manifest failed: %v", err)
	}
	if len(loaded.Entries) != 1 || loaded.Entries[0].Command != "git" {
		t.Fatalf("unexpected extracted manifest: %+v", loaded.Entries)
	}
}

func TestPackWithConfigs(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.yaml")
	bundlePath := filepath.Join(dir, "test.bundle.tar.gz")
	configPath := filepath.Join(dir, "gitconfig")

	if err := os.WriteFile(configPath, []byte("[user]\nname = Test\n"), 0o644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	source := api.SourceInfo{OS: api.OSWindows, Arch: "amd64"}
	m := manifest.New(source)
	if err := manifest.Save(manifestPath, m); err != nil {
		t.Fatalf("save manifest failed: %v", err)
	}

	if err := Pack(manifestPath, bundlePath, PackOptions{
		IncludeConfigs: true,
		ConfigPaths:    []string{configPath},
	}); err != nil {
		t.Fatalf("pack with configs failed: %v", err)
	}
}
