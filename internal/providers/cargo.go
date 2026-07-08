package providers

import (
	"context"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/GinForGit/cli-migration/pkg/api"
)

// CargoProvider detects and installs Rust binaries via cargo.
type CargoProvider struct{ BaseProvider }

// NewCargoProvider creates a new cargo provider.
func NewCargoProvider() Provider {
	return &CargoProvider{BaseProvider{name: api.ProviderCargo}}
}

func (c *CargoProvider) Available() bool {
	_, err := exec.LookPath("cargo")
	return err == nil
}

func (c *CargoProvider) Detect(ctx context.Context) ([]api.Entry, error) {
	cmd := exec.CommandContext(ctx, "cargo", "install", "--list", "--format", "json")
	out, err := cmd.Output()
	if err != nil {
		// Older cargo versions do not support --format json.
		return c.detectLegacy(ctx)
	}
	return parseCargoJSON(out)
}

func parseCargoJSON(data []byte) ([]api.Entry, error) {
	var crates []struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Binaries []string `json:"binaries"`
	}
	if err := json.Unmarshal(data, &crates); err != nil {
		return nil, err
	}
	var entries []api.Entry
	for _, c := range crates {
		cmd := c.Name
		if len(c.Binaries) > 0 {
			cmd = filepath.Base(c.Binaries[0])
		}
		entries = append(entries, api.Entry{
			Name:    cmd,
			Command: cmd,
			Version: c.Version,
			Provider: api.ProviderCargo,
			ProviderArgs: map[string]interface{}{
				"package": c.Name,
			},
		})
	}
	return entries, nil
}

func (c *CargoProvider) detectLegacy(ctx context.Context) ([]api.Entry, error) {
	cmd := exec.CommandContext(ctx, "cargo", "install", "--list")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var entries []api.Entry
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "(") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		name := parts[0]
		version := strings.Trim(parts[1], "():v")
		entries = append(entries, api.Entry{
			Name:    name,
			Command: name,
			Version: version,
			Provider: api.ProviderCargo,
			ProviderArgs: map[string]interface{}{
				"package": name,
			},
		})
	}
	return entries, nil
}

func (c *CargoProvider) Resolve(entry api.Entry) (InstallPlan, error) {
	return ResolveDefault(entry)
}

func (c *CargoProvider) Install(ctx context.Context, plan InstallPlan) error {
	args := []string{"cargo", "install", plan.Package}
	if plan.Version != "" && plan.Version != "unknown" {
		args = append(args, "--version", plan.Version)
	}
	return runCommand(ctx, args)
}

func (c *CargoProvider) Uninstall(ctx context.Context, plan InstallPlan) error {
	return runCommand(ctx, []string{"cargo", "uninstall", plan.Package})
}
