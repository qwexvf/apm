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

		claudeDir := claude.Dir(m.PluginManager.Scope)
		reg, err := claude.LoadRegistry(claudeDir)
		if err != nil {
			return err
		}

		lock, err := config.LoadLock(dir)
		if err != nil {
			return err
		}

		if len(m.Plugins) == 0 {
			fmt.Println("no plugins in manifest")
			return nil
		}

		green := color.New(color.FgGreen).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()
		orange := color.New(color.FgYellow, color.Bold).SprintFunc()
		dim := color.New(color.Faint).SprintFunc()

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "PLUGIN\tVERSION\tCONSTRAINT\tSTATUS\tPATH")

		for id, constraint := range m.Plugins {
			entries := reg.Get(id)
			locked := lock.Get(id)

			version := "-"
			status := yellow("not installed")
			path := dim("-")

			if len(entries) > 0 {
				version = entries[0].Version
				status = green("installed")
				path = entries[0].InstallPath
			} else if locked != nil {
				version = locked.Version
				installPath := locked.InstallPath
				if installPath == "" {
					// reconstruct expected path from locked version
					pluginName, mktplace, _ := config.SplitID(id)
					installPath = claude.PluginInstallPath(claudeDir, mktplace, pluginName, locked.Version)
				}
				if _, err := os.Stat(installPath); err == nil {
					// files on disk but not tracked in registry — needs apm sync
					status = orange("on disk")
					path = installPath
				} else {
					status = yellow("not installed")
					if locked.InstallPath != "" {
						path = dim(locked.InstallPath)
					}
				}
			}

			// trim path to be relative to home for readability
			if home, err := os.UserHomeDir(); err == nil {
				if rel, err := filepath.Rel(home, path); err == nil && len(rel) < len(path) {
					path = "~/" + rel
				}
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", id, version, constraint, status, path)
		}

		return w.Flush()
	},
}
