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

var (
	updateDryRun bool
	updateYes    bool
)

var updateCmd = &cobra.Command{
	Use:   "update [name@marketplace]",
	Short: "Update plugins to latest version matching constraint",
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

		claudeDir := claude.Dir(m.PluginManager.Scope)
		gh := fetcher.NewGitHub()
		ctx := context.Background()

		// filter to requested plugins or all
		toUpdate := m.Plugins
		if len(args) > 0 {
			toUpdate = map[string]string{}
			for _, arg := range args {
				pluginName, marketplace, _, err := config.ParsePluginArg(arg)
				if err != nil {
					return err
				}
				id := config.PluginID(pluginName, marketplace)
				constraint, ok := m.Plugins[id]
				if !ok {
					return fmt.Errorf("%s not in manifest", id)
				}
				toUpdate[id] = constraint
			}
		}

		type update struct {
			id         string
			oldVersion string
			newVersion string
			commitSHA  string
			repo       string
		}

		km, err := claude.LoadKnownMarketplaces(claudeDir)
		if err != nil {
			return fmt.Errorf("load marketplaces: %w", err)
		}
		var updates []update

		for id, constraint := range toUpdate {
			_, marketplace, err := config.SplitID(id)
			if err != nil {
				return err
			}
			repo := km.Repo(marketplace)
			if repo == "" {
				if ms, ok := m.Marketplaces[marketplace]; ok {
					repo = ms.Repo
				}
			}
			if repo == "" {
				fmt.Printf("  skip %s: unknown marketplace\n", id)
				continue
			}

			res, err := resolver.Resolve(ctx, gh, repo, constraint)
			if err != nil {
				fmt.Printf("  skip %s: %v\n", id, err)
				continue
			}

			locked := lock.Get(id)
			oldVersion := "(not installed)"
			if locked != nil {
				oldVersion = locked.Version
			}

			if locked != nil && locked.CommitSHA == res.CommitSHA {
				fmt.Printf("  %s: already at %s\n", id, res.Version)
				continue
			}

			updates = append(updates, update{
				id:         id,
				oldVersion: oldVersion,
				newVersion: res.Version,
				commitSHA:  res.CommitSHA,
				repo:       repo,
			})
		}

		if len(updates) == 0 {
			fmt.Println("all plugins up to date")
			return nil
		}

		fmt.Println("\nupdates available:")
		for _, u := range updates {
			fmt.Printf("  %s  %s → %s\n", u.id, u.oldVersion, u.newVersion)
		}

		if updateDryRun {
			return nil
		}

		if !updateYes {
			fmt.Print("\napply updates? [y/N]: ")
			var yn string
			fmt.Scanln(&yn)
			if yn != "y" && yn != "Y" {
				fmt.Println("aborted")
				return nil
			}
		}

		reg, err := claude.LoadRegistry(claudeDir)
		if err != nil {
			return err
		}
		settings, err := claude.LoadSettings(claudeDir)
		if err != nil {
			return err
		}

		for _, u := range updates {
			pluginName, marketplace, err := config.SplitID(u.id)
			if err != nil {
				return err
			}

			result, err := installer.Install(ctx, gh, claudeDir, marketplace, pluginName, u.repo, u.newVersion)
			if err != nil {
				fmt.Printf("  ✗ %s: %v\n", u.id, err)
				continue
			}

			reg.Upsert(u.id, claude.InstallEntry{
				Scope:        m.PluginManager.Scope,
				InstallPath:  result.InstallPath,
				Version:      u.newVersion,
				GitCommitSHA: u.commitSHA,
				InstalledAt:  time.Now().UTC(),
			})
			settings.EnablePlugin(u.id, true)
			lock.Upsert(config.LockedPlugin{
				ID:          u.id,
				Version:     u.newVersion,
				CommitSHA:   u.commitSHA,
				ResolvedURL: "https://github.com/" + u.repo,
				InstallPath: result.InstallPath,
				Integrity:   result.Integrity,
			})
			fmt.Printf("  ✓ %s @ %s\n", u.id, u.newVersion)
		}

		if err := reg.Save(claudeDir); err != nil {
			return err
		}
		if err := settings.Save(claudeDir); err != nil {
			return err
		}
		return lock.Save(dir)
	},
}

func init() {
	updateCmd.Flags().BoolVar(&updateDryRun, "dry-run", false, "show updates without applying")
	updateCmd.Flags().BoolVarP(&updateYes, "yes", "y", false, "skip confirmation prompt")
}
