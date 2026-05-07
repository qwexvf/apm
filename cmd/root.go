package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	globalScope bool
	localScope  bool
	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:     "apm",
	Short:   "Claude Code Plugin Manager",
	Long:    "apm manages Claude Code plugins with lockfile-based reproducible installs.",
	Version: buildVersion,
}

// SetVersion is called from main with ldflags-injected values.
func SetVersion(version, commit, date string) {
	buildVersion = version
	buildCommit  = commit
	buildDate    = date
	rootCmd.Version = version
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&globalScope, "global", false, "operate on user-scope (~/.claude/)")
	rootCmd.PersistentFlags().BoolVar(&localScope, "local", false, "operate on project-scope (.claude/)")

	rootCmd.AddCommand(
		initCmd,
		addCmd,
		removeCmd,
		installCmd,
		updateCmd,
		listCmd,
		searchCmd,
		infoCmd,
		lockCmd,
		syncCmd,
		marketplaceCmd,
	)
}

// resolveScope returns "user" or "local" based on flags and cwd.
func resolveScope() string {
	if localScope {
		return "local"
	}
	return "user"
}

// manifestDir returns the directory where apm.toml lives based on scope.
func manifestDir() string {
	scope := resolveScope()
	if scope == "local" {
		cwd, _ := os.Getwd()
		return cwd
	}
	home, _ := os.UserHomeDir()
	return home + "/.claude"
}
