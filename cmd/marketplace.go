package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/qwexvf/ccpm/internal/claude"
	"github.com/qwexvf/ccpm/internal/config"
	mktpkg "github.com/qwexvf/ccpm/internal/marketplace"
)

var marketplaceCmd = &cobra.Command{
	Use:   "marketplace",
	Short: "Manage plugin marketplaces",
}

var marketplaceAddCmd = &cobra.Command{
	Use:   "add <id> <github:owner/repo>",
	Short: "Register a new marketplace",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		rawSrc := args[1]

		// parse "github:owner/repo" or plain "owner/repo"
		repo := strings.TrimPrefix(rawSrc, "github:")
		if !strings.Contains(repo, "/") {
			return fmt.Errorf("invalid source %q: expected github:owner/repo", rawSrc)
		}

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

		installLocation := claude.MarketplaceCacheDir(claudeDir) + "/" + id
		km.Add(id, repo, installLocation)
		if err := km.Save(claudeDir); err != nil {
			return err
		}

		// also add to manifest
		m.Marketplaces[id] = config.MarketplaceSource{Source: "github", Repo: repo}
		if err := m.Save(dir); err != nil {
			return err
		}

		// clone it
		idx := mktpkg.New(id, repo, installLocation)
		fmt.Printf("cloning %s from github.com/%s...\n", id, repo)
		if err := idx.EnsureCloned(); err != nil {
			return err
		}

		fmt.Printf("✓ marketplace %q registered\n", id)
		return nil
	},
}

var marketplaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered marketplaces",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		if len(km) == 0 {
			fmt.Println("no marketplaces registered")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tREPO\tLOCATION")
		for id, rec := range km {
			fmt.Fprintf(w, "%s\tgithub.com/%s\t%s\n", id, rec.Source.Repo, rec.InstallLocation)
		}
		return w.Flush()
	},
}

var marketplaceUpdateCmd = &cobra.Command{
	Use:   "update [id]",
	Short: "Pull latest marketplace indexes",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		mktplaceDir := claude.MarketplaceCacheDir(claudeDir)
		for id, rec := range km {
			if len(args) > 0 && args[0] != id {
				continue
			}
			localPath := rec.InstallLocation
			if localPath == "" {
				localPath = mktplaceDir + "/" + id
			}
			idx := mktpkg.New(id, rec.Source.Repo, localPath)
			fmt.Printf("updating %s...\n", id)
			if err := idx.EnsureCloned(); err != nil {
				fmt.Printf("  ✗ %s: %v\n", id, err)
			} else {
				fmt.Printf("  ✓ %s updated\n", id)
			}
		}
		return nil
	},
}

func init() {
	marketplaceCmd.AddCommand(marketplaceAddCmd, marketplaceListCmd, marketplaceUpdateCmd)
}
