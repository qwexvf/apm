package claude

import (
	"os"
	"path/filepath"
)

// Dir returns the Claude config directory for the given scope.
// scope="user" → ~/.claude, scope="local" → ./.claude (relative to cwd).
func Dir(scope string) string {
	if scope == "local" {
		cwd, err := os.Getwd()
		if err != nil {
			cwd = "."
		}
		return filepath.Join(cwd, ".claude")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, ".claude")
}

// CacheDir returns the plugin cache root: <claudeDir>/plugins/cache.
func CacheDir(claudeDir string) string {
	return filepath.Join(claudeDir, "plugins", "cache")
}

// MarketplaceCacheDir returns the marketplace clone root.
func MarketplaceCacheDir(claudeDir string) string {
	return filepath.Join(claudeDir, "plugins", "marketplaces")
}

// PluginInstallPath returns the expected install path for a plugin.
// marketplaceID is the short ID (e.g. "claude-plugins-official").
// pluginName is the plugin's name (e.g. "figma").
// version is the resolved version or commit SHA.
func PluginInstallPath(claudeDir, marketplaceID, pluginName, version string) string {
	return filepath.Join(CacheDir(claudeDir), marketplaceID, pluginName, version)
}
