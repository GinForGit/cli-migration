// Package configs collects and applies shell aliases, environment variables,
// and dotfiles associated with CLI tools.
package configs

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/GinForGit/cli-migration/internal/platform"
	"github.com/GinForGit/cli-migration/pkg/api"
)

// Collect scans shell configuration files and common dotfiles, then attaches
// matching ConfigRef records to the supplied entries.
func Collect(plat platform.Platform, entries []api.Entry) ([]api.Entry, error) {
	shellConfigs := plat.ShellConfigFiles()
	aliases, envs, err := scanShellConfigs(shellConfigs)
	if err != nil {
		return nil, err
	}

	byCommand := make(map[string]int)
	for i := range entries {
		byCommand[entries[i].Command] = i
	}

	for _, a := range aliases {
		idx, ok := byCommand[a.Command]
		if !ok {
			continue
		}
		entries[idx].ConfigRefs = append(entries[idx].ConfigRefs, api.ConfigRef{
			Type:   "alias",
			Key:    a.Name,
			Value:  a.Value,
			Source: a.Source,
		})
	}

	for _, e := range envs {
		idx := bestEnvMatch(e.Key, entries)
		if idx < 0 {
			continue
		}
		entries[idx].ConfigRefs = append(entries[idx].ConfigRefs, api.ConfigRef{
			Type:   "env",
			Key:    e.Key,
			Value:  e.Value,
			Source: e.Source,
		})
	}

	for i := range entries {
		files := collectDotfiles(plat.HomeDir(), entries[i].Command)
		for _, f := range files {
			entries[i].ConfigRefs = append(entries[i].ConfigRefs, api.ConfigRef{
				Type:   "file",
				Source: f.source,
				Target: f.target,
			})
		}
	}

	return entries, nil
}

// Apply writes the ConfigRefs for an entry to the target system.
// It skips refs whose source file no longer exists.
func Apply(plat platform.Platform, entry api.Entry) error {
	if len(entry.ConfigRefs) == 0 {
		return nil
	}

	var shellTargets []string
	for _, ref := range entry.ConfigRefs {
		switch ref.Type {
		case "alias", "env":
			shellTargets = append(shellTargets, ref.Source)
		case "file":
			if err := applyFile(plat, ref); err != nil {
				return fmt.Errorf("apply file config for %s: %w", entry.Command, err)
			}
		}
	}

	if err := applyShellRefs(plat, shellTargets, entry); err != nil {
		return fmt.Errorf("apply shell config for %s: %w", entry.Command, err)
	}
	return nil
}

// ---------- shell config parsing ----------

type aliasEntry struct {
	Name   string
	Value  string
	Source string
	Command string
}

type envEntry struct {
	Key    string
	Value  string
	Source string
}

var (
	bashAliasRe  = regexp.MustCompile(`^\s*alias\s+([A-Za-z0-9_\-]+)=(?:["']?)(.+?)(?:["']?)\s*(?:#.*)?$`)
	bashExportRe = regexp.MustCompile(`^\s*export\s+([A-Za-z0-9_]+)=(?:["']?)(.*?)(?:["']?)\s*(?:#.*)?$`)
	fishSetRe    = regexp.MustCompile(`^\s*set\s+(?:-gx?\s+)?([A-Za-z0-9_]+)\s+(.+?)\s*(?:#.*)?$`)
	psAliasRe    = regexp.MustCompile(`^\s*Set-Alias\s+(?:-Name\s+)?([A-Za-z0-9_\-]+)\s+(?:-Value\s+)?([\S]+)`)
	psEnvRe      = regexp.MustCompile(`^\s*\$env:([A-Za-z0-9_]+)\s*=\s*(.+)$`)
)

