package providers

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/GinForGit/cli-migration/pkg/api"
)

// ScoopProvider detects and installs packages via Scoop.
type ScoopProvider struct{ BaseProvider }

// NewScoopProvider creates a new Scoop provider.
func NewScoopProvider() Provider {
	return &ScoopProvider{BaseProvider{name: api.ProviderScoop}}
}

func (s *ScoopProvider) Available() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	_, err := os.Stat(filepath.Join(scoopRoot(), "apps"))
	return err == nil
}

func (s *ScoopProvider) Detect(ctx context.Context) ([]api.Entry, error) {
	root := scoopRoot()
	appsDir := filepath.Join(root, "apps")
	entries, err := os.ReadDir(appsDir)
	if err != nil {
		return nil, err
	}

	var result []api.Entry
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		entry, ok := s.detectApp(appsDir, e.Name())
		if ok {
			result = append(result, entry)
		}
	}
	return result, nil
}

func (s *ScoopProvider) detectApp(appsDir, app string) (api.Entry, bool) {
	currentDir := filepath.Join(appsDir, app, "current")
	info, err := os.Lstat(currentDir)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		// fallback: list version directories and pick latest
		return s.detectAppFallback(appsDir, app)
	}

	// Try to read install.json for metadata.
	installJSON := filepath.Join(currentDir, "install.json")
	version := "unknown"
	if data, err := os.ReadFile(installJSON); err == nil {
		version = extractVersionFromJSON(string(data))
	}

	return api.Entry{
		Name:    app,
		Command: app,
		Version: version,
		Provider: api.ProviderScoop,
		ProviderArgs: map[string]interface{}{
			"bucket":  inferBucket(appsDir, app),
			"package": app,
		},
	}, true
}

func (s *ScoopProvider) detectAppFallback(appsDir, app string) (api.Entry, bool) {
	versionsDir := filepath.Join(appsDir, app)
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		return api.Entry{}, false
	}
	for _, e := range entries {
		if e.IsDir() {
			return api.Entry{
				Name:    app,
				Command: app,
				Version: e.Name(),
				Provider: api.ProviderScoop,
				ProviderArgs: map[string]interface{}{
					"bucket":  inferBucket(appsDir, app),
					"package": app,
				},
			}, true
		}
	}
	return api.Entry{}, false
}

func (s *ScoopProvider) Resolve(entry api.Entry) (InstallPlan, error) {
	return ResolveDefault(entry)
}

func (s *ScoopProvider) Install(ctx context.Context, plan InstallPlan) error {
	bucket, _ := plan.Args["bucket"].(string)
	pkg := plan.Package
	if bucket != "" && bucket != "main" {
		pkg = bucket + "/" + pkg
	}
	cmd := []string{"scoop", "install", pkg}
	if plan.Version != "" && plan.Version != "unknown" {
		// Scoop supports version pinning via bucket/package@version in some contexts.
		// We keep it simple here.
	}
	return runCommand(ctx, cmd)
}

func (s *ScoopProvider) Uninstall(ctx context.Context, plan InstallPlan) error {
	return runCommand(ctx, []string{"scoop", "uninstall", plan.Package})
}

func scoopRoot() string {
	if root := os.Getenv("SCOOP"); root != "" {
		return root
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "scoop")
}

func inferBucket(appsDir, app string) string {
	// Best-effort: inspect the parent of the current symlink if it points to buckets.
	// For Phase 0 we default to "main".
	_ = appsDir
	_ = app
	return "main"
}

func extractVersionFromJSON(data string) string {
	// Very naive extraction for Phase 0.
	const key = `"version"`
	idx := strings.Index(data, key)
	if idx == -1 {
		return "unknown"
	}
	start := strings.Index(data[idx:], `"`) + idx + 1
	start = strings.Index(data[start:], `"`) + start + 1
	end := strings.Index(data[start:], `"`)
	if end == -1 {
		return "unknown"
	}
	return data[start : start+end]
}
