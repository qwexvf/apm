package claude

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKnownMarketplaces_AddRepoSaveLoad(t *testing.T) {
	dir := t.TempDir()
	km := KnownMarketplaces{}
	km.Add("official", "anthropics/claude-plugins-official", "/loc/official")

	if got := km.Repo("official"); got != "anthropics/claude-plugins-official" {
		t.Errorf("Repo = %q", got)
	}
	if km.Repo("missing") != "" {
		t.Error("Repo for missing id should be empty")
	}

	if err := km.Save(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "plugins", "known_marketplaces.json")); err != nil {
		t.Error("known_marketplaces.json not created")
	}

	loaded, err := LoadKnownMarketplaces(dir)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Repo("official") != "anthropics/claude-plugins-official" {
		t.Error("did not round-trip")
	}
	if loaded["official"].LastUpdated == "" {
		t.Error("LastUpdated should be populated on Save")
	}
}

func TestLoadKnownMarketplaces_MissingFile(t *testing.T) {
	km, err := LoadKnownMarketplaces(t.TempDir())
	if err != nil {
		t.Fatalf("expected empty map for missing file, got: %v", err)
	}
	if len(km) != 0 {
		t.Errorf("expected empty, got %d entries", len(km))
	}
}
