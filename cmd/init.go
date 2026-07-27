package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/qwexvf/apm/internal/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Scaffold a apm.toml manifest",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := manifestDir()
		path := filepath.Join(dir, config.ManifestFile)

		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%s already exists", path)
		}

		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		m := config.NewManifest()
		m.PluginManager.Scope = resolveScope()
		t := resolveTarget(nil)
		if t != "claude" {
			m.PluginManager.Target = t
		}

		// seed the official marketplace (claude target only — marketplaces
		// are a Claude Code concept)
		if t == "claude" {
			m.Marketplaces["claude-plugins-official"] = config.MarketplaceSource{
				Source: "github",
				Repo:   "anthropics/claude-plugins-official",
			}
		}

		if err := m.Save(dir); err != nil {
			return err
		}

		fmt.Printf("created %s  (scope: %s, target: %s)\n", path, m.PluginManager.Scope, t)
		fmt.Println("\nNext steps:")
		fmt.Println("  apm add <plugin>@<marketplace>   # add a plugin")
		fmt.Println("  apm install                       # install from lockfile")
		return nil
	},
}
