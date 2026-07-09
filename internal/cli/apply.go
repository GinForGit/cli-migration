package cli

import (
	"fmt"

	"github.com/GinForGit/cli-migration/internal/apply"
	"github.com/GinForGit/cli-migration/internal/manifest"
	"github.com/GinForGit/cli-migration/internal/platform"
	"github.com/GinForGit/cli-migration/internal/plan"
	"github.com/GinForGit/cli-migration/pkg/api"
	"github.com/spf13/cobra"
)

func newApplyCommand() *cobra.Command {
	var manifestPath string
	var dryRun bool
	var skipManual bool
	var targetOS string
	var applyConfigs bool

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "执行还原计划",
		Long:  "根据清单在当前机器上安装、升级或跳过 CLI 工具。",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := manifest.Load(manifestPath)
			if err != nil {
				return err
			}
			plat, err := platform.New()
			if err != nil {
				return err
			}
			os := api.OSType(targetOS)
			if os == "" {
				os = api.OSType(plat.OS())
			}
			p, err := plan.Generate(cmd.Context(), plat, m, os)
			if err != nil {
				return err
			}

			if dryRun {
				fmt.Println("Dry-run mode. Planned actions:")
				fmt.Println(plan.Format(p))
				return nil
			}

			return apply.Execute(cmd.Context(), plat, p, apply.Options{SkipManual: skipManual, ApplyConfigs: applyConfigs})
		},
	}

	cmd.Flags().StringVarP(&manifestPath, "manifest", "m", "", "清单文件路径（必填）")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "只预览，不执行")
	cmd.Flags().BoolVar(&skipManual, "skip-manual", false, "跳过无法自动安装的 manual 条目")
	cmd.Flags().BoolVar(&applyConfigs, "apply-configs", false, "同时应用清单中的 config_refs")
	cmd.Flags().StringVar(&targetOS, "target-os", "", "目标操作系统：windows、linux（默认当前系统）")
	_ = cmd.MarkFlagRequired("manifest")
	return cmd
}
