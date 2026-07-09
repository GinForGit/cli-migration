package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/GinForGit/cli-migration/internal/bundle"
	"github.com/GinForGit/cli-migration/internal/manifest"
	"github.com/GinForGit/cli-migration/internal/platform"
	"github.com/spf13/cobra"
)

func newExportCommand() *cobra.Command {
	var manifestPath string
	var output string
	var includeConfigs bool

	cmd := &cobra.Command{
		Use:   "export",
		Short: "打包清单和配置文件",
		Long:  "将清单文件与可选的配置文件打包成 tar.gz 归档。",
		RunE: func(cmd *cobra.Command, args []string) error {
			plat, err := platform.New()
			if err != nil {
				return err
			}
			var configPaths []string
			if includeConfigs {
				configPaths = bundle.DefaultConfigPaths(plat.HomeDir())
			}
			return bundle.Pack(manifestPath, output, bundle.PackOptions{
				IncludeConfigs: includeConfigs,
				ConfigPaths:    configPaths,
			})
		},
	}

	cmd.Flags().StringVarP(&manifestPath, "manifest", "m", "", "清单文件路径（必填）")
	cmd.Flags().StringVarP(&output, "output", "o", "env.bundle.tar.gz", "输出 bundle 路径")
	cmd.Flags().BoolVar(&includeConfigs, "include-configs", false, "包含常用配置文件")
	_ = cmd.MarkFlagRequired("manifest")
	return cmd
}

func newImportCommand() *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "解包 bundle",
		Long:  "解压 bundle 归档并显示其中的清单路径。",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bundlePath := args[0]
			if outputDir == "" {
				base := filepath.Base(bundlePath)
				base = strings.TrimSuffix(base, filepath.Ext(base))
				base = strings.TrimSuffix(base, ".tar")
				outputDir = base
			}
			if err := os.MkdirAll(outputDir, 0o755); err != nil {
				return err
			}
			manifestPath, err := bundle.Unpack(bundlePath, outputDir)
			if err != nil {
				return err
			}
			if err := rewriteImportedConfigSources(manifestPath, outputDir); err != nil {
				return err
			}
			fmt.Printf("Imported to: %s\nManifest: %s\n", outputDir, manifestPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output-dir", "d", "", "解压目录（默认根据 bundle 名推断）")
	return cmd
}

// rewriteImportedConfigSources updates file ConfigRefs so they point to the
// configs/ directory inside the unpacked bundle instead of the original paths.
func rewriteImportedConfigSources(manifestPath, bundleDir string) error {
	m, err := manifest.Load(manifestPath)
	if err != nil {
		return err
	}
	configsDir := filepath.Join(bundleDir, "configs")
	rewritten := false
	for i := range m.Entries {
		for j := range m.Entries[i].ConfigRefs {
			ref := &m.Entries[i].ConfigRefs[j]
			if ref.Type != "file" || ref.Source == "" {
				continue
			}
			candidate := filepath.Join(configsDir, filepath.Base(ref.Source))
			if _, err := os.Stat(candidate); err == nil {
				ref.Source = candidate
				rewritten = true
			}
		}
	}
	if !rewritten {
		return nil
	}
	return manifest.Save(manifestPath, m)
}
