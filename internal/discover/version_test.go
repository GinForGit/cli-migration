package discover

import (
	"testing"
)

func TestExtractVersion(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"git version 2.45.0", "2.45.0"},
		{"node v20.0.0", "20.0.0"},
		{"go version go1.22.0 windows/amd64", "1.22.0"},
		{"rg 14.1.0", "14.1.0"},
		{"fzf 0.52.1", "0.52.1"},
		{"1.2.3-beta", "1.2.3-beta"},
		{"my-tool 1.0", "1.0"},
		{"Usage: my-tool [--version]", ""},
		{"invalid option --version", ""},
	}

	for _, c := range cases {
		got := extractVersion(c.input)
		if got != c.want {
			t.Errorf("extractVersion(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestLooksLikeHelpOrError(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"Usage: foo [options]", true},
		{"invalid option --version", true},
		{"error: unknown flag", true},
		{"git version 2.45.0", false},
	}

	for _, c := range cases {
		got := looksLikeHelpOrError(c.input)
		if got != c.want {
			t.Errorf("looksLikeHelpOrError(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}
