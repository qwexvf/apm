package claude

import (
	"path/filepath"
	"testing"
)

func TestPluginInstallPath(t *testing.T) {
	got := PluginInstallPath("/c", "official", "figma", "v2.1.0")
	want := filepath.Join("/c", "plugins", "cache", "official", "figma", "v2.1.0")
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestSkillInstallPath(t *testing.T) {
	got := SkillInstallPath("/c", "frontend-design")
	want := filepath.Join("/c", "skills", "frontend-design")
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestSkillsDir(t *testing.T) {
	got := SkillsDir("/c")
	want := filepath.Join("/c", "skills")
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestCacheDir(t *testing.T) {
	got := CacheDir("/c")
	want := filepath.Join("/c", "plugins", "cache")
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestMarketplaceCacheDir(t *testing.T) {
	got := MarketplaceCacheDir("/c")
	want := filepath.Join("/c", "plugins", "marketplaces")
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestDir_UserScope(t *testing.T) {
	t.Setenv("HOME", "/h")
	got := Dir("user")
	if got != filepath.Join("/h", ".claude") {
		t.Errorf("got %q", got)
	}
}
