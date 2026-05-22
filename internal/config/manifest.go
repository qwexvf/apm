package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const ManifestFile = "apm.toml"

type MarketplaceSource struct {
	Source string `toml:"source"` // "github"
	Repo   string `toml:"repo"`   // "org/repo"
}

type Manifest struct {
	PluginManager PluginManagerConfig          `toml:"plugin_manager"`
	Plugins       map[string]string            `toml:"plugins"`
	Skills        map[string]string            `toml:"skills,omitempty"`
	Marketplaces  map[string]MarketplaceSource `toml:"marketplaces"`
}

type PluginManagerConfig struct {
	Scope string `toml:"scope"` // "user" or "local"
}

func NewManifest() *Manifest {
	return &Manifest{
		PluginManager: PluginManagerConfig{Scope: "user"},
		Plugins:       map[string]string{},
		Skills:        map[string]string{},
		Marketplaces:  map[string]MarketplaceSource{},
	}
}

// LoadManifest reads apm.toml from dir.
func LoadManifest(dir string) (*Manifest, error) {
	path := filepath.Join(dir, ManifestFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	m := NewManifest()
	if _, err := toml.Decode(string(data), m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return m, nil
}

// Save writes the manifest to dir/apm.toml atomically.
func (m *Manifest) Save(dir string) error {
	path := filepath.Join(dir, ManifestFile)
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if err := toml.NewEncoder(f).Encode(m); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}

// AddPlugin adds or updates a plugin entry in the manifest.
func (m *Manifest) AddPlugin(id, constraint string) {
	if m.Plugins == nil {
		m.Plugins = map[string]string{}
	}
	if constraint == "" {
		constraint = "*"
	}
	m.Plugins[id] = constraint
}

// RemovePlugin removes a plugin from the manifest.
func (m *Manifest) RemovePlugin(id string) bool {
	if _, ok := m.Plugins[id]; ok {
		delete(m.Plugins, id)
		return true
	}
	return false
}

// ParsePluginArg parses "name@marketplace[@constraint]" into parts.
// Returns (name, marketplace, constraint, error).
func ParsePluginArg(arg string) (name, marketplace, constraint string, err error) {
	parts := strings.SplitN(arg, "@", 3)
	switch len(parts) {
	case 1:
		return "", "", "", fmt.Errorf("missing marketplace: use name@marketplace format")
	case 2:
		if parts[0] == "" {
			return "", "", "", fmt.Errorf("missing plugin name in %q", arg)
		}
		if parts[1] == "" {
			return "", "", "", fmt.Errorf("missing marketplace in %q", arg)
		}
		return parts[0], parts[1], "*", nil
	case 3:
		if parts[0] == "" {
			return "", "", "", fmt.Errorf("missing plugin name in %q", arg)
		}
		if parts[1] == "" {
			return "", "", "", fmt.Errorf("missing marketplace in %q", arg)
		}
		return parts[0], parts[1], parts[2], nil
	}
	return "", "", "", fmt.Errorf("invalid plugin argument: %q", arg)
}

// PluginID returns "<name>@<marketplace>".
func PluginID(name, marketplace string) string {
	return name + "@" + marketplace
}

// SplitID splits "<name>@<marketplace>" into (name, marketplace).
func SplitID(id string) (name, marketplace string, err error) {
	name, marketplace, ok := strings.Cut(id, "@")
	if !ok {
		return "", "", fmt.Errorf("invalid plugin ID %q: missing @", id)
	}
	return name, marketplace, nil
}

// AddSkill adds or updates a skill entry in the manifest.
func (m *Manifest) AddSkill(id, constraint string) {
	if m.Skills == nil {
		m.Skills = map[string]string{}
	}
	if constraint == "" {
		constraint = "*"
	}
	m.Skills[id] = constraint
}

// RemoveSkill removes a skill from the manifest.
func (m *Manifest) RemoveSkill(id string) bool {
	if _, ok := m.Skills[id]; ok {
		delete(m.Skills, id)
		return true
	}
	return false
}

// ParseSkillArg parses "name@owner/repo[:subpath][@constraint]" into parts.
// Returns (name, repo, subpath, constraint, error). subpath may be "".
func ParseSkillArg(arg string) (name, repo, subpath, constraint string, err error) {
	parts := strings.SplitN(arg, "@", 3)
	switch len(parts) {
	case 1:
		return "", "", "", "", fmt.Errorf("missing source: use name@owner/repo[:subpath] format")
	case 2:
		name = parts[0]
		repo, subpath = splitRepoSubpath(parts[1])
		constraint = "*"
	case 3:
		name = parts[0]
		repo, subpath = splitRepoSubpath(parts[1])
		constraint = parts[2]
	}
	if name == "" {
		return "", "", "", "", fmt.Errorf("missing skill name in %q", arg)
	}
	if repo == "" {
		return "", "", "", "", fmt.Errorf("missing source in %q", arg)
	}
	if !strings.Contains(repo, "/") {
		return "", "", "", "", fmt.Errorf("source must be owner/repo, got %q", repo)
	}
	return name, repo, subpath, constraint, nil
}

func splitRepoSubpath(src string) (repo, subpath string) {
	repo, subpath, _ = strings.Cut(src, ":")
	return repo, subpath
}

// SkillID returns "<name>@<repo>" or "<name>@<repo>:<subpath>".
func SkillID(name, repo, subpath string) string {
	if subpath == "" {
		return name + "@" + repo
	}
	return name + "@" + repo + ":" + subpath
}

// SplitSkillID splits a skill ID back into (name, repo, subpath).
func SplitSkillID(id string) (name, repo, subpath string, err error) {
	name, src, ok := strings.Cut(id, "@")
	if !ok {
		return "", "", "", fmt.Errorf("invalid skill ID %q: missing @", id)
	}
	repo, subpath = splitRepoSubpath(src)
	if !strings.Contains(repo, "/") {
		return "", "", "", fmt.Errorf("invalid skill ID %q: source must be owner/repo", id)
	}
	return name, repo, subpath, nil
}
