package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/GinForGit/cli-migration/internal/discover"
	"github.com/GinForGit/cli-migration/internal/manifest"
	"github.com/GinForGit/cli-migration/internal/platform"
	"github.com/GinForGit/cli-migration/pkg/api"
	"github.com/spf13/cobra"
)

func newDiffCommand() *cobra.Command {
	var manifestPath string

	cmd := &cobra.Command{
		Use:   "diff",
		Short: "对比清单与当前环境",
		Long:  "显示清单中哪些条目与当前机器不一致。",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := manifest.Load(manifestPath)
			if err != nil {
				return err
			}
			plat, err := platform.New()
			if err != nil {
				return err
			}

			current, err := discover.CurrentEnvironment(cmd.Context(), false)
			if err != nil {
				return err
			}

			diff := computeDiff(m.Entries, current)
			fmt.Println(formatDiff(diff, api.OSType(plat.OS()), m.Source.OS))
			return nil
		},
	}

	cmd.Flags().StringVarP(&manifestPath, "manifest", "m", "", "清单文件路径（必填）")
	_ = cmd.MarkFlagRequired("manifest")
	return cmd
}

type diffEntry struct {
	Command       string
	Manifest      *api.Entry
	Current       *api.Entry
	ManifestOnly  bool
	CurrentOnly   bool
	VersionDiffer bool
}

func computeDiff(manifestEntries, currentEntries []api.Entry) []diffEntry {
	byCommand := make(map[string]*api.Entry)
	for i := range currentEntries {
		e := currentEntries[i]
		byCommand[e.Command] = &e
	}

	var result []diffEntry
	seen := make(map[string]bool)
	for i := range manifestEntries {
		m := manifestEntries[i]
		seen[m.Command] = true
		cur, ok := byCommand[m.Command]
		if !ok {
			result = append(result, diffEntry{Command: m.Command, Manifest: &m, ManifestOnly: true})
			continue
		}
		if m.Version != cur.Version && m.Version != "unknown" && cur.Version != "unknown" {
			result = append(result, diffEntry{
				Command:       m.Command,
				Manifest:      &m,
				Current:       cur,
				VersionDiffer: true,
			})
		}
	}

	for i := range currentEntries {
		c := currentEntries[i]
		if !seen[c.Command] {
			result = append(result, diffEntry{Command: c.Command, Current: &c, CurrentOnly: true})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Command < result[j].Command
	})
	return result
}

func formatDiff(diff []diffEntry, targetOS, sourceOS api.OSType) string {
	if len(diff) == 0 {
		return "No differences found."
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Diff against current environment (%s, source was %s):\n\n", targetOS, sourceOS))
	for _, d := range diff {
		switch {
		case d.ManifestOnly:
			b.WriteString(fmt.Sprintf("  [-] %s (missing on this machine)\n", d.Command))
		case d.CurrentOnly:
			b.WriteString(fmt.Sprintf("  [+] %s (not in manifest)\n", d.Command))
		case d.VersionDiffer:
			b.WriteString(fmt.Sprintf("  [~] %s: manifest %s vs current %s\n", d.Command, d.Manifest.Version, d.Current.Version))
		}
	}
	return b.String()
}
