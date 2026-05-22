package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/qwexvf/apm/internal/claude"
	"github.com/qwexvf/apm/internal/config"
)

var pruneForce bool

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove cached plugin versions not referenced by any lockfile",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := manifestDir()
		m, err := config.LoadManifest(dir)
		if err != nil {
			return fmt.Errorf("no manifest — run: apm init")
		}

		lock, err := config.LoadLock(dir)
		if err != nil {
			return err
		}

		claudeDir := claude.Dir(m.PluginManager.Scope)
		cacheDir := claude.CacheDir(claudeDir)

		// build set of referenced install paths
		referenced := make(map[string]bool, len(lock.Plugins))
		for _, p := range lock.Plugins {
			if p.InstallPath != "" {
				referenced[filepath.Clean(p.InstallPath)] = true
			}
		}

		// build set of skill names referenced in the lockfile
		referencedSkills := make(map[string]bool, len(lock.Skills))
		for _, s := range lock.Skills {
			name, _, _, err := config.SplitSkillID(s.ID)
			if err == nil {
				referencedSkills[name] = true
			}
		}

		// walk the cache and collect unreferenced version directories
		// structure: <cacheDir>/<marketplace>/<plugin>/<version>/
		var orphans []string
		entries, err := os.ReadDir(cacheDir)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("cache is empty")
				return nil
			}
			return err
		}
		for _, mktEntry := range entries {
			if !mktEntry.IsDir() {
				continue
			}
			mktDir := filepath.Join(cacheDir, mktEntry.Name())
			plugins, err := os.ReadDir(mktDir)
			if err != nil {
				continue
			}
			for _, pluginEntry := range plugins {
				if !pluginEntry.IsDir() {
					continue
				}
				pluginDir := filepath.Join(mktDir, pluginEntry.Name())
				versions, err := os.ReadDir(pluginDir)
				if err != nil {
					continue
				}
				for _, vEntry := range versions {
					if !vEntry.IsDir() {
						continue
					}
					vPath := filepath.Clean(filepath.Join(pluginDir, vEntry.Name()))
					if !referenced[vPath] {
						orphans = append(orphans, vPath)
					}
				}
			}
		}

		// orphan skills: dirs under <claudeDir>/skills/ not referenced by lock.
		// always includes the .staging dir if present.
		skillsRoot := claude.SkillsDir(claudeDir)
		if sentries, err := os.ReadDir(skillsRoot); err == nil {
			for _, e := range sentries {
				if !e.IsDir() {
					continue
				}
				name := e.Name()
				if name == ".staging" {
					orphans = append(orphans, filepath.Clean(filepath.Join(skillsRoot, name)))
					continue
				}
				if !referencedSkills[name] {
					orphans = append(orphans, filepath.Clean(filepath.Join(skillsRoot, name)))
				}
			}
		}

		if len(orphans) == 0 {
			fmt.Println("nothing to prune")
			return nil
		}

		fmt.Printf("%d orphaned dir(s):\n", len(orphans))
		var totalSize int64
		for _, p := range orphans {
			size := dirSize(p)
			totalSize += size
			fmt.Printf("  %s  (%s)\n", p, formatBytes(size))
		}
		fmt.Printf("total: %s\n", formatBytes(totalSize))

		if !pruneForce {
			fmt.Print("\nremove? [y/N]: ")
			var yn string
			fmt.Scanln(&yn)
			if yn != "y" && yn != "Y" {
				fmt.Println("aborted")
				return nil
			}
		}

		for _, p := range orphans {
			if err := os.RemoveAll(p); err != nil {
				fmt.Fprintf(os.Stderr, "  warning: could not remove %s: %v\n", p, err)
			} else {
				fmt.Printf("  removed %s\n", p)
			}
		}
		return nil
	},
}

func dirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func init() {
	pruneCmd.Flags().BoolVarP(&pruneForce, "force", "f", false, "remove without confirmation")
}
