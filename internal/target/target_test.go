package target

import (
	"path/filepath"
	"testing"
)

func TestValid(t *testing.T) {
	if !Valid(Claude) || !Valid(OpenCode) {
		t.Error("expected claude and opencode to be valid")
	}
	if Valid("cursor") || Valid("") {
		t.Error("expected unknown targets to be invalid")
	}
}

func TestDir_ClaudeUser(t *testing.T) {
	t.Setenv("HOME", "/h")
	if got := Dir(Claude, "user"); got != filepath.Join("/h", ".claude") {
		t.Errorf("got %q", got)
	}
}

func TestDir_OpenCodeUser(t *testing.T) {
	t.Setenv("HOME", "/h")
	t.Setenv("XDG_CONFIG_HOME", "/h/.config")
	if got := Dir(OpenCode, "user"); got != filepath.Join("/h", ".config", "opencode") {
		t.Errorf("got %q", got)
	}
}

func TestDir_LocalScope(t *testing.T) {
	for _, tc := range []struct{ target, want string }{
		{Claude, ".claude"},
		{OpenCode, ".opencode"},
	} {
		got := Dir(tc.target, "local")
		if filepath.Base(got) != tc.want {
			t.Errorf("Dir(%q, local) = %q, want base %q", tc.target, got, tc.want)
		}
	}
}

func TestSkillInstallPath(t *testing.T) {
	got := SkillInstallPath("/cfg", "frontend")
	want := filepath.Join("/cfg", "skills", "frontend")
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
