package installer

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/qwexvf/apm/internal/claude"
)

// Downloader fetches and extracts plugin archives.
type Downloader interface {
	DownloadTarball(ctx context.Context, owner, repo, ref, destDir string) (string, error)
}

// Result is returned after a successful install.
type Result struct {
	InstallPath string
	Integrity   string
}

// Install downloads and extracts a plugin to the cache directory.
// If the installPath already exists and integrity matches the locked value, it is a no-op.
func Install(
	ctx context.Context,
	gh Downloader,
	claudeDir string,
	marketplaceID, pluginName, repo, ref string,
) (*Result, error) {
	// repo is "owner/repo", ref is the resolved tag or SHA
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repo: %q", repo)
	}
	owner, repoName := parts[0], parts[1]

	installPath := claude.PluginInstallPath(claudeDir, marketplaceID, pluginName, ref)

	// skip if already extracted
	if _, err := os.Stat(installPath); err == nil {
		return &Result{InstallPath: installPath, Integrity: ""}, nil
	}

	fmt.Printf("  downloading %s/%s @ %s...\n", owner, repoName, ref)

	integrity, err := gh.DownloadTarball(ctx, owner, repoName, ref, installPath)
	if err != nil {
		os.RemoveAll(installPath)
		return nil, fmt.Errorf("install %s@%s: %w", pluginName, ref, err)
	}

	return &Result{InstallPath: installPath, Integrity: integrity}, nil
}

// Uninstall removes the plugin install directory.
func Uninstall(installPath string) error {
	if installPath == "" {
		return nil
	}
	if _, err := os.Stat(installPath); os.IsNotExist(err) {
		return nil
	}
	return os.RemoveAll(installPath)
}
