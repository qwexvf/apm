package installer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/qwexvf/apm/internal/claude"
	"github.com/qwexvf/apm/internal/target"
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

// InstallSkill downloads a skill from owner/repo at ref, optionally selects a
// subpath within the extracted tree, and places it at <toolDir>/skills/<name>.
// toolDir is the target tool's config root (e.g. ~/.claude, ~/.config/opencode).
// subpath is a forward-slash relative path inside the repo (e.g. "skills/foo").
// The destination dir must contain a SKILL.md after extraction.
func InstallSkill(
	ctx context.Context,
	gh Downloader,
	toolDir, skillName, repo, ref, subpath string,
) (*Result, error) {
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repo: %q", repo)
	}
	owner, repoName := parts[0], parts[1]

	if subpath != "" {
		clean := filepath.ToSlash(filepath.Clean(subpath))
		if strings.HasPrefix(clean, "..") || strings.Contains(clean, "/../") || filepath.IsAbs(clean) {
			return nil, fmt.Errorf("invalid subpath %q: must be a relative path without ..", subpath)
		}
		subpath = clean
	}

	finalPath := target.SkillInstallPath(toolDir, skillName)

	skillsRoot := target.SkillsDir(toolDir)
	if err := os.MkdirAll(skillsRoot, 0755); err != nil {
		return nil, fmt.Errorf("mkdir skills root: %w", err)
	}
	stagingRoot := filepath.Join(skillsRoot, ".staging")
	if err := os.MkdirAll(stagingRoot, 0755); err != nil {
		return nil, fmt.Errorf("mkdir staging: %w", err)
	}

	staging, err := os.MkdirTemp(stagingRoot, skillName+"-")
	if err != nil {
		return nil, fmt.Errorf("create staging dir: %w", err)
	}
	defer os.RemoveAll(staging)

	fmt.Printf("  downloading %s/%s @ %s...\n", owner, repoName, ref)
	integrity, err := gh.DownloadTarball(ctx, owner, repoName, ref, staging)
	if err != nil {
		return nil, fmt.Errorf("install skill %s@%s: %w", skillName, ref, err)
	}

	src := staging
	if subpath != "" {
		src = filepath.Join(staging, filepath.FromSlash(subpath))
		info, err := os.Stat(src)
		if err != nil {
			return nil, fmt.Errorf("subpath %q not found in %s/%s@%s", subpath, owner, repoName, ref)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("subpath %q is not a directory", subpath)
		}
	}

	if _, err := os.Stat(filepath.Join(src, "SKILL.md")); err != nil {
		return nil, fmt.Errorf("no SKILL.md at %s (set subpath to the skill's directory)", subpath)
	}

	if err := os.RemoveAll(finalPath); err != nil {
		return nil, fmt.Errorf("remove existing skill dir: %w", err)
	}
	if err := os.Rename(src, finalPath); err != nil {
		return nil, fmt.Errorf("move skill into place: %w", err)
	}

	return &Result{InstallPath: finalPath, Integrity: integrity}, nil
}
