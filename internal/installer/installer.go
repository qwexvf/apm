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

// ExtractComponents scans a plugin cache directory for skills (skills/*/SKILL.md)
// and agents (agents/*.md) and copies them into the target tool's config layout.
// Returns the list of destination paths that were created.
func ExtractComponents(cacheDir, targetDir string) ([]string, error) {
	var extracted []string

	skillsSrc := filepath.Join(cacheDir, "skills")
	if entries, err := os.ReadDir(skillsSrc); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			src := filepath.Join(skillsSrc, e.Name())
			if _, err := os.Stat(filepath.Join(src, "SKILL.md")); err != nil {
				continue
			}
			dst := target.SkillInstallPath(targetDir, e.Name())
			if err := os.RemoveAll(dst); err != nil {
				return extracted, err
			}
			if err := copyDir(src, dst); err != nil {
				return extracted, err
			}
			extracted = append(extracted, dst)
		}
	}

	agentsSrc := filepath.Join(cacheDir, "agents")
	if entries, err := os.ReadDir(agentsSrc); err == nil {
		agentsDst := filepath.Join(targetDir, "agent")
		if err := os.MkdirAll(agentsDst, 0755); err != nil {
			return extracted, err
		}
		for _, e := range entries {
			if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
				continue
			}
			src := filepath.Join(agentsSrc, e.Name())
			dst := filepath.Join(agentsDst, e.Name())
			if err := copyFile(src, dst); err != nil {
				return extracted, err
			}
			extracted = append(extracted, dst)
		}
	}

	return extracted, nil
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		sp := filepath.Join(src, e.Name())
		dp := filepath.Join(dst, e.Name())
		if e.IsDir() {
			if err := copyDir(sp, dp); err != nil {
				return err
			}
		} else {
			if err := copyFile(sp, dp); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// Uninstall removes a directory tree (plugin cache dir or extracted component).
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