func scanShellConfigs(paths []string) ([]aliasEntry, []envEntry, error) {
	var aliases []aliasEntry
	var envs []envEntry

	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, nil, err
		}

		lines := strings.Split(string(data), "\n")
		isFish := strings.Contains(filepath.Base(p), "fish")
		isPS := strings.HasSuffix(strings.ToLower(p), ".ps1")

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			if isPS {
				if m := psAliasRe.FindStringSubmatch(line); m != nil {
					aliases = append(aliases, aliasEntry{
						Name:    m[1],
						Value:   m[2],
						Source:  p,
						Command: commandFromAliasValue(m[2]),
					})
					continue
				}
				if m := psEnvRe.FindStringSubmatch(line); m != nil {
					envs = append(envs, envEntry{Key: m[1], Value: strings.TrimSpace(m[2]), Source: p})
					continue
				}
				continue
			}

			if m := bashAliasRe.FindStringSubmatch(line); m != nil {
				aliases = append(aliases, aliasEntry{
					Name:    m[1],
					Value:   m[2],
					Source:  p,
					Command: commandFromAliasValue(m[2]),
				})
				continue
			}

			if isFish {
				if m := fishSetRe.FindStringSubmatch(line); m != nil {
					envs = append(envs, envEntry{Key: m[1], Value: strings.TrimSpace(m[2]), Source: p})
				}
				continue
			}

			if m := bashExportRe.FindStringSubmatch(line); m != nil {
				envs = append(envs, envEntry{Key: m[1], Value: strings.TrimSpace(m[2]), Source: p})
			}
		}
	}

	return aliases, envs, nil
}

func commandFromAliasValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"'`)
	if value == "" {
		return ""
	}
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return ""
	}
	return filepath.Base(fields[0])
}

func bestEnvMatch(key string, entries []api.Entry) int {
	lower := strings.ToLower(key)
	best := -1
	bestLen := 0
	for i, e := range entries {
		cmdLower := strings.ToLower(e.Command)
		if strings.HasPrefix(lower, cmdLower) && len(cmdLower) > bestLen {
			best = i
			bestLen = len(cmdLower)
		}
	}
	return best
}

// ---------- dotfiles ----------

type dotfile struct {
	source string
	target string
}

func collectDotfiles(home, command string) []dotfile {
	var result []dotfile
	for _, mapping := range knownDotfiles {
		if !strings.Contains(strings.ToLower(mapping.command), strings.ToLower(command)) {
			continue
		}
		src := filepath.Join(home, mapping.path)
		if _, err := os.Stat(src); err != nil {
			continue
		}
		result = append(result, dotfile{source: src, target: mapping.path})
	}
	return result
}

type dotfileMapping struct {
	command string
	path    string
}

var knownDotfiles = []dotfileMapping{
	{"git", ".gitconfig"},
	{"git", ".gitignore_global"},
	{"npm", ".npmrc"},
	{"ssh", ".ssh/config"},
	{"vim", ".vimrc"},
	{"nvim", ".config/nvim/init.vim"},
	{"nvim", ".config/nvim/init.lua"},
	{"tmux", ".tmux.conf"},
}

// ---------- application ----------

func applyFile(plat platform.Platform, ref api.ConfigRef) error {
	src := ref.Source
	if src == "" {
		return nil
	}
	if _, err := os.Stat(src); err != nil {
		return nil
	}

	target := ref.Target
	if target == "" {
		target = filepath.Base(src)
	}
	dest := filepath.Join(plat.HomeDir(), target)

	if _, err := os.Stat(dest); err == nil {
		fmt.Printf("[warn] config file already exists, skipping: %s\n", dest)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dest, data, 0o644)
}

func applyShellRefs(plat platform.Platform, sourceFiles []string, entry api.Entry) error {
	if len(sourceFiles) == 0 {
		return nil
	}

	targets := plat.ShellConfigFiles()
	if len(targets) == 0 {
		return nil
	}
	target := targets[0]

	var lines []string
	for _, ref := range entry.ConfigRefs {
		switch ref.Type {
		case "alias":
			lines = append(lines, fmt.Sprintf("alias %s='%s'", ref.Key, ref.Value))
		case "env":
			lines = append(lines, fmt.Sprintf("export %s=%s", ref.Key, ref.Value))
		}
	}
	if len(lines) == 0 {
		return nil
	}

	marker := fmt.Sprintf("# cli-mig: %s", entry.Command)
	content, err := os.ReadFile(target)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if strings.Contains(string(content), marker) {
		return nil
	}

	f, err := os.OpenFile(target, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	_, _ = w.WriteString("\n" + marker + "\n")
	for _, line := range lines {
		_, _ = w.WriteString(line + "\n")
	}
	return w.Flush()
}
