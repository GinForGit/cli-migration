// Package platform abstracts operating-system-specific details.
package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Platform provides OS-specific information and operations.
type Platform interface {
	OS() OSType
	Arch() string
	HomeDir() string
	ExecutableExtensions() []string
	ListPathEntries() []string
	ShellConfigFiles() []string
	IsElevated() bool
	QuoteCommand(args []string) string
	JoinPath(elem ...string) string
	Abs(path string) (string, error)
}

// OSType matches api.OSType to avoid importing api in low-level packages.
type OSType string

const (
	OSWindows OSType = "windows"
	OSLinux   OSType = "linux"
	OSDarwin  OSType = "darwin"
)

// New returns the Platform implementation for the current OS.
func New() (Platform, error) {
	switch runtime.GOOS {
	case "windows":
		return newWindows(), nil
	case "linux":
		return newLinux(), nil
	case "darwin":
		return newDarwin(), nil
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// Common helpers.

func splitPath(pathEnv string) []string {
	var sep string
	if runtime.GOOS == "windows" {
		sep = ";"
	} else {
		sep = ":"
	}
	parts := strings.Split(pathEnv, sep)
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

func homeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return h
}

func quoteCommand(args []string) string {
	for i, a := range args {
		if strings.ContainsAny(a, " \"\t\n\r") {
			args[i] = `"` + strings.ReplaceAll(a, `"`, `\"`) + `"`
		}
	}
	return strings.Join(args, " ")
}

func joinPath(elem ...string) string {
	return filepath.Join(elem...)
}

func abs(path string) (string, error) {
	return filepath.Abs(path)
}
