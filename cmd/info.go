package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/qwexvf/apm/internal/claude"
	"github.com/qwexvf/apm/internal/config"
	"github.com/qwexvf/apm/internal/marketplace"
	"github.com/qwexvf/apm/internal/target"
)

var infoCmd = &cobra.Command{
	Use:   "info <name@marketplace> | <name@owner/repo[:subpath]>",
	Short: "Show plugin or skill details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if isSkillArg(args[0]) {
			return runInfoSkill(args[0])
		}
		pluginName, mktplace, _, err := config.ParsePluginArg(args[0])
		if err != nil {
			return err
		}
		id := config.PluginID(pluginName, mktplace)

		dir := manifestDir()
		m, err := config.LoadManifest(dir)
		if err != nil {
			m = config.NewManifest()
			m.PluginManager.Scope = resolveScope()
		}

		if err := requireClaude(m, "marketplace plugins"); err != nil {
			return err
		}

		claudeDir := targetDir(m)
		km, err := claude.LoadKnownMarketplaces(claudeDir)
		if err != nil {
			return err
		}

		repo := km.Repo(mktplace)
		if repo == "" {
			if ms, ok := m.Marketplaces[mktplace]; ok {
				repo = ms.Repo
			}
		}

		mktplaceDir := claude.MarketplaceCacheDir(claudeDir)
		localPath := filepath.Join(mktplaceDir, mktplace)
		if rec, ok := km[mktplace]; ok && rec.InstallLocation != "" {
			localPath = rec.InstallLocation
		}

		idx := marketplace.New(mktplace, repo, localPath)
		plugins, err := idx.ListPlugins()
		if err != nil {
			return fmt.Errorf("read marketplace %q: %w", mktplace, err)
		}

		var entry *marketplace.ListEntry
		for i, p := range plugins {
			if p.ID == id {
				entry = &plugins[i]
				break
			}
		}
		if entry == nil {
			return fmt.Errorf("plugin %q not found in marketplace %q", pluginName, mktplace)
		}

		reg, err := claude.LoadRegistry(claudeDir)
		if err != nil {
			return err
		}
		lock, err := config.LoadLock(dir)
		if err != nil {
			return err
		}

		fmt.Printf("name:        %s\n", entry.Name)
		fmt.Printf("id:          %s\n", entry.ID)
		fmt.Printf("description: %s\n", entry.Description)
		fmt.Printf("author:      %s\n", entry.Author)
		fmt.Printf("marketplace: %s  (github.com/%s)\n", mktplace, repo)
		fmt.Printf("version:     %s\n", entry.Version)

		if locked := lock.Get(id); locked != nil {
			fmt.Printf("locked:      %s  (%s)\n", locked.Version, shortSHA(locked.CommitSHA))
		}
		if entries := reg.Get(id); len(entries) > 0 {
			fmt.Printf("installed:   %s\n", entries[0].Version)
			fmt.Printf("path:        %s\n", entries[0].InstallPath)
		}

		return nil
	},
}

func runInfoSkill(arg string) error {
	skillName, repo, subpath, _, err := config.ParseSkillArg(arg)
	if err != nil {
		return err
	}
	id := config.SkillID(skillName, repo, subpath)

	dir := manifestDir()
	m, err := config.LoadManifest(dir)
	if err != nil {
		m = config.NewManifest()
		m.PluginManager.Scope = resolveScope()
	}

	toolDir := targetDir(m)
	expected := target.SkillInstallPath(toolDir, skillName)

	fmt.Printf("name:        %s\n", skillName)
	fmt.Printf("id:          %s\n", id)
	fmt.Printf("source:      github.com/%s\n", repo)
	if subpath != "" {
		fmt.Printf("subpath:     %s\n", subpath)
	}
	if constraint, ok := m.Skills[id]; ok {
		fmt.Printf("constraint:  %s\n", constraint)
	}

	lock, err := config.LoadLock(dir)
	if err != nil {
		return err
	}
	if locked := lock.GetSkill(id); locked != nil {
		fmt.Printf("locked:      %s  (%s)\n", locked.Version, shortSHA(locked.CommitSHA))
	}
	if _, err := os.Stat(filepath.Join(expected, "SKILL.md")); err == nil {
		fmt.Printf("installed:   yes\n")
		fmt.Printf("path:        %s\n", expected)
	} else {
		fmt.Printf("installed:   no\n")
	}
	return nil
}
