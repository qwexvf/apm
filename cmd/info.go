package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/qwexvf/apm/internal/claude"
	"github.com/qwexvf/apm/internal/config"
	"github.com/qwexvf/apm/internal/marketplace"
)

var infoCmd = &cobra.Command{
	Use:   "info <name@marketplace>",
	Short: "Show plugin details and available versions",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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

		claudeDir := claude.Dir(m.PluginManager.Scope)
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
		localPath := mktplaceDir + "/" + mktplace
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
			fmt.Printf("locked:      %s  (%s)\n", locked.Version, locked.CommitSHA[:12])
		}
		if entries := reg.Get(id); len(entries) > 0 {
			fmt.Printf("installed:   %s\n", entries[0].Version)
			fmt.Printf("path:        %s\n", entries[0].InstallPath)
		}

		return nil
	},
}
