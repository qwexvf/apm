package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/qwexvf/apm/internal/claude"
	"github.com/qwexvf/apm/internal/config"
	"github.com/qwexvf/apm/internal/fetcher"
	"github.com/qwexvf/apm/internal/resolver"
)

var lockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Regenerate apm.lock from the manifest constraints",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := manifestDir()
		m, err := config.LoadManifest(dir)
		if err != nil {
			return err
		}

		claudeDir := claude.Dir(m.PluginManager.Scope)
		km, err := claude.LoadKnownMarketplaces(claudeDir)
		if err != nil {
			return fmt.Errorf("load marketplaces: %w", err)
		}
		reg, err := claude.LoadRegistry(claudeDir)
		if err != nil {
			return fmt.Errorf("load registry: %w", err)
		}

		gh := fetcher.NewGitHub()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		lock := config.NewLock()
		scope := m.PluginManager.Scope

		for id, constraint := range m.Plugins {
			_, mktplace, err := config.SplitID(id)
			if err != nil {
				return err
			}
			repo := km.Repo(mktplace)
			if repo == "" {
				if ms, ok := m.Marketplaces[mktplace]; ok {
					repo = ms.Repo
				}
			}
			if repo == "" {
				fmt.Printf("  skip %s: unknown marketplace\n", id)
				continue
			}

			fmt.Printf("resolving %s @ %s...\n", id, constraint)
			res, err := resolver.Resolve(ctx, gh, repo, constraint)
			if err != nil {
				return err
			}

			// use existing install path for the manifest's scope
			var installPath string
			for _, e := range reg.Get(id) {
				if e.Scope == scope {
					installPath = e.InstallPath
					break
				}
			}

			lock.Upsert(config.LockedPlugin{
				ID:          id,
				Version:     res.Version,
				CommitSHA:   res.CommitSHA,
				ResolvedURL: "https://github.com/" + repo,
				InstallPath: installPath,
			})
			fmt.Printf("  locked %s @ %s\n", id, res.Version)
		}

		if err := lock.Save(dir); err != nil {
			return err
		}

		fmt.Printf("\n%s updated (%d plugins)\n", config.LockFile, len(lock.Plugins))
		return nil
	},
}
