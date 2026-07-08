// Package apply executes a plan on the current machine.
package apply

import (
	"context"
	"fmt"

	"github.com/GinForGit/cli-migration/internal/platform"
	"github.com/GinForGit/cli-migration/internal/providers"
	"github.com/GinForGit/cli-migration/pkg/api"
)

// Options controls apply behavior.
type Options struct {
	SkipManual bool
}

// Execute runs the actions in a plan.
func Execute(ctx context.Context, plat platform.Platform, p *api.Plan, opts Options) error {
	registry := providers.NewRegistry()
	var installed, upgraded, skipped, failed int

	for _, action := range p.Actions {
		switch action.Kind {
		case api.ActionInstall, api.ActionUpgrade, api.ActionDowngrade:
			if opts.SkipManual && action.Entry.Provider == api.ProviderManual {
				fmt.Printf("[skip-manual] %s\n", action.Entry.Command)
				skipped++
				continue
			}
			if err := install(ctx, registry, action.Entry); err != nil {
				fmt.Printf("[error] %s: %v\n", action.Entry.Command, err)
				failed++
				continue
			}
			if action.Kind == api.ActionInstall {
				installed++
			} else {
				upgraded++
			}
		case api.ActionSkip:
			fmt.Printf("[skip] %s\n", action.Entry.Command)
			skipped++
		case api.ActionUnavailable:
			fmt.Printf("[unavailable] %s: %s\n", action.Entry.Command, action.Message)
			skipped++
		case api.ActionWarn:
			fmt.Printf("[warn] %s: %s\n", action.Entry.Command, action.Message)
		}
	}

	fmt.Printf("\nSummary: %d installed, %d upgraded, %d skipped, %d failed\n", installed, upgraded, skipped, failed)
	if failed > 0 {
		return fmt.Errorf("apply completed with %d failure(s)", failed)
	}
	return nil
}

func install(ctx context.Context, registry *providers.Registry, entry api.Entry) error {
	if entry.Provider == api.ProviderManual {
		return fmt.Errorf("manual provider cannot install automatically; original path: %v", entry.ProviderArgs["path"])
	}
	provider := registry.Find(entry.Provider)
	if provider == nil {
		return fmt.Errorf("unknown provider %s", entry.Provider)
	}
	plan, err := provider.Resolve(entry)
	if err != nil {
		return err
	}
	fmt.Printf("[%s] %s@%s via %s\n", "install", entry.Command, entry.Version, entry.Provider)
	return registry.Install(ctx, plan)
}
