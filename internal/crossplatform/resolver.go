// Package crossplatform maps CLI entries from one OS to equivalent installation
// plans on another OS. This is intentionally a Phase 4 extension point.
package crossplatform

import (
	"fmt"

	"github.com/GinForGit/cli-migration/pkg/api"
)

// Resolver maps a CLI entry from a source platform to a target platform.
type Resolver interface {
	Resolve(entry api.Entry, sourceOS, targetOS api.OSType) (api.Entry, error)
}

// DefaultResolver is a hard-coded mapping table for common CLI tools.
type DefaultResolver struct{}

// NewDefaultResolver creates the default cross-platform resolver.
func NewDefaultResolver() Resolver {
	return &DefaultResolver{}
}

// mappingKey identifies a provider + package on a source OS.
type mappingKey struct {
	OS       api.OSType
	Provider api.ProviderName
	Package  string
}

// defaultMappings: source key -> target provider + package.
var defaultMappings = map[mappingKey]api.ProviderOverride{
	{OS: api.OSWindows, Provider: api.ProviderScoop, Package: "git"}:       {Provider: api.ProviderApt, ProviderArgs: map[string]interface{}{"package": "git"}},
	{OS: api.OSWindows, Provider: api.ProviderScoop, Package: "nodejs"}:    {Provider: api.ProviderApt, ProviderArgs: map[string]interface{}{"package": "nodejs"}},
	{OS: api.OSWindows, Provider: api.ProviderScoop, Package: "python"}:    {Provider: api.ProviderApt, ProviderArgs: map[string]interface{}{"package": "python3"}},
	{OS: api.OSWindows, Provider: api.ProviderScoop, Package: "curl"}:      {Provider: api.ProviderApt, ProviderArgs: map[string]interface{}{"package": "curl"}},
	{OS: api.OSWindows, Provider: api.ProviderScoop, Package: "wget"}:      {Provider: api.ProviderApt, ProviderArgs: map[string]interface{}{"package": "wget"}},
	{OS: api.OSWindows, Provider: api.ProviderScoop, Package: "htop"}:      {Provider: api.ProviderApt, ProviderArgs: map[string]interface{}{"package": "htop"}},
	{OS: api.OSWindows, Provider: api.ProviderScoop, Package: "ripgrep"}:   {Provider: api.ProviderApt, ProviderArgs: map[string]interface{}{"package": "ripgrep"}},
	{OS: api.OSWindows, Provider: api.ProviderScoop, Package: "fd"}:        {Provider: api.ProviderApt, ProviderArgs: map[string]interface{}{"package": "fd-find"}},
	{OS: api.OSWindows, Provider: api.ProviderScoop, Package: "fzf"}:       {Provider: api.ProviderApt, ProviderArgs: map[string]interface{}{"package": "fzf"}},
	{OS: api.OSWindows, Provider: api.ProviderScoop, Package: "jq"}:        {Provider: api.ProviderApt, ProviderArgs: map[string]interface{}{"package": "jq"}},
	{OS: api.OSWindows, Provider: api.ProviderScoop, Package: "bat"}:       {Provider: api.ProviderApt, ProviderArgs: map[string]interface{}{"package": "bat"}},
	{OS: api.OSWindows, Provider: api.ProviderScoop, Package: "exa"}:       {Provider: api.ProviderApt, ProviderArgs: map[string]interface{}{"package": "exa"}},
	{OS: api.OSWindows, Provider: api.ProviderScoop, Package: "gh"}:        {Provider: api.ProviderApt, ProviderArgs: map[string]interface{}{"package": "gh"}},

	{OS: api.OSLinux, Provider: api.ProviderApt, Package: "git"}:           {Provider: api.ProviderScoop, ProviderArgs: map[string]interface{}{"bucket": "main", "package": "git"}},
	{OS: api.OSLinux, Provider: api.ProviderApt, Package: "nodejs"}:        {Provider: api.ProviderScoop, ProviderArgs: map[string]interface{}{"bucket": "main", "package": "nodejs"}},
	{OS: api.OSLinux, Provider: api.ProviderApt, Package: "python3"}:       {Provider: api.ProviderScoop, ProviderArgs: map[string]interface{}{"bucket": "main", "package": "python"}},
	{OS: api.OSLinux, Provider: api.ProviderApt, Package: "curl"}:          {Provider: api.ProviderScoop, ProviderArgs: map[string]interface{}{"bucket": "main", "package": "curl"}},
	{OS: api.OSLinux, Provider: api.ProviderApt, Package: "wget"}:          {Provider: api.ProviderScoop, ProviderArgs: map[string]interface{}{"bucket": "main", "package": "wget"}},
	{OS: api.OSLinux, Provider: api.ProviderApt, Package: "htop"}:          {Provider: api.ProviderScoop, ProviderArgs: map[string]interface{}{"bucket": "main", "package": "htop"}},
	{OS: api.OSLinux, Provider: api.ProviderApt, Package: "ripgrep"}:       {Provider: api.ProviderScoop, ProviderArgs: map[string]interface{}{"bucket": "main", "package": "ripgrep"}},
	{OS: api.OSLinux, Provider: api.ProviderApt, Package: "fd-find"}:       {Provider: api.ProviderScoop, ProviderArgs: map[string]interface{}{"bucket": "main", "package": "fd"}},
	{OS: api.OSLinux, Provider: api.ProviderApt, Package: "fzf"}:           {Provider: api.ProviderScoop, ProviderArgs: map[string]interface{}{"bucket": "main", "package": "fzf"}},
	{OS: api.OSLinux, Provider: api.ProviderApt, Package: "jq"}:            {Provider: api.ProviderScoop, ProviderArgs: map[string]interface{}{"bucket": "main", "package": "jq"}},
	{OS: api.OSLinux, Provider: api.ProviderApt, Package: "bat"}:           {Provider: api.ProviderScoop, ProviderArgs: map[string]interface{}{"bucket": "main", "package": "bat"}},
	{OS: api.OSLinux, Provider: api.ProviderApt, Package: "gh"}:            {Provider: api.ProviderScoop, ProviderArgs: map[string]interface{}{"bucket": "main", "package": "gh"}},
}

