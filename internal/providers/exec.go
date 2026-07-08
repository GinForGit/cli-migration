package providers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// runCommand executes a shell command and returns a formatted error on failure.
func runCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("empty command")
	}
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %s: %w", strings.Join(args, " "), err)
	}
	return nil
}
