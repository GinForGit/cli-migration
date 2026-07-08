package cli

import (
	"fmt"

	"github.com/GinForGit/cli-migration/internal/discover"
	"github.com/spf13/cobra"
)

func newDiscoverCommand() *cobra.Command {
	var output string
	var format string
	var probeVersions bool

	cmd := &cobra.Command{
		Use:   "discover",
		Short: "扫描当前系统的 CLI 环境",
		Long:  "发现当前系统中已安装的所有 CLI 工具，并生成清单文件。",
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, source, err := discover.Scan(cmd.Context(), probeVersions)
			if err != nil {
				return err
			}
			return discover.WriteManifest(output, format, source, entries)
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "env.yaml", "输出文件路径")
	cmd.Flags().StringVarP(&format, "format", "f", "yaml", "输出格式：yaml 或 json")
	cmd.Flags().BoolVar(&probeVersions, "probe-versions", false, "探测 manual 条目的版本（较慢）")
	return cmd
}

func init() {
	_ = fmt.Sprintf
}
