package discover

import (
	"path/filepath"
	"strings"

	"github.com/GinForGit/cli-migration/internal/platform"
)

// pathCategory classifies a PATH directory.
type pathCategory int

const (
	categoryUnknown pathCategory = iota
	categorySystem
	categoryUser
)

// classifyPath returns whether a directory is likely a system or user directory.
func classifyPath(dir string, osType platform.OSType) pathCategory {
	if isSystemPath(dir, osType) {
		return categorySystem
	}
	if isUserPath(dir, osType) {
		return categoryUser
	}
	return categoryUnknown
}

// isSystemPath returns true for directories that contain operating-system
// executables rather than user-installed CLI tools.
func isSystemPath(dir string, osType platform.OSType) bool {
	lower := strings.ToLower(dir)
	if osType == platform.OSWindows {
		// Match C:\Windows, C:\Windows\System32, C:\Windows\SysWOW64, etc.
		return hasPrefixAny(lower, []string{
			`c:\windows`,
			`c:\windows\`,
		})
	}
	// Linux: exact match or subdir of common system paths.
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
