package providers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/GinForGit/cli-migration/pkg/api"
)

// ChocolateyProvider detects and installs packages via Chocolatey.
type ChocolateyProvider struct{ BaseProvider }

// NewChocolateyProvider creates a new Chocolatey provider.
func NewChocolateyProvider() Provider {
	return &ChocolateyProvider{BaseProvider{name: api.ProviderChoco}}
}

func (c *ChocolateyProvider) Available() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	_, err := exec.LookPath("choco")
	return err == nil
}

func (c *ChocolateyProvider) Detect(ctx context.Context) ([]api.Entry, error) {
	cmd := exec.CommandContext(ctx, "choco", "list", "--local-only", "--limit-output")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("choco list failed: %w", err)
	}
	return parseChocoList(string(out)), nil
}

func parseChocoList(output string) []api.Entry {
	var entries []api.Entry
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		version := strings.TrimSpace(parts[1])
		entries = append(entries, api.Entry{
			Name:    name,
			Command: name,
			Version: version,
			Provider: api.ProviderChoco,
			ProviderArgs: map[string]interface{}{
				"package": name,
			},
		})
	}
	return entries
}

func (c *ChocolateyProvider) Resolve(entry api.Entry) (InstallPlan, error) {
	return ResolveDefault(entry)
}

func (c *ChocolateyProvider) Install(ctx context.Context, plan InstallPlan) error {
	args := []string{"choco", "install", plan.Package, "--yes"}
	if plan.Version != "" && plan.Version != "unknown" {
		args = append(args, "--version", plan.Version)
	}
	return runCommand(ctx, args)
}

func (c *ChocolateyProvider) Uninstall(ctx context.Context, plan InstallPlan) error {
	return runCommand(ctx, []string{"choco", "uninstall", plan.Package, "--yes"})
}

// chocoLibDir returns the Chocolatey lib directory.
func chocoLibDir() string {
	if root := os.Getenv("ChocolateyInstall"); root != "" {
		return filepath.Join(root, "lib")
	}
	return `C:\ProgramData\chocolatey\lib`
}
