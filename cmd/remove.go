package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/qwexvf/apm/internal/claude"
	"github.com/qwexvf/apm/internal/config"
	"github.com/qwexvf/apm/internal/installer"
)

var removeCmd = &cobra.Command{
	Use:   "remove <name@marketplace>",
	Short: "Remove a plugin from the manifest and uninstall it",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginName, marketplace, _, err := config.ParsePluginArg(args[0])
		if err != nil {
			return err
		}
		id := config.PluginID(pluginName, marketplace)

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

		locked := lock.Get(id)
		if locked != nil {
			if err := installer.Uninstall(locked.InstallPath); err != nil {
				return fmt.Errorf("uninstall: %w", err)
			}
		}

		reg, err := claude.LoadRegistry(claudeDir)
		if err != nil {
			return err
		}
		reg.Remove(id, m.PluginManager.Scope)
		if err := reg.Save(claudeDir); err != nil {
			return err
		}

		settings, err := claude.LoadSettings(claudeDir)
		if err != nil {
			return err
		}
		settings.DisablePlugin(id)
		if err := settings.Save(claudeDir); err != nil {
			return err
		}

		if !m.RemovePlugin(id) {
			fmt.Printf("warning: %s not in manifest\n", id)
		}
		if err := m.Save(dir); err != nil {
			return err
		}

		lock.Remove(id)
		if err := lock.Save(dir); err != nil {
			return err
		}

		fmt.Printf("✓ removed %s\n", id)
		return nil
	},
}
