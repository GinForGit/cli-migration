// Package cli implements the command-line interface.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// NewRootCommand builds the root cobra command.
func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "cli-mig",
		Short: "CLI 无痛搬家工具",
		Long:  "cli-mig 帮助你在不同机器之间迁移命令行工具环境。",
	}

	root.AddCommand(newVersionCommand())
	root.AddCommand(newDiscoverCommand())
	root.AddCommand(newPlanCommand())
	root.AddCommand(newApplyCommand())

	return root
}

// Execute runs the CLI.
func Execute() {
	if err := NewRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// SetBuildInfo sets version metadata.
func SetBuildInfo(v, c, d string) {
	version = v
	commit = c
	date = d
}