func (r *DefaultResolver) Resolve(entry api.Entry, sourceOS, targetOS api.OSType) (api.Entry, error) {
	pkg := packageName(entry)
	if pkg == "" {
		return api.Entry{}, fmt.Errorf("cannot resolve entry without provider_args.package")
	}

	key := mappingKey{OS: sourceOS, Provider: entry.Provider, Package: pkg}
	override, ok := defaultMappings[key]
	if !ok {
		return api.Entry{}, fmt.Errorf("no cross-platform mapping for %s/%s on %s", entry.Provider, pkg, sourceOS)
	}

	resolved := entry
	resolved.Provider = override.Provider
	resolved.ProviderArgs = override.ProviderArgs
	if resolved.TargetOverrides == nil {
		resolved.TargetOverrides = make(map[api.OSType]api.ProviderOverride)
	}
	resolved.TargetOverrides[targetOS] = override
	return resolved, nil
}

func packageName(entry api.Entry) string {
	// Try common provider_arg keys.
	for _, key := range []string{"package", "package_id"} {
		if v, ok := entry.ProviderArgs[key].(string); ok && v != "" {
			return v
		}
	}
	return ""
}

// CanResolve returns true if a mapping exists for the entry.
func CanResolve(entry api.Entry, sourceOS, targetOS api.OSType) bool {
	if sourceOS == targetOS {
		return false
	}
	pkg := packageName(entry)
	if pkg == "" {
		return false
	}
	_, ok := defaultMappings[mappingKey{OS: sourceOS, Provider: entry.Provider, Package: pkg}]
	return ok
}
