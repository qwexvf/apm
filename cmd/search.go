package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/qwexvf/apm/internal/claude"
	"github.com/qwexvf/apm/internal/config"
	"github.com/qwexvf/apm/internal/marketplace"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search available plugins across marketplaces",
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")

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

		// merge manifest marketplaces
		for id, ms := range m.Marketplaces {
			if _, exists := km[id]; !exists {
				km.Add(id, ms.Repo, "")
			}
		}

		if len(km) == 0 {
			fmt.Println("no marketplaces registered — run: apm marketplace add")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "PLUGIN ID\tDESCRIPTION\tAUTHOR")

		mktplaceDir := claude.MarketplaceCacheDir(claudeDir)
		for id, rec := range km {
			localPath := rec.InstallLocation
			if localPath == "" {
				localPath = filepath.Join(mktplaceDir, id)
			}

			idx := marketplace.New(id, rec.Source.Repo, localPath)
			plugins, err := idx.ListPlugins()
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not read marketplace %q (try: apm marketplace update)\n", id)
				continue
			}

			for _, p := range plugins {
				if query != "" && !strings.Contains(strings.ToLower(p.Name+" "+p.Description), strings.ToLower(query)) {
					continue
				}
				desc := p.Description
				if runes := []rune(desc); len(runes) > 60 {
					desc = string(runes[:57]) + "..."
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", p.ID, desc, p.Author)
			}
		}

		return w.Flush()
	},
}
