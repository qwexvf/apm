package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const LockFile = "apm.lock"

type Lock struct {
	Plugins []LockedPlugin `toml:"plugin"`
	Skills  []LockedSkill  `toml:"skill,omitempty"`
}

type LockedPlugin struct {
	ID          string `toml:"id"`
	Version     string `toml:"version"`
	CommitSHA   string `toml:"commit_sha"`
	ResolvedURL string `toml:"resolved_url"`
	InstallPath string `toml:"install_path"`
	Integrity   string `toml:"integrity"` // "sha256:<hex>"
}

type LockedSkill struct {
	ID          string `toml:"id"` // "<name>@<owner/repo>[:subpath]"
	Version     string `toml:"version"`
	CommitSHA   string `toml:"commit_sha"`
	ResolvedURL string `toml:"resolved_url"`
	InstallPath string `toml:"install_path"`
	Integrity   string `toml:"integrity"`
}

func NewLock() *Lock {
	return &Lock{Plugins: []LockedPlugin{}, Skills: []LockedSkill{}}
}

// LoadLock reads apm.lock from dir.
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

// Save writes the lock to dir/apm.lock atomically.
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
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
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

// GetSkill returns the locked entry for a skill ID, or nil.
func (l *Lock) GetSkill(id string) *LockedSkill {
	for i := range l.Skills {
		if l.Skills[i].ID == id {
			return &l.Skills[i]
		}
	}
	return nil
}

// UpsertSkill adds or replaces the locked entry for a skill.
func (l *Lock) UpsertSkill(s LockedSkill) {
	for i := range l.Skills {
		if l.Skills[i].ID == s.ID {
			l.Skills[i] = s
			return
		}
	}
	l.Skills = append(l.Skills, s)
}

// RemoveSkill deletes the locked entry for a skill ID.
func (l *Lock) RemoveSkill(id string) bool {
	for i, s := range l.Skills {
		if s.ID == id {
			l.Skills = append(l.Skills[:i], l.Skills[i+1:]...)
			return true
		}
	}
	return false
}
