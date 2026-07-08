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

func TestFilterShouldScan(t *testing.T) {
	f := DefaultFilter()
	if !f.ShouldScan(`/usr/local/bin`, platform.OSLinux, 5) {
		t.Error("expected /usr/local/bin with few files to be scanned")
	}
	if f.ShouldScan(`/usr/bin`, platform.OSLinux, 5) {
		t.Error("expected /usr/bin to be skipped as system path")
	}
	if f.ShouldScan(`C:\Program Files\BundledTools`, platform.OSWindows, 100) {
		t.Error("expected crowded non-user directory to be skipped")
	}
	if !f.ShouldScan(`C:\Users\x\scoop\apps\git\current`, platform.OSWindows, 100) {
		t.Error("expected scoop directory to be scanned even if crowded")
	}
}

func TestFilterZeroCrowdedThreshold(t *testing.T) {
	f := &Filter{CrowdedThreshold: 0}
	if !f.ShouldScan(`C:\Tools`, platform.OSWindows, 1000) {
		t.Error("expected crowded directory to be scanned when threshold is disabled")
	}
}
