package discover

import (
	"testing"

	"github.com/GinForGit/cli-migration/internal/platform"
)

func TestClassifyPath(t *testing.T) {
	cases := []struct {
		dir    string
		osType platform.OSType
		want   pathCategory
	}{
		{`C:\Windows\System32`, platform.OSWindows, categorySystem},
		{`C:\Windows`, platform.OSWindows, categorySystem},
		{`C:\Users\x\scoop\apps\git\current`, platform.OSWindows, categoryUser},
		{`/usr/bin`, platform.OSLinux, categorySystem},
		{`/usr/local/bin`, platform.OSLinux, categoryUser},
		{`/opt/myapp/bin`, platform.OSLinux, categoryUser},
	}

	for _, c := range cases {
		got := classifyPath(c.dir, c.osType)
		if got != c.want {
			t.Errorf("classifyPath(%q, %s) = %v, want %v", c.dir, c.osType, got, c.want)
		}
	}
}

func TestIsNoisePath(t *testing.T) {
	noise := []string{
		`C:\Users\x\AppData\Local\Microsoft\WindowsApps`,
		`C:\Program Files\Git\usr\bin`,
		`C:\Program Files\Git\mingw64\bin`,
	}
	for _, p := range noise {
		if !isNoisePath(p, platform.OSWindows) {
			t.Errorf("expected %q to be noise", p)
		}
	}
}
