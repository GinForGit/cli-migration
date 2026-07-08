package platform

import (
	"os"
	"path/filepath"
	"runtime"
)

type linux struct{}

func newLinux() Platform { return linux{} }

func (linux) OS() OSType { return OSLinux }

func (linux) Arch() string { return runtime.GOARCH }

func (linux) HomeDir() string { return homeDir() }

func (linux) ExecutableExtensions() []string { return []string{""} }

func (linux) ListPathEntries() []string {
	return splitPath(os.Getenv("PATH"))
}

func (linux) ShellConfigFiles() []string {
	home := homeDir()
	return []string{
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".config", "fish", "config.fish"),
	}
}

func (linux) IsElevated() bool {
	return os.Geteuid() == 0
}

func (linux) QuoteCommand(args []string) string { return quoteCommand(args) }

func (linux) JoinPath(elem ...string) string { return joinPath(elem...) }

func (linux) Abs(path string) (string, error) { return abs(path) }
