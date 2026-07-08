package discover

import (
	"context"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var versionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)v?(\d+\.\d+(?:\.\d+)?(?:[-+.]?[a-z0-9]+)*)`),
}

// ProbeVersion attempts to detect the version of a command by running common flags.
func ProbeVersion(command string) string {
	for _, flag := range []string{"--version", "-v", "-V", "version"} {
		v := tryVersion(command, flag)
		if v != "" {
			return v
		}
	}
	return "unknown"
}

func tryVersion(command, flag string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, flag)
	cmd.Dir = os.TempDir()
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return extractVersion(string(out))
}

func extractVersion(output string) string {
	for _, re := range versionPatterns {
		matches := re.FindStringSubmatch(output)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}
	return ""
}
