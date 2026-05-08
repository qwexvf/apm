package marketplace

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PluginMeta is the minimal plugin.json content under .claude-plugin/.
type PluginMeta struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Version     string `json:"version"`
}

// ListEntry is one plugin available in a marketplace.
type ListEntry struct {
	Name        string
	ID          string // "name@marketplace"
	Description string
	Author      string
	Version     string
}

// UpdateResult is returned by EnsureCloned with change details.
type UpdateResult struct {
	LocalPath string
	OldSHA    string // empty if freshly cloned
	NewSHA    string
	Cloned    bool // true if this was a fresh clone vs pull
}

// Index manages a local git clone of a marketplace repo.
type Index struct {
	ID        string // e.g. "claude-plugins-official"
	Repo      string // e.g. "anthropics/claude-plugins-official"
	LocalPath string // e.g. ~/.claude/plugins/marketplaces/claude-plugins-official
}

// New creates an Index for a marketplace.
func New(id, repo, localPath string) *Index {
	return &Index{ID: id, Repo: repo, LocalPath: localPath}
}

// EnsureCloned clones the marketplace repo if not present, or fast-forwards.
// Returns an UpdateResult with before/after SHAs and the local path.
func (idx *Index) EnsureCloned(ctx context.Context) (*UpdateResult, error) {
	res := &UpdateResult{LocalPath: idx.LocalPath}

	if _, err := os.Stat(filepath.Join(idx.LocalPath, ".git")); err == nil {
		res.OldSHA = idx.headSHA(ctx)
		if err := idx.pull(ctx); err != nil {
			return nil, err
		}
		res.NewSHA = idx.headSHA(ctx)
		return res, nil
	}

	// directory exists but has no .git — stale or partial; remove and re-clone
	if _, err := os.Stat(idx.LocalPath); err == nil {
		if err := os.RemoveAll(idx.LocalPath); err != nil {
			return nil, fmt.Errorf("remove stale dir %s: %w", idx.LocalPath, err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(idx.LocalPath), 0755); err != nil {
		return nil, err
	}
	url := "https://github.com/" + idx.Repo + ".git"
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth=1", url, idx.LocalPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("clone %s: %s", url, strings.TrimSpace(string(out)))
	}
	res.Cloned = true
	res.NewSHA = idx.headSHA(ctx)
	return res, nil
}

func (idx *Index) pull(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "git", "-C", idx.LocalPath, "pull", "--ff-only")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pull %s: %s", idx.Repo, strings.TrimSpace(string(out)))
	}
	return nil
}

func (idx *Index) headSHA(ctx context.Context) string {
	out, err := exec.CommandContext(ctx, "git", "-C", idx.LocalPath, "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// ListPlugins returns all plugins in the marketplace.
// Looks for directories containing .claude-plugin/plugin.json.
func (idx *Index) ListPlugins() ([]ListEntry, error) {
	pluginsDir := filepath.Join(idx.LocalPath, "plugins")
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		// marketplace might have plugins at root level
		pluginsDir = idx.LocalPath
		entries, err = os.ReadDir(pluginsDir)
		if err != nil {
			return nil, fmt.Errorf("read marketplace dir: %w", err)
		}
	}

	var result []ListEntry
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		meta, err := readPluginMeta(filepath.Join(pluginsDir, e.Name()))
		if err != nil {
			continue // skip dirs without plugin.json
		}
		name := e.Name()
		if meta.Name != "" {
			name = meta.Name
		}
		result = append(result, ListEntry{
			Name:        name,
			ID:          name + "@" + idx.ID,
			Description: meta.Description,
			Author:      meta.Author,
			Version:     meta.Version,
		})
	}
	return result, nil
}

// readPluginMeta reads .claude-plugin/plugin.json from a plugin directory.
func readPluginMeta(dir string) (*PluginMeta, error) {
	path := filepath.Join(dir, ".claude-plugin", "plugin.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	m := &PluginMeta{}
	if err := json.Unmarshal(data, m); err != nil {
		return nil, err
	}
	return m, nil
}

// PluginDir returns the directory for a named plugin within the marketplace clone.
func (idx *Index) PluginDir(name string) (string, error) {
	// try plugins/<name> first
	d := filepath.Join(idx.LocalPath, "plugins", name)
	if _, err := os.Stat(d); err == nil {
		return d, nil
	}
	// fallback: root level
	d = filepath.Join(idx.LocalPath, name)
	if _, err := os.Stat(d); err == nil {
		return d, nil
	}
	return "", fmt.Errorf("plugin %q not found in marketplace %q", name, idx.ID)
}
