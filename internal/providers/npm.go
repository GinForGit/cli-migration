package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/GinForGit/cli-migration/pkg/api"
)

// NpmProvider detects globally installed npm packages.
type NpmProvider struct{ BaseProvider }

// NewNpmProvider creates a new npm provider.
func NewNpmProvider() Provider {
	return &NpmProvider{BaseProvider{name: api.ProviderNpm}}
}

func (n *NpmProvider) Available() bool {
	_, err := exec.LookPath("npm")
	return err == nil
}

func (n *NpmProvider) Detect(ctx context.Context) ([]api.Entry, error) {
	cmd := exec.CommandContext(ctx, "npm", "list", "-g", "--depth=0", "--json")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("npm list failed: %w", err)
	}
	return parseNpmList(out)
}

func parseNpmList(data []byte) ([]api.Entry, error) {
	type npmDep struct {
		Version   string            `json:"version"`
		From      string            `json:"from"`
		Resolved  string            `json:"resolved"`
		Dependencies map[string]*npmDep `json:"dependencies"`
	}
	type npmList struct {
		Name         string            `json:"name"`
		Dependencies map[string]*npmDep `json:"dependencies"`
	}

	var list npmList
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, err
	}

	var entries []api.Entry
	for name, dep := range list.Dependencies {
		if name == "" {
			continue
		}
		entries = append(entries, api.Entry{
			Name:    name,
			Command: name,
			Version: dep.Version,
			Provider: api.ProviderNpm,
			ProviderArgs: map[string]interface{}{
				"package": name,
				"global":  true,
			},
		})
	}
	return entries, nil
}

func (n *NpmProvider) Resolve(entry api.Entry) (InstallPlan, error) {
	pkg, _ := entry.ProviderArgs["package"].(string)
	if pkg == "" {
		return InstallPlan{}, fmt.Errorf("provider_args.package is required")
	}
	return InstallPlan{
		Provider: api.ProviderNpm,
		Package:  pkg,
		Version:  entry.Version,
		Args:     entry.ProviderArgs,
	}, nil
}

func (n *NpmProvider) Install(ctx context.Context, plan InstallPlan) error {
	args := []string{"npm", "install", "-g", plan.Package}
	if plan.Version != "" && plan.Version != "unknown" {
		args = append(args, "@"+plan.Version)
	}
	return runCommand(ctx, args)
}

func (n *NpmProvider) Uninstall(ctx context.Context, plan InstallPlan) error {
	return runCommand(ctx, []string{"npm", "uninstall", "-g", plan.Package})
}
