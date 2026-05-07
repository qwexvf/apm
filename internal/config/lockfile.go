package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const LockFile = "ccpm.lock"

type Lock struct {
	Plugins []LockedPlugin `toml:"plugin"`
}

type LockedPlugin struct {
	ID          string `toml:"id"`
	Version     string `toml:"version"`
	CommitSHA   string `toml:"commit_sha"`
	ResolvedURL string `toml:"resolved_url"`
	InstallPath string `toml:"install_path"`
	Integrity   string `toml:"integrity"` // "sha256:<hex>"
}

func NewLock() *Lock {
	return &Lock{Plugins: []LockedPlugin{}}
}

// LoadLock reads ccpm.lock from dir.
func LoadLock(dir string) (*Lock, error) {
	path := filepath.Join(dir, LockFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewLock(), nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	l := NewLock()
	if _, err := toml.Decode(string(data), l); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return l, nil
}

// Save writes the lock to dir/ccpm.lock atomically.
func (l *Lock) Save(dir string) error {
	path := filepath.Join(dir, LockFile)
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if err := toml.NewEncoder(f).Encode(l); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	f.Close()
	return os.Rename(tmp, path)
}

// Get returns the locked entry for a plugin ID, or nil.
func (l *Lock) Get(id string) *LockedPlugin {
	for i := range l.Plugins {
		if l.Plugins[i].ID == id {
			return &l.Plugins[i]
		}
	}
	return nil
}

// Upsert adds or replaces the locked entry for a plugin.
func (l *Lock) Upsert(p LockedPlugin) {
	for i := range l.Plugins {
		if l.Plugins[i].ID == p.ID {
			l.Plugins[i] = p
			return
		}
	}
	l.Plugins = append(l.Plugins, p)
}

// Remove deletes the locked entry for a plugin ID.
func (l *Lock) Remove(id string) bool {
	for i, p := range l.Plugins {
		if p.ID == id {
			l.Plugins = append(l.Plugins[:i], l.Plugins[i+1:]...)
			return true
		}
	}
	return false
}
