package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/qwexvf/ccpm/internal/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Scaffold a ccpm.toml manifest",
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

		// seed the official marketplace
		m.Marketplaces["claude-plugins-official"] = config.MarketplaceSource{
			Source: "github",
			Repo:   "anthropics/claude-plugins-official",
		}

		if err := m.Save(dir); err != nil {
			return err
		}

		fmt.Printf("created %s\n", path)
		fmt.Println("\nNext steps:")
		fmt.Println("  ccpm add <plugin>@<marketplace>   # add a plugin")
		fmt.Println("  ccpm install                       # install from lockfile")
		return nil
	},
}
