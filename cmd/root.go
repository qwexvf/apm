package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/qwexvf/apm/internal/config"
	"github.com/qwexvf/apm/internal/target"
)

var (
	globalScope  bool
	localScope   bool
	targetFlag   string
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
	rootCmd.PersistentFlags().BoolVar(&globalScope, "global", false, "operate on user-scope (~/.claude/ or ~/.config/opencode/)")
	rootCmd.PersistentFlags().BoolVar(&localScope, "local", false, "operate on project-scope (.claude/ or .opencode/)")
	rootCmd.PersistentFlags().StringVar(&targetFlag, "target", "", "target tool: claude (default) or opencode")

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

// resolveTarget returns the effective target tool: the --target flag wins,
// then the manifest's target, then "claude".
func resolveTarget(m *config.Manifest) string {
	t := targetFlag
	if t == "" && m != nil {
		t = m.PluginManager.TargetOrDefault()
	}
	if t == "" {
		t = target.Claude
	}
	if !target.Valid(t) {
		fmt.Fprintf(os.Stderr, "error: invalid target %q (want %q or %q)\n", t, target.Claude, target.OpenCode)
		os.Exit(1)
	}
	return t
}

// targetDir returns the config root for the manifest's target and scope,
// e.g. ~/.claude, ./.claude, ~/.config/opencode, ./.opencode.
func targetDir(m *config.Manifest) string {
	return target.Dir(resolveTarget(m), m.PluginManager.Scope)
}

// requireClaude fails the command when the effective target is not claude.
// Marketplace plugins, registry, and settings are Claude Code concepts;
// only skills are portable to other targets.
func requireClaude(m *config.Manifest, feature string) error {
	if t := resolveTarget(m); t != target.Claude {
		return fmt.Errorf("%s: not supported for target %q (claude-only); skills work with any target", feature, t)
	}
	return nil
}

// manifestDir returns the directory where apm.toml lives based on scope.
// For user scope it follows the --target flag (each tool has its own
// manifest); without a flag it falls back to an existing opencode manifest
// when no claude manifest is present. Local scope is always the project root.
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
	t := targetFlag
	if t == "" {
		t = target.Claude
		claudeManifest := filepath.Join(target.Dir(target.Claude, scope), config.ManifestFile)
		opencodeManifest := filepath.Join(target.Dir(target.OpenCode, scope), config.ManifestFile)
		if _, err := os.Stat(claudeManifest); os.IsNotExist(err) {
			if _, err := os.Stat(opencodeManifest); err == nil {
				t = target.OpenCode
			}
		}
	}
	return target.Dir(t, scope)
}

// shortSHA returns the first 12 chars of a SHA, or the whole string if shorter.
func shortSHA(sha string) string {
	if len(sha) >= 12 {
		return sha[:12]
	}
	return sha
}
