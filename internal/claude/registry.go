package claude

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const registryFile = "plugins/installed_plugins.json"

// Registry represents installed_plugins.json v2.
type Registry struct {
	Version int                      `json:"version"`
	Plugins map[string][]InstallEntry `json:"plugins"`
}

// InstallEntry is one installation record for a plugin (one per scope).
type InstallEntry struct {
	Scope       string    `json:"scope"`
	ProjectPath string    `json:"projectPath,omitempty"`
	InstallPath string    `json:"installPath"`
	Version     string    `json:"version"`
	GitCommitSHA string   `json:"gitCommitSha,omitempty"`
	InstalledAt time.Time `json:"installedAt,omitempty"`
}

// LoadRegistry reads installed_plugins.json from claudeDir.
func LoadRegistry(claudeDir string) (*Registry, error) {
	path := filepath.Join(claudeDir, registryFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Registry{Version: 2, Plugins: map[string][]InstallEntry{}}, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	r := &Registry{}
	if err := json.Unmarshal(data, r); err != nil {
		return nil, fmt.Errorf("parse installed_plugins.json: %w", err)
	}
	if r.Plugins == nil {
		r.Plugins = map[string][]InstallEntry{}
	}
	return r, nil
}

// Save writes installed_plugins.json atomically.
func (r *Registry) Save(claudeDir string) error {
	path := filepath.Join(claudeDir, registryFile)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Upsert adds or replaces the user-scope entry for a plugin.
func (r *Registry) Upsert(id string, e InstallEntry) {
	entries := r.Plugins[id]
	for i, existing := range entries {
		if existing.Scope == e.Scope && existing.ProjectPath == e.ProjectPath {
			entries[i] = e
			r.Plugins[id] = entries
			return
		}
	}
	r.Plugins[id] = append(entries, e)
}

// Remove deletes all entries for a plugin ID (or just the scoped one).
func (r *Registry) Remove(id, scope string) {
	entries := r.Plugins[id]
	kept := entries[:0]
	for _, e := range entries {
		if e.Scope != scope {
			kept = append(kept, e)
		}
	}
	if len(kept) == 0 {
		delete(r.Plugins, id)
	} else {
		r.Plugins[id] = kept
	}
}

// Get returns entries for a plugin ID.
func (r *Registry) Get(id string) []InstallEntry {
	return r.Plugins[id]
}

// UserEntry returns the user-scope entry for a plugin, or nil.
func (r *Registry) UserEntry(id string) *InstallEntry {
	for i, e := range r.Plugins[id] {
		if e.Scope == "user" {
			return &r.Plugins[id][i]
		}
	}
	return nil
}
