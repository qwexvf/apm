package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/qwexvf/apm/internal/claude"
	"github.com/qwexvf/apm/internal/config"
	"github.com/qwexvf/apm/internal/target"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := manifestDir()
		m, err := config.LoadManifest(dir)
		if err != nil {
			return fmt.Errorf("no manifest — run: apm init")
		}

		toolDir := targetDir(m)

		var reg *claude.Registry
		if resolveTarget(m) == "claude" {
			var err error
			reg, err = claude.LoadRegistry(toolDir)
			if err != nil {
				return err
			}
		}

		lock, err := config.LoadLock(dir)
		if err != nil {
			return err
		}

		if len(m.Plugins) == 0 && len(m.Skills) == 0 {
			fmt.Println("nothing in manifest")
			return nil
		}

		green := color.New(color.FgGreen).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()
		orange := color.New(color.FgYellow, color.Bold).SprintFunc()
		dim := color.New(color.Faint).SprintFunc()

		home, _ := os.UserHomeDir()
		shorten := func(p string) string {
			if home == "" {
				return p
			}
			if rel, err := filepath.Rel(home, p); err == nil && len(rel) < len(p) {
				return "~/" + rel
			}
			return p
		}

		if len(m.Plugins) > 0 {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "PLUGIN\tVERSION\tCONSTRAINT\tSTATUS\tPATH")

			for id, constraint := range m.Plugins {
				var entries []claude.InstallEntry
				if reg != nil {
					entries = reg.Get(id)
				}
				locked := lock.Get(id)

				version := "-"
				status := yellow("not installed")
				path := dim("-")

				if len(entries) > 0 {
					version = entries[0].Version
					status = green("installed")
					path = shorten(entries[0].InstallPath)
				} else if locked != nil {
					version = locked.Version
					installPath := locked.InstallPath
					if installPath == "" {
						pluginName, mktplace, _ := config.SplitID(id)
						installPath = claude.PluginInstallPath(toolDir, mktplace, pluginName, locked.Version)
					}
					if _, err := os.Stat(installPath); err == nil {
						status = orange("on disk")
						path = shorten(installPath)
					} else {
						status = yellow("not installed")
						if locked.InstallPath != "" {
							path = dim(shorten(locked.InstallPath))
						}
					}
				}

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", id, version, constraint, status, path)
			}
			if err := w.Flush(); err != nil {
				return err
			}
		}

		if len(m.Skills) > 0 {
			if len(m.Plugins) > 0 {
				fmt.Println()
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "SKILL\tVERSION\tCONSTRAINT\tSTATUS\tPATH")

			for id, constraint := range m.Skills {
				locked := lock.GetSkill(id)
				skillName, _, _, _ := config.SplitSkillID(id)
				expected := target.SkillInstallPath(toolDir, skillName)

				version := "-"
				status := yellow("not installed")
				path := dim("-")

				if locked != nil {
					version = locked.Version
				}
				if _, err := os.Stat(filepath.Join(expected, "SKILL.md")); err == nil {
					status = green("installed")
					path = shorten(expected)
				} else if locked != nil && locked.InstallPath != "" {
					path = dim(shorten(locked.InstallPath))
				}

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", id, version, constraint, status, path)
			}
			return w.Flush()
		}

		return nil
	},
}
