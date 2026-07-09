package cli

import (
	"fmt"

	"github.com/GinForGit/cli-migration/internal/crossplatform"
	"github.com/GinForGit/cli-migration/internal/discover"
	"github.com/GinForGit/cli-migration/pkg/api"
	"github.com/spf13/cobra"
)

func newDiscoverCommand() *cobra.Command {
	var output string
	var format string
	var probeVersions bool
	var targetOS string
	var includeConfigs bool

	cmd := &cobra.Command{
		Use:   "discover",
		Short: "扫描当前系统的 CLI 环境",
		Long:  "发现当前系统中已安装的所有 CLI 工具，并生成清单文件。",
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, source, err := discover.Scan(cmd.Context(), probeVersions, includeConfigs)
			if err != nil {
				return err
			}
			if targetOS != "" {
				entries = resolveTargetOverrides(entries, source.OS, api.OSType(targetOS))
			}
			return discover.WriteManifest(output, format, source, entries)
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "env.yaml", "输出文件路径")
	cmd.Flags().StringVarP(&format, "format", "f", "yaml", "输出格式：yaml 或 json")
	cmd.Flags().BoolVar(&probeVersions, "probe-versions", false, "探测 manual 条目的版本（较慢）")
	cmd.Flags().BoolVar(&includeConfigs, "include-configs", false, "收集 shell alias、环境变量和 dotfiles")
	cmd.Flags().StringVar(
	&targetOS, "target-os", "", "为指定目标系统生成 target_overrides")
	return cmd
}

func resolveTargetOverrides(entries []api.Entry, sourceOS, targetOS api.OSType) []api.Entry {
	if sourceOS == targetOS {
		return entries
	}
	resolver := crossplatform.NewDefaultResolver()
	var result []api.Entry
	for _, e := range entries {
		if crossplatform.CanResolve(e, sourceOS, targetOS) {
			resolved, err := resolver.Resolve(e, sourceOS, targetOS)
			if err == nil {
				result = append(result, resolved)
				continue
			}
		}
		result = append(result, e)
	}
	return result
}

func init() {
	_ = fmt.Sprintf
}
