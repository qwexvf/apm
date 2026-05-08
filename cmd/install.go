package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/qwexvf/apm/internal/claude"
	"github.com/qwexvf/apm/internal/config"
	"github.com/qwexvf/apm/internal/fetcher"
	"github.com/qwexvf/apm/internal/installer"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install all plugins from apm.lock (deterministic)",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := manifestDir()

		m, err := config.LoadManifest(dir)
		if err != nil {
			return fmt.Errorf("no manifest found — run: apm init")
		}

		lock, err := config.LoadLock(dir)
		if err != nil {
			return err
		}

		if len(lock.Plugins) == 0 {
			fmt.Println("lockfile is empty — run: apm add <plugin>")
			return nil
		}

		claudeDir := claude.Dir(m.PluginManager.Scope)
		gh := fetcher.NewGitHub()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		reg, err := claude.LoadRegistry(claudeDir)
		if err != nil {
			return err
		}
		settings, err := claude.LoadSettings(claudeDir)
		if err != nil {
			return err
		}

		// load marketplaces once, not per plugin
		km, err := claude.LoadKnownMarketplaces(claudeDir)
		if err != nil {
			return fmt.Errorf("load marketplaces: %w", err)
		}

		lockDirty := false
		var failed []string
		for i, locked := range lock.Plugins {
			pluginName, marketplace, err := config.SplitID(locked.ID)
			if err != nil {
				failed = append(failed, locked.ID+": "+err.Error())
				continue
			}

			repo := km.Repo(marketplace)
			if repo == "" {
				if ms, ok := m.Marketplaces[marketplace]; ok {
					repo = ms.Repo
				}
			}
			if repo == "" {
				failed = append(failed, locked.ID+": unknown marketplace "+marketplace)
				continue
			}

			result, err := installer.Install(ctx, gh, claudeDir, marketplace, pluginName, repo, locked.Version)
			if err != nil {
				failed = append(failed, locked.ID+": "+err.Error())
				continue
			}

			// patch install path back into the lock if it was empty (e.g. written by apm lock)
			if lock.Plugins[i].InstallPath != result.InstallPath {
				lock.Plugins[i].InstallPath = result.InstallPath
				lockDirty = true
			}

			reg.Upsert(locked.ID, claude.InstallEntry{
				Scope:        m.PluginManager.Scope,
				InstallPath:  result.InstallPath,
				Version:      locked.Version,
				GitCommitSHA: locked.CommitSHA,
				InstalledAt:  time.Now().UTC(),
			})
			settings.EnablePlugin(locked.ID, true)
			if result.Integrity == "" {
				fmt.Printf("✓ %s @ %s  (cached)\n", locked.ID, locked.Version)
			} else {
				fmt.Printf("✓ %s @ %s\n", locked.ID, locked.Version)
			}
		}

		if lockDirty {
			if err := lock.Save(dir); err != nil {
				return err
			}
		}
		if err := reg.Save(claudeDir); err != nil {
			return err
		}
		if err := settings.Save(claudeDir); err != nil {
			return err
		}

		if len(failed) > 0 {
			fmt.Println("\nfailed:")
			for _, f := range failed {
				fmt.Println("  ✗", f)
			}
			return fmt.Errorf("%d plugin(s) failed to install", len(failed))
		}

		fmt.Printf("\n%d plugin(s) installed\n", len(lock.Plugins)-len(failed))
		return nil
	},
}
