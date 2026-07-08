package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/GinForGit/cli-migration/pkg/api"
)

// WingetProvider detects and installs packages via winget.
type WingetProvider struct{ BaseProvider }

// NewWingetProvider creates a new winget provider.
func NewWingetProvider() Provider {
	return &WingetProvider{BaseProvider{name: api.ProviderWinget}}
}

func (w *WingetProvider) Available() bool {
	if runtime.GOOS != "windows" {
		return false
	}
	_, err := exec.LookPath("winget")
	return err == nil
}

func (w *WingetProvider) Detect(ctx context.Context) ([]api.Entry, error) {
	cmd := exec.CommandContext(ctx, "winget", "export", "--output", "-", "--source", "winget")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("winget export failed: %w", err)
	}
	return parseWingetExport(out)
}

func parseWingetExport(data []byte) ([]api.Entry, error) {
	// winget export outputs a JSON document with Sources/Entries.
	type wingetEntry struct {
		PackageIdentifier string `json:"PackageIdentifier"`
		Version           string `json:"Version"`
		SourceDetails     struct {
			Name string `json:"Name"`
		} `json:"SourceDetails"`
	}
	type wingetExport struct {
		Sources []struct {
			SourceDetails struct {
				Name      string `json:"Name"`
				Argument  string `json:"Argument"`
				Identifier string `json:"Identifier"`
			} `json:"SourceDetails"`
			Packages []wingetEntry `json:"Packages"`
		} `json:"Sources"`
	}

	var export wingetExport
	if err := json.Unmarshal(data, &export); err != nil {
		return nil, err
	}

	var entries []api.Entry
	for _, source := range export.Sources {
		for _, p := range source.Packages {
			entries = append(entries, api.Entry{
				Name:    packageName(p.PackageIdentifier),
				Command: packageName(p.PackageIdentifier),
				Version: p.Version,
				Provider: api.ProviderWinget,
				ProviderArgs: map[string]interface{}{
					"package_id": p.PackageIdentifier,
					"source":     source.SourceDetails.Name,
				},
			})
		}
	}
	return entries, nil
}

func packageName(id string) string {
	for i := len(id) - 1; i >= 0; i-- {
		if id[i] == '.' {
			return id[i+1:]
		}
	}
	return id
}

func (w *WingetProvider) Resolve(entry api.Entry) (InstallPlan, error) {
	id, _ := entry.ProviderArgs["package_id"].(string)
	if id == "" {
		return InstallPlan{}, fmt.Errorf("provider_args.package_id is required")
	}
	return InstallPlan{
		Provider: api.ProviderWinget,
		Package:  id,
		Version:  entry.Version,
		Args:     entry.ProviderArgs,
	}, nil
}

func (w *WingetProvider) Install(ctx context.Context, plan InstallPlan) error {
	args := []string{"winget", "install", "--id", plan.Package, "--source", "winget", "--accept-package-agreements", "--accept-source-agreements"}
	if plan.Version != "" && plan.Version != "unknown" {
		args = append(args, "--version", plan.Version)
	}
	return runCommand(ctx, args)
}

func (w *WingetProvider) Uninstall(ctx context.Context, plan InstallPlan) error {
	return runCommand(ctx, []string{"winget", "uninstall", "--id", plan.Package})
}
