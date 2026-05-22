package claude

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRegistryUpsertGet(t *testing.T) {
	r := &Registry{Version: 2, Plugins: map[string][]InstallEntry{}}
	r.Upsert("figma@official", InstallEntry{
		Scope:       "user",
		InstallPath: "/p1",
		Version:     "2.1.0",
		InstalledAt: time.Now().UTC(),
	})
	if got := r.Get("figma@official"); len(got) != 1 || got[0].Version != "2.1.0" {
		t.Fatalf("Get after Upsert failed: %+v", got)
	}

	// upsert same scope = replace
	r.Upsert("figma@official", InstallEntry{Scope: "user", InstallPath: "/p2", Version: "2.2.0"})
	got := r.Get("figma@official")
	if len(got) != 1 || got[0].Version != "2.2.0" {
		t.Errorf("Upsert did not replace same-scope entry: %+v", got)
	}

	// upsert different scope = append
	r.Upsert("figma@official", InstallEntry{Scope: "local", InstallPath: "/p3", Version: "2.0.0"})
	got = r.Get("figma@official")
	if len(got) != 2 {
		t.Errorf("Upsert different scope did not append, got %d entries", len(got))
	}
}

func TestRegistryRemoveByScope(t *testing.T) {
	r := &Registry{Version: 2, Plugins: map[string][]InstallEntry{}}
	r.Upsert("p@m", InstallEntry{Scope: "user", InstallPath: "/u"})
	r.Upsert("p@m", InstallEntry{Scope: "local", InstallPath: "/l"})

	r.Remove("p@m", "user")
	got := r.Get("p@m")
	if len(got) != 1 || got[0].Scope != "local" {
		t.Errorf("remove user-scope left wrong state: %+v", got)
	}

	r.Remove("p@m", "local")
	if got := r.Get("p@m"); len(got) != 0 {
		t.Errorf("expected empty after removing last scope, got %+v", got)
	}
	if _, ok := r.Plugins["p@m"]; ok {
		t.Error("empty key should be deleted")
	}
}

func TestRegistrySaveLoad(t *testing.T) {
	dir := t.TempDir()
	r := &Registry{Version: 2, Plugins: map[string][]InstallEntry{}}
	r.Upsert("caveman@caveman", InstallEntry{
		Scope:        "user",
		InstallPath:  "/cache/caveman/main",
		Version:      "main",
		GitCommitSHA: "ef6050c5",
	})
	if err := r.Save(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "plugins", "installed_plugins.json")); err != nil {
		t.Error("registry file not created")
	}

	r2, err := LoadRegistry(dir)
	if err != nil {
		t.Fatal(err)
	}
	got := r2.Get("caveman@caveman")
	if len(got) != 1 || got[0].GitCommitSHA != "ef6050c5" {
		t.Errorf("registry did not round-trip: %+v", got)
	}
}

func TestLoadRegistry_MissingFile(t *testing.T) {
	r, err := LoadRegistry(t.TempDir())
	if err != nil {
		t.Fatalf("expected empty registry, got: %v", err)
	}
	if r.Version != 2 {
		t.Errorf("default version = %d, want 2", r.Version)
	}
	if r.Plugins == nil {
		t.Error("Plugins map should be initialized")
	}
}
