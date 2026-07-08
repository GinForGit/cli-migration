// Package manifest handles serialization and validation of environment manifests.
package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/GinForGit/cli-migration/pkg/api"
	"gopkg.in/yaml.v3"
)

const CurrentVersion = "1.0"

// Load reads a manifest from a YAML or JSON file.
func Load(path string) (*api.Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	var m api.Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	if err := Validate(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Save writes a manifest to a YAML file.
func Save(path string, m *api.Manifest) error {
	if err := Validate(m); err != nil {
		return err
	}
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}
	return nil
}

// SaveJSON writes a manifest to a JSON file.
func SaveJSON(path string, m *api.Manifest) error {
	if err := Validate(m); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}
	return nil
}

// Validate checks that a manifest is well-formed.
func Validate(m *api.Manifest) error {
	if m.Version == "" {
		return fmt.Errorf("manifest version is required")
	}
	if m.GeneratedAt.IsZero() {
		return fmt.Errorf("manifest generated_at is required")
	}
	if m.Source.OS == "" {
		return fmt.Errorf("manifest source.os is required")
	}
	for i, e := range m.Entries {
		if e.Name == "" {
			return fmt.Errorf("entry %d: name is required", i)
		}
		if e.Command == "" {
			return fmt.Errorf("entry %d: command is required", i)
		}
		if e.Provider == "" {
			return fmt.Errorf("entry %d: provider is required", i)
		}
	}
	return nil
}

// New creates a new manifest with required metadata filled in.
func New(source api.SourceInfo) *api.Manifest {
	return &api.Manifest{
		Version:     CurrentVersion,
		GeneratedAt: time.Now().UTC(),
		Source:      source,
		Entries:     []api.Entry{},
	}
}
