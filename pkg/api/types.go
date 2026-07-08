// Package api contains stable types that may be used by external consumers.
package api

import "time"

// OSType represents an operating system family.
type OSType string

const (
	OSWindows OSType = "windows"
	OSLinux   OSType = "linux"
	OSDarwin  OSType = "darwin"
)

// ProviderName identifies a package manager or installation source.
type ProviderName string

const (
	ProviderScoop    ProviderName = "scoop"
	ProviderWinget   ProviderName = "winget"
	ProviderApt      ProviderName = "apt"
	ProviderDnf      ProviderName = "dnf"
	ProviderPacman   ProviderName = "pacman"
	ProviderHomebrew ProviderName = "homebrew"
	ProviderCargo    ProviderName = "cargo"
	ProviderNpm      ProviderName = "npm"
	ProviderPnpm     ProviderName = "pnpm"
	ProviderYarn     ProviderName = "yarn"
	ProviderPipx     ProviderName = "pipx"
	ProviderManual   ProviderName = "manual"
)

// Entry describes a single installed CLI tool.
type Entry struct {
	Name         string                 `json:"name" yaml:"name"`
	Command      string                 `json:"command" yaml:"command"`
	Version      string                 `json:"version" yaml:"version"`
	Provider     ProviderName           `json:"provider" yaml:"provider"`
	ProviderArgs map[string]interface{} `json:"provider_args,omitempty" yaml:"provider_args,omitempty"`
	Aliases      []string               `json:"aliases,omitempty" yaml:"aliases,omitempty"`
	ConfigRefs   []ConfigRef            `json:"config_refs,omitempty" yaml:"config_refs,omitempty"`
}

// ConfigRef records an environment variable, alias, or file associated with a CLI.
type ConfigRef struct {
	Type   string `json:"type" yaml:"type"`
	Key    string `json:"key,omitempty" yaml:"key,omitempty"`
	Value  string `json:"value,omitempty" yaml:"value,omitempty"`
	Source string `json:"source,omitempty" yaml:"source,omitempty"`
	Target string `json:"target,omitempty" yaml:"target,omitempty"`
}

// SourceInfo describes the machine that generated a manifest.
type SourceInfo struct {
	OS              OSType `json:"os" yaml:"os"`
	Arch            string `json:"arch" yaml:"arch"`
	Shell           string `json:"shell,omitempty" yaml:"shell,omitempty"`
	PlatformVersion string `json:"platform_version,omitempty" yaml:"platform_version,omitempty"`
}

// TargetOverride allows specifying different provider arguments per target OS.
type TargetOverride map[OSType]ProviderOverride

// ProviderOverride describes how to install an entry on a specific OS.
type ProviderOverride struct {
	Provider     ProviderName           `json:"provider" yaml:"provider"`
	ProviderArgs map[string]interface{} `json:"provider_args,omitempty" yaml:"provider_args,omitempty"`
}

// Manifest is the serialized representation of a CLI environment.
type Manifest struct {
	Version         string         `json:"version" yaml:"version"`
	GeneratedAt     time.Time      `json:"generated_at" yaml:"generated_at"`
	Source          SourceInfo     `json:"source" yaml:"source"`
	Entries         []Entry        `json:"entries" yaml:"entries"`
	TargetOverrides TargetOverride `json:"target_overrides,omitempty" yaml:"target_overrides,omitempty"`
}

// ActionKind describes what should happen to an entry during apply.
type ActionKind string

const (
	ActionInstall  ActionKind = "install"
	ActionUpgrade  ActionKind = "upgrade"
	ActionDowngrade ActionKind = "downgrade"
	ActionSkip     ActionKind = "skip"
	ActionRemove   ActionKind = "remove"
	ActionUnavailable ActionKind = "unavailable"
	ActionWarn     ActionKind = "warn"
)

// Action is a single planned operation.
type Action struct {
	Kind     ActionKind `json:"kind" yaml:"kind"`
	Entry    Entry      `json:"entry" yaml:"entry"`
	Current  string     `json:"current,omitempty" yaml:"current,omitempty"`
	Message  string     `json:"message,omitempty" yaml:"message,omitempty"`
}

// Plan is the output of the planning phase.
type Plan struct {
	ManifestPath string   `json:"manifest_path" yaml:"manifest_path"`
	TargetOS     OSType   `json:"target_os" yaml:"target_os"`
	Actions      []Action `json:"actions" yaml:"actions"`
}
