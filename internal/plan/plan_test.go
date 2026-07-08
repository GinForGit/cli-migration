package plan

import (
	"context"
	"testing"

	"github.com/GinForGit/cli-migration/internal/platform"
	"github.com/GinForGit/cli-migration/pkg/api"
)

type fakePlatform struct{}

func (fakePlatform) OS() platform.OSType   { return platform.OSLinux }
func (fakePlatform) Arch() string          { return "amd64" }
func (fakePlatform) HomeDir() string       { return "/home/test" }
func (fakePlatform) ExecutableExtensions() []string { return []string{""} }
func (fakePlatform) ListPathEntries() []string      { return nil }
func (fakePlatform) ShellConfigFiles() []string     { return nil }
func (fakePlatform) IsElevated() bool               { return false }
func (fakePlatform) QuoteCommand(args []string) string { return "" }
func (fakePlatform) JoinPath(elem ...string) string  { return "" }
func (fakePlatform) Abs(path string) (string, error) { return path, nil }

func TestResolveActionInstall(t *testing.T) {
	entry := api.Entry{Name: "node", Command: "node", Version: "20.0.0", Provider: api.ProviderApt}
	current := map[string]api.Entry{}
	available := map[api.ProviderName]bool{api.ProviderApt: true}

	action := resolveAction(entry, current, available, fakePlatform{})
	if action.Kind != api.ActionInstall {
		t.Fatalf("expected install, got %s", action.Kind)
	}
}

func TestResolveActionSkip(t *testing.T) {
	entry := api.Entry{Name: "git", Command: "git", Version: "2.45.0", Provider: api.ProviderApt}
	current := map[string]api.Entry{
		"git": {Name: "git", Command: "git", Version: "2.45.0", Provider: api.ProviderApt},
	}
	available := map[api.ProviderName]bool{api.ProviderApt: true}

	action := resolveAction(entry, current, available, fakePlatform{})
	if action.Kind != api.ActionSkip {
		t.Fatalf("expected skip, got %s", action.Kind)
	}
}

func TestResolveActionUpgrade(t *testing.T) {
	entry := api.Entry{Name: "git", Command: "git", Version: "2.45.0", Provider: api.ProviderApt}
	current := map[string]api.Entry{
		"git": {Name: "git", Command: "git", Version: "2.40.0", Provider: api.ProviderApt},
	}
	available := map[api.ProviderName]bool{api.ProviderApt: true}

	action := resolveAction(entry, current, available, fakePlatform{})
	if action.Kind != api.ActionUpgrade {
		t.Fatalf("expected upgrade, got %s", action.Kind)
	}
}

func TestCompareVersions(t *testing.T) {
	cases := []struct{ a, b string; want int }{
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"1.0.0", "1.0.0", 0},
		{"v1.2.0", "1.3.0", -1},
	}
	for _, c := range cases {
		got := compareVersions(c.a, c.b)
		if got != c.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestGenerate(t *testing.T) {
	m := &api.Manifest{
		Version: "1.0",
		Source:  api.SourceInfo{OS: api.OSLinux},
		Entries: []api.Entry{
			{Name: "git", Command: "git", Version: "2.45.0", Provider: api.ProviderApt},
		},
	}
	p, err := Generate(context.Background(), fakePlatform{}, m, api.OSLinux)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	if len(p.Actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(p.Actions))
	}
}
