package providers

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/GinForGit/cli-migration/pkg/api"
)

// AptProvider detects and installs packages via apt/dpkg.
type AptProvider struct{ BaseProvider }

// NewAptProvider creates a new Apt provider.
func NewAptProvider() Provider {
	return &AptProvider{BaseProvider{name: api.ProviderApt}}
}

func (a *AptProvider) Available() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	_, err := exec.LookPath("dpkg")
	return err == nil
}

func (a *AptProvider) Detect(ctx context.Context) ([]api.Entry, error) {
	cmd := exec.CommandContext(ctx, "dpkg-query", "-W", "-f=${Package}\t${Version}\n")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("dpkg-query failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var result []api.Entry
	for _, line := range lines {
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		version := strings.TrimSpace(parts[1])
		// Heuristic: many CLI packages are exactly the command name.
		// We do not try to map binary names here; PATH scanning in discovery
		// will correlate these later.
		result = append(result, api.Entry{
			Name:    name,
			Command: name,
			Version: version,
			Provider: api.ProviderApt,
			ProviderArgs: map[string]interface{}{
				"package": name,
			},
		})
	}
	return result, nil
}

func (a *AptProvider) Resolve(entry api.Entry) (InstallPlan, error) {
	return ResolveDefault(entry)
}

func (a *AptProvider) Install(ctx context.Context, plan InstallPlan) error {
	pkg := plan.Package
	if plan.Version != "" && plan.Version != "unknown" {
		pkg = fmt.Sprintf("%s=%s", pkg, plan.Version)
	}
	return runCommand(ctx, []string{"apt-get", "install", "-y", pkg})
}

func (a *AptProvider) Uninstall(ctx context.Context, plan InstallPlan) error {
	return runCommand(ctx, []string{"apt-get", "remove", "-y", plan.Package})
}
