package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type windows struct{}

func newWindows() Platform { return windows{} }

func (windows) OS() OSType { return OSWindows }

func (windows) Arch() string { return runtime.GOARCH }

func (windows) HomeDir() string { return homeDir() }

var windowsExts = []string{".exe", ".cmd", ".bat", ".ps1", ".com"}

func (windows) ExecutableExtensions() []string {
	return windowsExts
}

func (windows) ListPathEntries() []string {
	return splitPath(os.Getenv("PATH"))
}

func (windows) ShellConfigFiles() []string {
	// PowerShell and Command Prompt profiles.
	home := homeDir()
	return []string{
		filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1"),
		filepath.Join(home, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1"),
	}
}

func (windows) IsElevated() bool {
	// A simple heuristic: check if we can write to the system-wide Program Files.
	testPath := filepath.Join(os.Getenv("ProgramFiles"), ".cli-mig-admin-test")
	f, err := os.Create(testPath)
	if err != nil {
		return false
	}
	_ = f.Close()
	_ = os.Remove(testPath)
	return true
}

func (windows) QuoteCommand(args []string) string { return quoteCommand(args) }

func (windows) JoinPath(elem ...string) string { return joinPath(elem...) }

func (windows) Abs(path string) (string, error) { return abs(path) }

// hasExtension checks if a name ends with any known executable extension.
func hasExtension(name string) bool {
	lower := strings.ToLower(name)
	for _, ext := range windowsExts {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

// normalizeExecutableName strips the extension from a Windows executable name.
func normalizeExecutableName(name string) string {
	lower := strings.ToLower(name)
	for _, ext := range windowsExts {
		if strings.HasSuffix(lower, ext) {
			return name[:len(name)-len(ext)]
		}
	}
	return name
}
