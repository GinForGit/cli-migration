// Package plan compares a manifest with the current environment.
package plan

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/GinForGit/cli-migration/internal/discover"
	"github.com/GinForGit/cli-migration/internal/platform"
	"github.com/GinForGit/cli-migration/internal/providers"
	"github.com/GinForGit/cli-migration/pkg/api"
)

// Generate creates an action plan for applying a manifest on the current platform.
func Generate(ctx context.Context, plat platform.Platform, m *api.Manifest) (*api.Plan, error) {
	registry := providers.NewRegistry()
	available := registry.Available()
	availableNames := make(map[api.ProviderName]bool)
	for _, p := range available {
		availableNames[p.Name()] = true
	}

	currentEntries := currentEnvironment(ctx)

	var actions []api.Action
	for _, entry := range m.Entries {
		action := resolveAction(entry, currentEntries, availableNames, plat)
		actions = append(actions, action)
	}

	sort.SliceStable(actions, func(i, j int) bool {
		return actions[i].Entry.Command < actions[j].Entry.Command
	})

	return &api.Plan{
		ManifestPath: "",
		TargetOS:     api.OSType(plat.OS()),
		Actions:      actions,
	}, nil
}

func currentEnvironment(ctx context.Context) map[string]api.Entry {
	entries, err := discover.CurrentEnvironment(ctx, false)
	if err != nil {
		return map[string]api.Entry{}
	}
	m := make(map[string]api.Entry, len(entries))
	for _, e := range entries {
		m[e.Command] = e
	}
	return m
}

func resolveAction(entry api.Entry, current map[string]api.Entry, availableNames map[api.ProviderName]bool, plat platform.Platform) api.Action {
	// Cross-platform check is intentionally left for Phase 4.
	_ = plat

	if !availableNames[entry.Provider] {
		return api.Action{
			Kind:    api.ActionUnavailable,
			Entry:   entry,
			Message: fmt.Sprintf("provider %s is not available on this machine", entry.Provider),
		}
	}

	cur, installed := current[entry.Command]
	if !installed {
		return api.Action{Kind: api.ActionInstall, Entry: entry}
	}

	if cur.Version == entry.Version || entry.Version == "unknown" {
		return api.Action{Kind: api.ActionSkip, Entry: entry, Current: cur.Version}
	}

	switch compareVersions(cur.Version, entry.Version) {
	case -1:
		return api.Action{Kind: api.ActionUpgrade, Entry: entry, Current: cur.Version}
	case 1:
		return api.Action{Kind: api.ActionDowngrade, Entry: entry, Current: cur.Version}
	default:
		return api.Action{Kind: api.ActionSkip, Entry: entry, Current: cur.Version}
	}
}

// compareVersions is a very naive semver-like comparison.
// Returns -1 if a < b, 0 if equal, 1 if a > b.
func compareVersions(a, b string) int {
	// Strip leading 'v'.
	a = strings.TrimPrefix(a, "v")
	b = strings.TrimPrefix(b, "v")
	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")
	for i := 0; i < len(partsA) && i < len(partsB); i++ {
		var na, nb int
		fmt.Sscanf(partsA[i], "%d", &na)
		fmt.Sscanf(partsB[i], "%d", &nb)
		if na < nb {
			return -1
		}
		if na > nb {
			return 1
		}
	}
	if len(partsA) < len(partsB) {
		return -1
	}
	if len(partsA) > len(partsB) {
		return 1
	}
	return 0
}

// Format renders a plan as human-readable text.
func Format(p *api.Plan) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Plan for target %s:\n\n", p.TargetOS))
	counts := make(map[api.ActionKind]int)
	for _, a := range p.Actions {
		counts[a.Kind]++
		switch a.Kind {
		case api.ActionInstall:
			b.WriteString(fmt.Sprintf("  [install] %s@%s via %s\n", a.Entry.Command, a.Entry.Version, a.Entry.Provider))
		case api.ActionUpgrade:
			b.WriteString(fmt.Sprintf("  [upgrade] %s@%s via %s (current: %s)\n", a.Entry.Command, a.Entry.Version, a.Entry.Provider, a.Current))
		case api.ActionDowngrade:
			b.WriteString(fmt.Sprintf("  [downgrade] %s@%s via %s (current: %s)\n", a.Entry.Command, a.Entry.Version, a.Entry.Provider, a.Current))
		case api.ActionSkip:
			b.WriteString(fmt.Sprintf("  [skip]    %s already installed (current: %s)\n", a.Entry.Command, a.Current))
		case api.ActionUnavailable:
			b.WriteString(fmt.Sprintf("  [unavailable] %s via %s\n            %s\n", a.Entry.Command, a.Entry.Provider, a.Message))
		case api.ActionWarn:
			b.WriteString(fmt.Sprintf("  [warn]    %s\n            %s\n", a.Entry.Command, a.Message))
		}
	}

	b.WriteString(fmt.Sprintf("\nSummary: %d install, %d upgrade, %d skip, %d unavailable, %d warn\n",
		counts[api.ActionInstall], counts[api.ActionUpgrade], counts[api.ActionSkip],
		counts[api.ActionUnavailable], counts[api.ActionWarn]))
	return b.String()
}
