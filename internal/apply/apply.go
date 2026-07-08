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

	for _, action := range p.Actions {
		switch action.Kind {
		case api.ActionInstall, api.ActionUpgrade, api.ActionDowngrade:
			if err := install(ctx, registry, action.Entry); err != nil {
				return fmt.Errorf("failed to %s %s: %w", action.Kind, action.Entry.Command, err)
			}
		case api.ActionSkip:
			fmt.Printf("[skip] %s\n", action.Entry.Command)
		case api.ActionUnavailable:
			fmt.Printf("[unavailable] %s: %s\n", action.Entry.Command, action.Message)
		case api.ActionWarn:
			fmt.Printf("[warn] %s: %s\n", action.Entry.Command, action.Message)
		}
	}
	return nil
}

func install(ctx context.Context, registry *providers.Registry, entry api.Entry) error {
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
