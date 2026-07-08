// Package discover scans the current system for installed CLI tools.
package discover

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/GinForGit/cli-migration/internal/manifest"
	"github.com/GinForGit/cli-migration/internal/platform"
	"github.com/GinForGit/cli-migration/internal/providers"
	"github.com/GinForGit/cli-migration/pkg/api"
)

// Scan discovers installed CLI entries and the source platform info.
func Scan(ctx context.Context, probeVersions bool) ([]api.Entry, api.SourceInfo, error) {
	plat, err := platform.New()
	if err != nil {
		return nil, api.SourceInfo{}, err
	}

	registry := providers.NewRegistry()
	providerEntries, err := registry.DetectAll(ctx)
	if err != nil {
		return nil, api.SourceInfo{}, fmt.Errorf("detect providers: %w", err)
	}

	// PATH scan for executables not claimed by providers.
	pathEntries, err := scanPath(ctx, plat, probeVersions)
	if err != nil {
		return nil, api.SourceInfo{}, fmt.Errorf("scan PATH: %w", err)
	}

	merged := mergeEntries(providerEntries, pathEntries)

	source := api.SourceInfo{
		OS:              api.OSType(plat.OS()),
		Arch:            plat.Arch(),
		Shell:           detectShell(),
		PlatformVersion: "",
	}
	return merged, source, nil
}

// WriteManifest writes discovered entries to a manifest file.
func WriteManifest(path, format string, source api.SourceInfo, entries []api.Entry) error {
	m := manifest.New(source)
	m.Entries = entries

	switch strings.ToLower(format) {
	case "json":
		return manifest.SaveJSON(path, m)
	case "yaml", "yml":
		return manifest.Save(path, m)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func scanPath(ctx context.Context, plat platform.Platform, probeVersions bool) ([]api.Entry, error) {
	paths := plat.ListPathEntries()
	seen := make(map[string]bool)
	var entries []api.Entry

	exts := plat.ExecutableExtensions()

	for _, dir := range paths {
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		// Skip pure system directories and bundled utility directories.
		cat := classifyPath(dir, plat.OS())
		if cat == categorySystem || isNoisePath(dir, plat.OS()) {
			continue
		}
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() {
				continue
			}
			name := f.Name()
			if !isExecutable(name, plat.OS(), exts) {
				continue
			}
			command := commandName(name, plat.OS())
			if seen[command] {
				continue
			}
			seen[command] = true
			fullPath := filepath.Join(dir, name)
			entry := providers.NewManualEntry(fullPath)
			if probeVersions {
				entry.Version = ProbeVersion(command)
			}
			entries = append(entries, entry)
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Command < entries[j].Command
	})
	return entries, nil
}

func isExecutable(name string, osType platform.OSType, exts []string) bool {
	if osType == platform.OSWindows {
		lower := strings.ToLower(name)
		for _, ext := range exts {
			if ext != "" && strings.HasSuffix(lower, ext) {
				return true
			}
		}
		return false
	}
	info, err := os.Stat(name)
	if err != nil {
		// name is just the basename; we need the full path for permission check.
		// For simplicity, accept all regular files on non-Windows.
		return true
	}
	return !info.IsDir() && info.Mode().Perm()&011 != 0
}

func commandName(name string, osType platform.OSType) string {
	if osType == platform.OSWindows {
		ext := filepath.Ext(name)
		return name[:len(name)-len(ext)]
	}
	return name
}

func mergeEntries(providerEntries, pathEntries []api.Entry) []api.Entry {
	byCommand := make(map[string]api.Entry)
	for _, e := range providerEntries {
		byCommand[e.Command] = e
	}
	for _, e := range pathEntries {
		if _, ok := byCommand[e.Command]; !ok {
			byCommand[e.Command] = e
		}
	}

	commands := make([]string, 0, len(byCommand))
	for c := range byCommand {
		commands = append(commands, c)
	}
	sort.Strings(commands)

	result := make([]api.Entry, 0, len(commands))
	for _, c := range commands {
		result = append(result, byCommand[c])
	}
	return result
}

func detectShell() string {
	if runtime.GOOS == "windows" {
		if p := os.Getenv("PSModulePath"); p != "" {
			return "powershell"
		}
		return "cmd"
	}
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "sh"
	}
	return filepath.Base(shell)
}
