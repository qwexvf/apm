package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/qwexvf/apm/internal/claude"
	"github.com/qwexvf/apm/internal/config"
	"github.com/qwexvf/apm/internal/target"
)

// syncCmd repairs the target tool's state (registry + settings for claude)
// from apm.lock without downloading anything. Skills are verified on disk.
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Repair tool state from apm.lock (no downloads)",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := manifestDir()
		m, err := config.LoadManifest(dir)
		if err != nil {
			return err
		}

		lock, err := config.LoadLock(dir)
		if err != nil {
			return err
		}

		if len(lock.Plugins) == 0 && len(lock.Skills) == 0 {
			fmt.Println("lockfile is empty")
			return nil
		}

		toolDir := targetDir(m)

		if resolveTarget(m) == "claude" {
			reg, err := claude.LoadRegistry(toolDir)
			if err != nil {
				return err
			}
			settings, err := claude.LoadSettings(toolDir)
			if err != nil {
				return err
			}

			for _, locked := range lock.Plugins {
				if locked.InstallPath == "" {
					fmt.Printf("  warning: %s install path missing — run: apm install\n", locked.ID)
					continue
				}
				if _, err := os.Stat(locked.InstallPath); err != nil {
					fmt.Printf("  warning: %s install path missing — run: apm install\n", locked.ID)
					continue
				}

				reg.Upsert(locked.ID, claude.InstallEntry{
					Scope:        m.PluginManager.Scope,
					InstallPath:  locked.InstallPath,
					Version:      locked.Version,
					GitCommitSHA: locked.CommitSHA,
					InstalledAt:  time.Now().UTC(),
				})
				settings.EnablePlugin(locked.ID, true)
				fmt.Printf("  ✓ synced %s @ %s\n", locked.ID, locked.Version)
			}

			if err := reg.Save(toolDir); err != nil {
				return err
			}
			if err := settings.Save(toolDir); err != nil {
				return err
			}
		} else if len(lock.Plugins) > 0 {
			fmt.Println("  note: skipping plugins — not supported for target opencode")
		}

		// skills don't write to registry/settings — just verify presence
		for _, locked := range lock.Skills {
			skillName, _, _, err := config.SplitSkillID(locked.ID)
			if err != nil {
				fmt.Printf("  warning: %s: %v\n", locked.ID, err)
				continue
			}
			expected := target.SkillInstallPath(toolDir, skillName)
			if _, err := os.Stat(filepath.Join(expected, "SKILL.md")); err != nil {
				fmt.Printf("  warning: skill %s missing on disk — run: apm install\n", locked.ID)
				continue
			}
			fmt.Printf("  ✓ skill %s @ %s\n", locked.ID, locked.Version)
		}

		fmt.Println("\nstate synced from lockfile")
		return nil
	},
}
