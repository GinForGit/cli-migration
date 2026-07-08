package providers

import (
	"context"
	"fmt"
	"sort"

	"github.com/GinForGit/cli-migration/pkg/api"
)

// Registry holds all registered providers.
type Registry struct {
	providers []Provider
}

// NewRegistry creates a registry with built-in providers.
func NewRegistry() *Registry {
	return &Registry{
		providers: []Provider{
			NewScoopProvider(),
			NewAptProvider(),
			NewManualProvider(),
		},
	}
}

// Register adds a provider to the registry.
func (r *Registry) Register(p Provider) {
	r.providers = append(r.providers, p)
}

// All returns all registered providers.
func (r *Registry) All() []Provider {
	return r.providers
}

// Available returns providers that are usable on the current system.
func (r *Registry) Available() []Provider {
	var out []Provider
	for _, p := range r.providers {
		if p.Available() {
			out = append(out, p)
		}
	}
	return out
}

// Find returns the provider with the given name, or nil if not found.
func (r *Registry) Find(name api.ProviderName) Provider {
	for _, p := range r.providers {
		if p.Name() == name {
			return p
		}
	}
	return nil
}

// DetectAll runs Detect on every available provider and merges the results.
// Entries from earlier providers take precedence unless overrideByName is set.
func (r *Registry) DetectAll(ctx context.Context) ([]api.Entry, error) {
	available := r.Available()
	byCommand := make(map[string]api.Entry)

	for _, p := range available {
		entries, err := p.Detect(ctx)
		if err != nil {
			// Log but continue: one failing provider should not break discovery.
			continue
		}
		for _, e := range entries {
			existing, ok := byCommand[e.Command]
			if !ok {
				byCommand[e.Command] = e
				continue
			}
			// Prefer non-manual providers over manual ones.
			if existing.Provider == api.ProviderManual && e.Provider != api.ProviderManual {
				byCommand[e.Command] = e
			}
		}
	}

	var commands []string
	for cmd := range byCommand {
		commands = append(commands, cmd)
	}
	sort.Strings(commands)

	var result []api.Entry
	for _, cmd := range commands {
		result = append(result, byCommand[cmd])
	}
	return result, nil
}

// Install executes the install plan using the appropriate provider.
func (r *Registry) Install(ctx context.Context, plan InstallPlan) error {
	p := r.Find(plan.Provider)
	if p == nil {
		return fmt.Errorf("unknown provider: %s", plan.Provider)
	}
	return p.Install(ctx, plan)
}
