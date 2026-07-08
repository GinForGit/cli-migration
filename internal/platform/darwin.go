package platform

import (
	"path/filepath"
	"runtime"
)

type darwin struct{}

func newDarwin() Platform { return darwin{} }

func (darwin) OS() OSType { return OSDarwin }

func (darwin) Arch() string { return runtime.GOARCH }

func (darwin) HomeDir() string { return homeDir() }

func (darwin) ExecutableExtensions() []string { return []string{""} }

func (darwin) ListPathEntries() []string {
	// Reuse linux path splitting logic.
	return linux{}.ListPathEntries()
}

func (darwin) ShellConfigFiles() []string {
	home := homeDir()
	return []string{
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".bash_profile"),
		filepath.Join(home, ".bashrc"),
	}
}

func (darwin) IsElevated() bool {
	// Darwin is not a first-class target in this phase.
	return false
}

func (darwin) QuoteCommand(args []string) string { return quoteCommand(args) }

func (darwin) JoinPath(elem ...string) string { return joinPath(elem...) }

func (darwin) Abs(path string) (string, error) { return abs(path) }
