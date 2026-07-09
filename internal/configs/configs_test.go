package configs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GinForGit/cli-migration/internal/platform"
	"github.com/GinForGit/cli-migration/pkg/api"
)

type fakePlat struct {
	home   string
	shells []string
}

func (f fakePlat) OS() platform.OSType                { return platform.OSLinux }
func (f fakePlat) Arch() string                       { return "amd64" }
func (f fakePlat) HomeDir() string                    { return f.home }
func (f fakePlat) ExecutableExtensions() []string     { return nil }
func (f fakePlat) ListPathEntries() []string          { return nil }
func (f fakePlat) ShellConfigFiles() []string         { return f.shells }
func (f fakePlat) IsElevated() bool                   { return false }
func (f fakePlat) QuoteCommand(args []string) string  { return "" }
func (f fakePlat) JoinPath(elem ...string) string     { return filepath.Join(elem...) }
func (f fakePlat) Abs(path string) (string, error)    { return filepath.Abs(path) }

func TestCollectAlias(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".bashrc")
	if err := os.WriteFile(rc, []byte("alias gs='git status'\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	plat := fakePlat{home: dir, shells: []string{rc}}
	entries := []api.Entry{{Name: "git", Command: "git", Provider: api.ProviderApt}}

	out, err := Collect(plat, entries)
	if err != nil {
		t.Fatal(err)
	}
	if len(out[0].ConfigRefs) != 1 {
		t.Fatalf("expected 1 config ref, got %d", len(out[0].ConfigRefs))
	}
	if out[0].ConfigRefs[0].Type != "alias" {
		t.Fatalf("expected alias, got %s", out[0].ConfigRefs[0].Type)
	}
}

func TestCollectEnv(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".bashrc")
	if err := os.WriteFile(rc, []byte("export GIT_EDITOR=vim\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	plat := fakePlat{home: dir, shells: []string{rc}}
	entries := []api.Entry{
		{Name: "git", Command: "git", Provider: api.ProviderApt},
		{Name: "vim", Command: "vim", Provider: api.ProviderApt},
	}

	out, err := Collect(plat, entries)
	if err != nil {
		t.Fatal(err)
	}
	if len(out[0].ConfigRefs) != 1 {
		t.Fatalf("expected git to get env ref, got %d", len(out[0].ConfigRefs))
	}
	if out[0].ConfigRefs[0].Key != "GIT_EDITOR" {
		t.Fatalf("expected GIT_EDITOR, got %s", out[0].ConfigRefs[0].Key)
	}
}

func TestApplyShellRefs(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".bashrc")
	plat := fakePlat{home: dir, shells: []string{rc}}
	entry := api.Entry{
		Name:    "git",
		Command: "git",
		ConfigRefs: []api.ConfigRef{
			{Type: "alias", Key: "gs", Value: "git status"},
			{Type: "env", Key: "GIT_EDITOR", Value: "vim"},
		},
	}

	if err := Apply(plat, entry); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(rc)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !contains(content, "alias gs='git status'") {
		t.Fatalf("expected alias in shell config, got:\n%s", content)
	}
	if !contains(content, "export GIT_EDITOR=vim") {
		t.Fatalf("expected env in shell config, got:\n%s", content)
	}
}

func TestApplyFile(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	src := filepath.Join(srcDir, ".gitconfig")
	if err := os.WriteFile(src, []byte("[user]\n\tname = Test\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	plat := fakePlat{home: dstDir}
	entry := api.Entry{
		Name:    "git",
		Command: "git",
		ConfigRefs: []api.ConfigRef{
			{Type: "file", Source: src, Target: ".gitconfig"},
		},
	}

	if err := Apply(plat, entry); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dstDir, ".gitconfig"))
	if err != nil {
		t.Fatal(err)
	}
	if !contains(string(data), "[user]") {
		t.Fatalf("expected copied gitconfig content, got:\n%s", string(data))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
