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
	"github.com/qwexvf/apm/internal/resolver"
)

var addCmd = &cobra.Command{
	Use:   "add <name@marketplace[@constraint]>",
	Short: "Add a plugin to the manifest and install it",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginName, marketplace, constraint, err := config.ParsePluginArg(args[0])
		if err != nil {
			return err
		}
		id := config.PluginID(pluginName, marketplace)

		dir := manifestDir()
		m, err := config.LoadManifest(dir)
		if err != nil {
			// no manifest yet — treat as init+add
			m = config.NewManifest()
			m.PluginManager.Scope = resolveScope()
		}

		lock, err := config.LoadLock(dir)
		if err != nil {
			return err
		}

		claudeDir := claude.Dir(m.PluginManager.Scope)

		// find marketplace repo
		km, err := claude.LoadKnownMarketplaces(claudeDir)
		if err != nil {
			return err
		}
		repo := km.Repo(marketplace)
		if repo == "" {
			// check manifest marketplaces
			if ms, ok := m.Marketplaces[marketplace]; ok {
				repo = ms.Repo
			}
		}
		if repo == "" {
			return fmt.Errorf("unknown marketplace %q — register it with: apm marketplace add %s github:org/repo", marketplace, marketplace)
		}

		ctx := context.Background()
		gh := fetcher.NewGitHub()

		fmt.Printf("resolving %s...\n", id)
		res, err := resolver.Resolve(ctx, gh, repo, constraint)
		if err != nil {
			return err
		}
		fmt.Printf("resolved: %s @ %s (%s)\n", id, res.Version, res.CommitSHA[:12])

		result, err := installer.Install(ctx, gh, claudeDir, marketplace, pluginName, repo, res.Version)
		if err != nil {
			return err
		}

		// update registry
		reg, err := claude.LoadRegistry(claudeDir)
		if err != nil {
			return err
		}
		reg.Upsert(id, claude.InstallEntry{
			Scope:        m.PluginManager.Scope,
			InstallPath:  result.InstallPath,
			Version:      res.Version,
			GitCommitSHA: res.CommitSHA,
			InstalledAt:  time.Now().UTC(),
		})
		if err := reg.Save(claudeDir); err != nil {
			return err
		}

		// update settings
		settings, err := claude.LoadSettings(claudeDir)
		if err != nil {
			return err
		}
		settings.EnablePlugin(id, true)
		if err := settings.Save(claudeDir); err != nil {
			return err
		}

		// update manifest
		m.AddPlugin(id, constraint)
		if err := m.Save(dir); err != nil {
			return err
		}

		// update lockfile
		lock.Upsert(config.LockedPlugin{
			ID:          id,
			Version:     res.Version,
			CommitSHA:   res.CommitSHA,
			ResolvedURL: "https://github.com/" + repo,
			InstallPath: result.InstallPath,
			Integrity:   result.Integrity,
		})
		if err := lock.Save(dir); err != nil {
			return err
		}

		fmt.Printf("✓ installed %s @ %s\n", id, res.Version)
		return nil
	},
}
