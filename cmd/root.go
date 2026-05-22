package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	globalScope  bool
	localScope   bool
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
	buildCommit = commit
	buildDate = date
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
		pruneCmd,
		marketplaceCmd,
		scaffoldCmd,
	)
}

// resolveScope returns "user" or "local" based on flags.
func resolveScope() string {
	if localScope && globalScope {
		fmt.Fprintln(os.Stderr, "error: --local and --global are mutually exclusive")
		os.Exit(1)
	}
	if localScope {
		return "local"
	}
	return "user"
}

// manifestDir returns the directory where apm.toml lives based on scope.
func manifestDir() string {
	scope := resolveScope()
	if scope == "local" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, "warning: cannot determine working directory:", err)
			cwd = "."
		}
		return cwd
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	if home == "" {
		fmt.Fprintln(os.Stderr, "fatal: cannot determine home directory (HOME not set)")
		os.Exit(1)
	}
	return filepath.Join(home, ".claude")
}

// shortSHA returns the first 12 chars of a SHA, or the whole string if shorter.
func shortSHA(sha string) string {
	if len(sha) >= 12 {
		return sha[:12]
	}
	return sha
}
