package discover

import (
	"path/filepath"
	"strings"

	"github.com/GinForGit/cli-migration/internal/platform"
)

// Filter decides which PATH directories should be scanned for CLI tools.
type Filter struct {
	CrowdedThreshold int
}

// DefaultFilter returns a filter with sensible defaults.
func DefaultFilter() *Filter {
	return &Filter{CrowdedThreshold: 50}
}

// ShouldScan returns true if a directory should be scanned.
func (f *Filter) ShouldScan(dir string, osType platform.OSType, fileCount int) bool {
	if isSystemPath(dir, osType) {
		return false
	}
	if isNoisePath(dir, osType) {
		return false
	}
	// Directories with many executables are usually bundled utility collections
	// (e.g. Git for Windows usr/bin) rather than individually installed CLI tools.
	if f.CrowdedThreshold > 0 && fileCount > f.CrowdedThreshold && !isUserPath(dir, osType) {
		return false
	}
	return true
}

// isSystemPath returns true for directories that contain operating-system
// executables rather than user-installed CLI tools.
func isSystemPath(dir string, osType platform.OSType) bool {
	lower := strings.ToLower(dir)
	if osType == platform.OSWindows {
		return hasPrefixAny(lower, []string{
			`c:\windows`,
			`c:\windows\`,
		})
	}
	systemPaths := []string{
		"/bin", "/sbin", "/usr/bin", "/usr/sbin",
		"/lib", "/usr/lib", "/lib64", "/usr/lib64",
	}
	for _, p := range systemPaths {
		if lower == p || strings.HasPrefix(lower, p+"/") {
			return true
		}
	}
	return false
}

// isUserPath returns true for directories that typically contain user-installed tools.
func isUserPath(dir string, osType platform.OSType) bool {
	lower := strings.ToLower(dir)
	if osType == platform.OSWindows {
		return containsAny(lower, []string{
			`\scoop\`,
			`\programdata\chocolatey`,
			`\.local\bin`,
			`\.cargo\bin`,
			`\nodejs`,
			`\program files\nodejs`,
			`\program files\dotnet`,
			`\appdata\local\programs`,
			`\appdata\local\hermes`,
		})
	}
	return containsAny(lower, []string{
		"/home/",
		"/opt/",
		"/usr/local/",
		"/.cargo/",
		"/.local/",
		"/usr/local/go/",
	})
}

// isNoisePath returns true for directories that contain bundled utilities we
// do not want to record as standalone CLI tools.
func isNoisePath(dir string, osType platform.OSType) bool {
	lower := strings.ToLower(dir)
	if osType == platform.OSWindows {
		return containsAny(lower, []string{
			`\microsoft\windowsapps`,
			`\program files\git\usr\bin`,
			`\program files\git\mingw64\bin`,
			`\program files\git\cmd`,
			`\venv\scripts`,
			`\.venv\scripts`,
		})
	}
	return containsAny(lower, []string{
		"/usr/lib/git-core/",
	})
}

func containsAny(s string, subs []string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func hasPrefixAny(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

// normalizePath returns a clean absolute-style lowercased path for comparison.
func normalizePath(p string) string {
	return strings.ToLower(filepath.Clean(p))
}

// pathCategory is kept for backward compatibility with existing tests.
type pathCategory int

const (
	categoryUnknown pathCategory = iota
	categorySystem
	categoryUser
)

// classifyPath classifies a directory for tests.
func classifyPath(dir string, osType platform.OSType) pathCategory {
	if isSystemPath(dir, osType) {
		return categorySystem
	}
	if isUserPath(dir, osType) {
		return categoryUser
	}
	return categoryUnknown
}
