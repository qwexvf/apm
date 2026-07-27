// Package target resolves tool-specific config roots for apm's supported
// target tools (Claude Code, OpenCode).
package target

import (
	"fmt"
	"os"
	"path/filepath"
)

// Supported target tools.
const (
	Claude   = "claude"
	OpenCode = "opencode"
)

// Valid reports whether t names a supported target.
func Valid(t string) bool {
	return t == Claude || t == OpenCode
}

// Dir returns the tool's config root for the given target and scope.
//
//	claude:   user → ~/.claude,          local → <cwd>/.claude
//	opencode: user → ~/.config/opencode, local → <cwd>/.opencode (XDG_CONFIG_HOME respected)
func Dir(t, scope string) string {
	if scope == "local" {
		cwd, err := os.Getwd()
		if err != nil {
			cwd = "."
		}
		return filepath.Join(cwd, "."+t)
	}
	if t == OpenCode {
		if cfg, err := os.UserConfigDir(); err == nil && cfg != "" {
			return filepath.Join(cfg, "opencode")
		}
		return filepath.Join(home(), ".config", "opencode")
	}
	return filepath.Join(home(), ".claude")
}

// SkillsDir returns the skills root: <dir>/skills.
// The layout is identical for every supported target.
func SkillsDir(dir string) string {
	return filepath.Join(dir, "skills")
}

// SkillInstallPath returns the install path for a skill: <dir>/skills/<name>.
func SkillInstallPath(dir, name string) string {
	return filepath.Join(SkillsDir(dir), name)
}

func home() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	if home == "" {
		fmt.Fprintln(os.Stderr, "fatal: cannot determine home directory (HOME not set)")
		os.Exit(1)
	}
	return home
}
