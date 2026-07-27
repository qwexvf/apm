package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/qwexvf/apm/internal/claude"
	"github.com/qwexvf/apm/internal/config"
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

		toolDir := targetDir(m)
		gh := newGH()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		// filter to requested plugins/skills or all
		toUpdate := m.Plugins
		toUpdateSkills := m.Skills
		if len(args) > 0 {
			toUpdate = map[string]string{}
			toUpdateSkills = map[string]string{}
			for _, arg := range args {
				if isSkillArg(arg) {
					skillName, repo, subpath, _, err := config.ParseSkillArg(arg)
					if err != nil {
						return err
					}
					id := config.SkillID(skillName, repo, subpath)
					constraint, ok := m.Skills[id]
					if !ok {
						return fmt.Errorf("%s not in manifest", id)
					}
					toUpdateSkills[id] = constraint
					continue
				}
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
			constraint string
		}

		isClaude := resolveTarget(m) == "claude"
		if len(toUpdate) > 0 && !isClaude {
			return fmt.Errorf("marketplace plugins are not supported for target %q (claude-only); skills work with any target", resolveTarget(m))
		}

		var km claude.KnownMarketplaces
		if isClaude {
			var err error
			km, err = claude.LoadKnownMarketplaces(toolDir)
			if err != nil {
				return fmt.Errorf("load marketplaces: %w", err)
			}
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
				constraint: constraint,
			})
		}

		type skillUpdate struct {
			id         string
			oldVersion string
			newVersion string
			commitSHA  string
			repo       string
			subpath    string
			skillName  string
			constraint string
		}
		var skillUpdates []skillUpdate

		for id, constraint := range toUpdateSkills {
			skillName, repo, subpath, err := config.SplitSkillID(id)
			if err != nil {
				return err
			}
			res, err := resolver.Resolve(ctx, gh, repo, constraint)
			if err != nil {
				fmt.Printf("  skip %s: %v\n", id, err)
				continue
			}
			locked := lock.GetSkill(id)
			oldVersion := "(not installed)"
			if locked != nil {
				oldVersion = locked.Version
				if locked.CommitSHA == res.CommitSHA {
					fmt.Printf("  %s: already at %s\n", id, res.Version)
					continue
				}
			}
			skillUpdates = append(skillUpdates, skillUpdate{
				id:         id,
				oldVersion: oldVersion,
				newVersion: res.Version,
				commitSHA:  res.CommitSHA,
				repo:       repo,
				subpath:    subpath,
				skillName:  skillName,
				constraint: constraint,
			})
		}

		if len(updates) == 0 && len(skillUpdates) == 0 {
			fmt.Println("all items up to date")
			return nil
		}

		if len(updates) > 0 {
			fmt.Println("\nplugin updates available:")
			for _, u := range updates {
				fmt.Printf("  %s  %s → %s  (%s)\n", u.id, u.oldVersion, u.newVersion, u.constraint)
			}
		}
		if len(skillUpdates) > 0 {
			fmt.Println("\nskill updates available:")
			for _, u := range skillUpdates {
				fmt.Printf("  %s  %s → %s  (%s)\n", u.id, u.oldVersion, u.newVersion, u.constraint)
			}
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

		var reg *claude.Registry
		var settings *claude.Settings
		if isClaude {
			var err error
			reg, err = claude.LoadRegistry(toolDir)
			if err != nil {
				return err
			}
			settings, err = claude.LoadSettings(toolDir)
			if err != nil {
				return err
			}
		}

		var failed []string
		for _, u := range updates {
			pluginName, marketplace, err := config.SplitID(u.id)
			if err != nil {
				return err
			}

			result, err := installer.Install(ctx, gh, toolDir, marketplace, pluginName, u.repo, u.newVersion)
			if err != nil {
				fmt.Printf("  ✗ %s: %v\n", u.id, err)
				failed = append(failed, u.id+": "+err.Error())
				continue
			}

			// preserve existing integrity if installer skipped a cached directory
			integrity := result.Integrity
			if integrity == "" {
				if existing := lock.Get(u.id); existing != nil {
					integrity = existing.Integrity
				}
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
				Integrity:   integrity,
			})
			fmt.Printf("  ✓ %s @ %s\n", u.id, u.newVersion)
		}

		for _, u := range skillUpdates {
			result, err := installer.InstallSkill(ctx, gh, toolDir, u.skillName, u.repo, u.newVersion, u.subpath)
			if err != nil {
				fmt.Printf("  ✗ %s: %v\n", u.id, err)
				failed = append(failed, u.id+": "+err.Error())
				continue
			}
			integrity := result.Integrity
			if integrity == "" {
				if existing := lock.GetSkill(u.id); existing != nil {
					integrity = existing.Integrity
				}
			}
			lock.UpsertSkill(config.LockedSkill{
				ID:          u.id,
				Version:     u.newVersion,
				CommitSHA:   u.commitSHA,
				ResolvedURL: "https://github.com/" + u.repo,
				InstallPath: result.InstallPath,
				Integrity:   integrity,
			})
			fmt.Printf("  ✓ skill %s @ %s\n", u.id, u.newVersion)
		}

		// lockfile first — runtime state can be rebuilt with apm sync
		if err := lock.Save(dir); err != nil {
			return err
		}
		if isClaude {
			if err := reg.Save(toolDir); err != nil {
				return err
			}
			if err := settings.Save(toolDir); err != nil {
				return err
			}
		}
		if len(failed) > 0 {
			fmt.Println("\nfailed:")
			for _, f := range failed {
				fmt.Println("  ✗", f)
			}
			return fmt.Errorf("%d item(s) failed to update", len(failed))
		}
		return nil
	},
}

func init() {
	updateCmd.Flags().BoolVar(&updateDryRun, "dry-run", false, "show updates without applying")
	updateCmd.Flags().BoolVarP(&updateYes, "yes", "y", false, "skip confirmation prompt")
}
