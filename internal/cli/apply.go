package cli

import (
	"fmt"

	"github.com/GinForGit/cli-migration/internal/apply"
	"github.com/GinForGit/cli-migration/internal/manifest"
	"github.com/GinForGit/cli-migration/internal/platform"
	"github.com/GinForGit/cli-migration/internal/plan"
	"github.com/spf13/cobra"
)

func newApplyCommand() *cobra.Command {
	var manifestPath string
	var dryRun bool
	var skipManual bool

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
			p, err := plan.Generate(cmd.Context(), plat, m)
			if err != nil {
				return err
			}

			if dryRun {
				fmt.Println("Dry-run mode. Planned actions:")
				fmt.Println(plan.Format(p))
				return nil
			}

			return apply.Execute(cmd.Context(), plat, p, apply.Options{SkipManual: skipManual})
		},
	}

	cmd.Flags().StringVarP(&manifestPath, "manifest", "m", "", "清单文件路径（必填）")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "只预览，不执行")
	cmd.Flags().BoolVar(&skipManual, "skip-manual", false, "跳过无法自动安装的 manual 条目")
	_ = cmd.MarkFlagRequired("manifest")
	return cmd
}
