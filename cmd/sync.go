package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/qwexvf/apm/internal/claude"
	"github.com/qwexvf/apm/internal/config"
)

// syncCmd repairs Claude Code's state (registry + settings) from apm.lock
// without downloading anything. Useful when Claude Code's files get out of sync.
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Repair Claude Code state from apm.lock (no downloads)",
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

		claudeDir := claude.Dir(m.PluginManager.Scope)
		reg, err := claude.LoadRegistry(claudeDir)
		if err != nil {
			return err
		}
		settings, err := claude.LoadSettings(claudeDir)
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

		if err := reg.Save(claudeDir); err != nil {
			return err
		}
		if err := settings.Save(claudeDir); err != nil {
			return err
		}

		// skills don't write to registry/settings — just verify presence
		for _, locked := range lock.Skills {
			skillName, _, _, err := config.SplitSkillID(locked.ID)
			if err != nil {
				fmt.Printf("  warning: %s: %v\n", locked.ID, err)
				continue
			}
			expected := claude.SkillInstallPath(claudeDir, skillName)
			if _, err := os.Stat(filepath.Join(expected, "SKILL.md")); err != nil {
				fmt.Printf("  warning: skill %s missing on disk — run: apm install\n", locked.ID)
				continue
			}
			fmt.Printf("  ✓ skill %s @ %s\n", locked.ID, locked.Version)
		}

		fmt.Println("\nClaude Code state synced from lockfile")
		return nil
	},
}
