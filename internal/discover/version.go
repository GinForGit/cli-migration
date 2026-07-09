package discover

import (
	"context"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

var versionPatterns = []*regexp.Regexp{
	// Matches "git version 2.45.0" style.
	regexp.MustCompile(`(?i)\bversion\s+v?(\d+(?:\.\d+)*(?:[-+.]?[a-z0-9]+)*)\b`),
	// Matches "go version go1.22.0 windows/amd64" style.
	regexp.MustCompile(`(?i)\bgo(\d+(?:\.\d+)*(?:[-+.]?[a-z0-9]+)*)\b`),
	// Matches "node v20.0.0" or "v20.0.0" style.
	regexp.MustCompile(`(?i)\bv(\d+(?:\.\d+)*(?:[-+.]?[a-z0-9]+)*)\b`),
	// Matches bare "1.2.3" or "1.2.3-beta" style.
	regexp.MustCompile(`(?i)\b(\d+(?:\.\d+)+(?:[-+.]?[a-z0-9]+)*)\b`),
}

// versionProbeTimeout limits how long we wait for a single command invocation.
const versionProbeTimeout = 3 * time.Second

// ProbeVersion attempts to detect the version of a command by running common
// version flags concurrently and parsing the output.
func ProbeVersion(command string) string {
	flags := []string{"--version", "version", "-v", "-V"}

	ctx, cancel := context.WithTimeout(context.Background(), versionProbeTimeout+500*time.Millisecond)
	defer cancel()

	results := make(chan string, len(flags))
	var wg sync.WaitGroup
	for _, flag := range flags {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			if v := tryVersion(ctx, command, f); v != "" {
				select {
				case results <- v:
				default:
				}
			}
		}(flag)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	if v, ok := <-results; ok {
		return v
	}
	return "unknown"
}

func tryVersion(ctx context.Context, command, flag string) string {
	ctx, cancel := context.WithTimeout(ctx, versionProbeTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, flag)
	cmd.Dir = os.TempDir()
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	output := string(out)
	if looksLikeHelpOrError(output) {
		return ""
	}
	return extractVersion(output)
}

// looksLikeHelpOrError rejects output that is clearly not a version string.
func looksLikeHelpOrError(output string) bool {
	lower := strings.ToLower(output)
	return strings.Contains(lower, "usage:") ||
		strings.Contains(lower, "usage ") ||
		strings.Contains(lower, "invalid option") ||
		strings.Contains(lower, "unknown option") ||
		strings.Contains(lower, "unrecognized option") ||
		strings.Contains(lower, "error:") ||
		strings.Contains(lower, "bad option") ||
		strings.Contains(lower, "help")
}

func extractVersion(output string) string {
	firstLine := strings.Split(output, "\n")[0]
	for _, re := range versionPatterns {
		matches := re.FindStringSubmatch(firstLine)
		if len(matches) > 1 {
			candidate := strings.TrimSpace(matches[1])
			if candidate != "" {
				return candidate
			}
		}
	}
	return ""
}
