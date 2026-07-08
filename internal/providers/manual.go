package providers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/GinForGit/cli-migration/pkg/api"
)

// ManualProvider handles CLI tools that were installed by hand and cannot be
// automatically managed by a package manager.
type ManualProvider struct{ BaseProvider }

// NewManualProvider creates the manual provider.
func NewManualProvider() Provider {
	return &ManualProvider{BaseProvider{name: api.ProviderManual}}
}

func (m *ManualProvider) Available() bool {
	return true
}

func (m *ManualProvider) Detect(ctx context.Context) ([]api.Entry, error) {
	// Manual provider does not discover on its own. Discovery engine uses it as
	// a fallback for executables not claimed by other providers.
	return []api.Entry{}, nil
}

func (m *ManualProvider) Resolve(entry api.Entry) (InstallPlan, error) {
	return InstallPlan{}, fmt.Errorf("manual provider cannot resolve automatic installation")
}

func (m *ManualProvider) Install(ctx context.Context, plan InstallPlan) error {
	return fmt.Errorf("manual provider cannot install automatically")
}

func (m *ManualProvider) Uninstall(ctx context.Context, plan InstallPlan) error {
	path, _ := plan.Args["path"].(string)
	if path == "" {
		return fmt.Errorf("manual uninstall requires path")
	}
	return os.Remove(path)
}

// NewManualEntry creates a fallback entry for an executable found on PATH.
func NewManualEntry(path string) api.Entry {
	name := filepath.Base(path)
	if runtime.GOOS == "windows" {
		ext := filepath.Ext(name)
		name = name[:len(name)-len(ext)]
	}
	return api.Entry{
		Name:    name,
		Command: name,
		Version: "unknown",
		Provider: api.ProviderManual,
		ProviderArgs: map[string]interface{}{
			"path": path,
		},
	}
}
