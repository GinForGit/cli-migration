package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/GinForGit/cli-migration/pkg/api"
)

// PipxProvider detects Python CLI tools installed via pipx.
type PipxProvider struct{ BaseProvider }

// NewPipxProvider creates a new pipx provider.
func NewPipxProvider() Provider {
	return &PipxProvider{BaseProvider{name: api.ProviderPipx}}
}

func (p *PipxProvider) Available() bool {
	_, err := exec.LookPath("pipx")
	return err == nil
}

func (p *PipxProvider) Detect(ctx context.Context) ([]api.Entry, error) {
	cmd := exec.CommandContext(ctx, "pipx", "list", "--json")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("pipx list failed: %w", err)
	}
	return parsePipxList(out)
}

func parsePipxList(data []byte) ([]api.Entry, error) {
	var result struct {
		Venvs map[string]struct {
			Package   string `json:"package"`
			PackageOrURL string `json:"package_or_url"`
			Version   string `json:"version"`
			Suffix    string `json:"suffix"`
		} `json:"venvs"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	var entries []api.Entry
	for name, v := range result.Venvs {
		pkg := v.Package
		if pkg == "" {
			pkg = v.PackageOrURL
		}
		entries = append(entries, api.Entry{
			Name:    name,
			Command: name,
			Version: v.Version,
			Provider: api.ProviderPipx,
			ProviderArgs: map[string]interface{}{
				"package": pkg,
			},
		})
	}
	return entries, nil
}

func (p *PipxProvider) Resolve(entry api.Entry) (InstallPlan, error) {
	return ResolveDefault(entry)
}

func (p *PipxProvider) Install(ctx context.Context, plan InstallPlan) error {
	args := []string{"pipx", "install", plan.Package}
	if plan.Version != "" && plan.Version != "unknown" {
		args = append(args, "=="+plan.Version)
	}
	return runCommand(ctx, args)
}

func (p *PipxProvider) Uninstall(ctx context.Context, plan InstallPlan) error {
	return runCommand(ctx, []string{"pipx", "uninstall", plan.Package})
}

// normalizePipxName strips extras and version specifiers.
func normalizePipxName(s string) string {
	idx := strings.IndexAny(s, "[>=<~!")
	if idx == -1 {
		return s
	}
	return strings.TrimSpace(s[:idx])
}
