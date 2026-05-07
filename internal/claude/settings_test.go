package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSettings_preservesUnknownKeys(t *testing.T) {
	dir := t.TempDir()
	raw := `{
  "model": "sonnet",
  "theme": "dark",
  "extraKnownMarketplaces": {"caveman": {"source": "github"}},
  "enabledPlugins": {"figma@claude-plugins-official": true}
}`
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(raw), 0644); err != nil {
		t.Fatal(err)
	}

	s, err := LoadSettings(dir)
	if err != nil {
		t.Fatalf("LoadSettings: %v", err)
	}

	if !s.EnabledPlugins["figma@claude-plugins-official"] {
		t.Error("expected figma plugin enabled")
	}

	// Save and verify unknown keys round-trip.
	if err := s.Save(dir); err != nil {
		t.Fatalf("Save: %v", err)
	}

	out := map[string]json.RawMessage{}
	data, _ := os.ReadFile(filepath.Join(dir, "settings.json"))
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	for _, key := range []string{"model", "theme", "extraKnownMarketplaces"} {
		if _, ok := out[key]; !ok {
			t.Errorf("key %q lost after save", key)
		}
	}
}

func TestLoadSettings_missingFile(t *testing.T) {
	s, err := LoadSettings(t.TempDir())
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if s.EnabledPlugins == nil {
		t.Error("EnabledPlugins should be initialised")
	}
}

func TestSettings_enableDisable(t *testing.T) {
	s := &Settings{EnabledPlugins: map[string]bool{}}
	s.EnablePlugin("caveman@caveman", true)
	if !s.EnabledPlugins["caveman@caveman"] {
		t.Error("plugin should be enabled")
	}
	s.DisablePlugin("caveman@caveman")
	if _, ok := s.EnabledPlugins["caveman@caveman"]; ok {
		t.Error("plugin should be removed")
	}
}

// Regression: extraKnownMarketplaces as object must not crash LoadSettings.
func TestLoadSettings_extraKnownMarketplacesObject(t *testing.T) {
	dir := t.TempDir()
	// Claude Code stores this as an object, not an array.
	raw := `{"extraKnownMarketplaces": {"caveman": {"source": {"source": "github", "repo": "JuliusBrussee/caveman"}, "installLocation": "/home/user/.claude/plugins/marketplaces/caveman"}}}`
	os.WriteFile(filepath.Join(dir, "settings.json"), []byte(raw), 0644)

	if _, err := LoadSettings(dir); err != nil {
		t.Fatalf("should not fail on object-shaped extraKnownMarketplaces: %v", err)
	}
}
