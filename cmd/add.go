package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/qwexvf/apm/internal/claude"
	"github.com/qwexvf/apm/internal/config"
	"github.com/qwexvf/apm/internal/installer"
	"github.com/qwexvf/apm/internal/resolver"
)

// isSkillArg returns true when the second @-segment looks like a git source
// (contains "/"), e.g. "name@owner/repo[:subpath]". Plugin marketplace IDs
// never contain "/", so this is unambiguous.
func isSkillArg(arg string) bool {
	parts := strings.SplitN(arg, "@", 3)
	if len(parts) < 2 {
		return false
	}
	src, _, _ := strings.Cut(parts[1], ":")
	return strings.Contains(src, "/")
}

var addCmd = &cobra.Command{
	Use:   "add <name@marketplace[@constraint]> | <name@owner/repo[:subpath][@constraint]>",
	Short: "Add a plugin or skill to the manifest and install it",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if isSkillArg(args[0]) {
			return runAddSkill(args[0])
		}
		return runAddPlugin(args[0])
	},
}

func runAddPlugin(arg string) error {
	pluginName, marketplace, constraint, err := config.ParsePluginArg(arg)
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
		if targetFlag != "" {
			m.PluginManager.Target = resolveTarget(nil)
		}
	}

	lock, err := config.LoadLock(dir)
	if err != nil {
		return err
	}

	tgt := resolveTarget(m)
	isOC := tgt != "claude"
	toolDir := targetDir(m)

	// find marketplace repo — for claude use the known_marketplaces.json
	// runtime state; for opencode use the manifest only
	var repo string
	if !isOC {
		km, err := claude.LoadKnownMarketplaces(toolDir)
		if err != nil {
			return err
		}
		repo = km.Repo(marketplace)
	}
	if repo == "" {
		if ms, ok := m.Marketplaces[marketplace]; ok {
			repo = ms.Repo
		}
	}
	if repo == "" {
		return fmt.Errorf("unknown marketplace %q — register it with: apm marketplace add %s github:org/repo", marketplace, marketplace)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	gh := newGH()

	fmt.Printf("resolving %s...\n", id)
	res, err := resolver.Resolve(ctx, gh, repo, constraint)
	if err != nil {
		return err
	}
	fmt.Printf("resolved: %s @ %s (%s)\n", id, res.Version, shortSHA(res.CommitSHA))

	result, err := installer.Install(ctx, gh, toolDir, marketplace, pluginName, repo, res.Version)
	if err != nil {
		return err
	}

	var extracted []string
	if isOC {
		extracted, err = installer.ExtractComponents(result.InstallPath, toolDir)
		if err != nil {
			return fmt.Errorf("extract components: %w", err)
		}
		for _, p := range extracted {
			fmt.Printf("  extracted %s\n", p)
		}
	}

	// preserve existing integrity if installer skipped a cached directory
	integrity := result.Integrity
	if integrity == "" {
		if existing := lock.Get(id); existing != nil {
			integrity = existing.Integrity
		}
	}

	// manifest and lockfile first — runtime state can be rebuilt with apm sync
	m.AddPlugin(id, constraint)
	if err := m.Save(dir); err != nil {
		return err
	}

	lock.Upsert(config.LockedPlugin{
		ID:          id,
		Version:     res.Version,
		CommitSHA:   res.CommitSHA,
		ResolvedURL: "https://github.com/" + repo,
		InstallPath: result.InstallPath,
		Integrity:   integrity,
		Extracted:   extracted,
	})
	if err := lock.Save(dir); err != nil {
		return err
	}

	// runtime state — can be rebuilt with apm sync
	if !isOC {
		reg, err := claude.LoadRegistry(toolDir)
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
		if err := reg.Save(toolDir); err != nil {
			return err
		}

		settings, err := claude.LoadSettings(toolDir)
		if err != nil {
			return err
		}
		settings.EnablePlugin(id, true)
		if err := settings.Save(toolDir); err != nil {
			return err
		}
	}

	fmt.Printf("✓ installed %s @ %s\n", id, res.Version)
	return nil
}

func runAddSkill(arg string) error {
	skillName, repo, subpath, constraint, err := config.ParseSkillArg(arg)
	if err != nil {
		return err
	}
	id := config.SkillID(skillName, repo, subpath)

	dir := manifestDir()
	m, err := config.LoadManifest(dir)
	if err != nil {
		m = config.NewManifest()
		m.PluginManager.Scope = resolveScope()
		if targetFlag != "" {
			m.PluginManager.Target = resolveTarget(nil)
		}
	}

	lock, err := config.LoadLock(dir)
	if err != nil {
		return err
	}

	toolDir := targetDir(m)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	gh := newGH()

	fmt.Printf("resolving %s...\n", id)
	res, err := resolver.Resolve(ctx, gh, repo, constraint)
	if err != nil {
		return err
	}
	fmt.Printf("resolved: %s @ %s (%s)\n", id, res.Version, shortSHA(res.CommitSHA))

	result, err := installer.InstallSkill(ctx, gh, toolDir, skillName, repo, res.Version, subpath)
	if err != nil {
		return err
	}

	integrity := result.Integrity
	if integrity == "" {
		if existing := lock.GetSkill(id); existing != nil {
			integrity = existing.Integrity
		}
	}

	m.AddSkill(id, constraint)
	if err := m.Save(dir); err != nil {
		return err
	}

	lock.UpsertSkill(config.LockedSkill{
		ID:          id,
		Version:     res.Version,
		CommitSHA:   res.CommitSHA,
		ResolvedURL: "https://github.com/" + repo,
		InstallPath: result.InstallPath,
		Integrity:   integrity,
	})
	if err := lock.Save(dir); err != nil {
		return err
	}

	fmt.Printf("✓ installed skill %s @ %s\n", id, res.Version)
	return nil
}
