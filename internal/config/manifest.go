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
	Marketplaces  map[string]MarketplaceSource `toml:"marketplaces"`
}

type PluginManagerConfig struct {
	Scope string `toml:"scope"` // "user" or "local"
}

func NewManifest() *Manifest {
	return &Manifest{
		PluginManager: PluginManagerConfig{Scope: "user"},
		Plugins:       map[string]string{},
		Marketplaces:  map[string]MarketplaceSource{},
	}
}

// LoadManifest reads ccpm.toml from dir.
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

// Save writes the manifest to dir/apm.toml.
func (m *Manifest) Save(dir string) error {
	path := filepath.Join(dir, ManifestFile)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := toml.NewEncoder(f).Encode(m); err != nil {
		f.Close()
		return err
	}
	return f.Close()
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
		return parts[0], parts[1], "*", nil
	case 3:
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
