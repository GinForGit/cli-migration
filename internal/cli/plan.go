package cli

import (
	"fmt"

	"github.com/GinForGit/cli-migration/internal/manifest"
	"github.com/GinForGit/cli-migration/internal/platform"
	"github.com/GinForGit/cli-migration/internal/plan"
	"github.com/spf13/cobra"
)

func newPlanCommand() *cobra.Command {
	var manifestPath string

	cmd := &cobra.Command{
		Use:   "plan",
		Short: "预览还原计划",
		Long:  "对比清单与当前环境，输出将要执行的操作（只读）。",
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
			fmt.Println(plan.Format(p))
			return nil
		},
	}

	cmd.Flags().StringVarP(&manifestPath, "manifest", "m", "", "清单文件路径（必填）")
	_ = cmd.MarkFlagRequired("manifest")
	return cmd
}
