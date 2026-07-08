// Package providers abstracts package managers and installation sources.
package providers

import (
	"context"
	"fmt"

	"github.com/GinForGit/cli-migration/pkg/api"
)

// InstallPlan describes how to install a specific entry.
type InstallPlan struct {
	Provider     api.ProviderName
	Package      string
	Version      string
	Args         map[string]interface{}
	NeedsElevated bool
}

// Provider is the interface implemented by every package manager source.
type Provider interface {
	Name() api.ProviderName
	Available() bool
	Detect(ctx context.Context) ([]api.Entry, error)
	Resolve(entry api.Entry) (InstallPlan, error)
	Install(ctx context.Context, plan InstallPlan) error
	Uninstall(ctx context.Context, plan InstallPlan) error
}

// BaseProvider implements common helpers for concrete providers.
type BaseProvider struct {
	name api.ProviderName
}

func (b BaseProvider) Name() api.ProviderName { return b.name }

// ErrUnavailable is returned when a provider cannot be used.
var ErrUnavailable = fmt.Errorf("provider unavailable")

// ResolveDefault returns a simple install plan based on provider_args.package.
func ResolveDefault(entry api.Entry) (InstallPlan, error) {
	pkg, _ := entry.ProviderArgs["package"].(string)
	if pkg == "" {
		return InstallPlan{}, fmt.Errorf("provider_args.package is required for %s", entry.Provider)
	}
	return InstallPlan{
		Provider: entry.Provider,
		Package:  pkg,
		Version:  entry.Version,
		Args:     entry.ProviderArgs,
	}, nil
}
