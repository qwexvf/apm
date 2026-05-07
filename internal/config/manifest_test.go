package config

import (
	"os"
	"testing"
)

func TestParsePluginArg(t *testing.T) {
	cases := []struct {
		in         string
		name, mkt  string
		constraint string
		wantErr    bool
	}{
		{"figma@claude-plugins-official", "figma", "claude-plugins-official", "*", false},
		{"caveman@caveman@^1.0.0", "caveman", "caveman", "^1.0.0", false},
		{"gopls-lsp@claude-plugins-official@1.0.0", "gopls-lsp", "claude-plugins-official", "1.0.0", false},
		{"noslash", "", "", "", true},
	}
	for _, c := range cases {
		n, m, con, err := ParsePluginArg(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("ParsePluginArg(%q): expected error", c.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParsePluginArg(%q): %v", c.in, err)
			continue
		}
		if n != c.name || m != c.mkt || con != c.constraint {
			t.Errorf("ParsePluginArg(%q) = (%q,%q,%q), want (%q,%q,%q)",
				c.in, n, m, con, c.name, c.mkt, c.constraint)
		}
	}
}

func TestSplitID(t *testing.T) {
	n, m, err := SplitID("figma@claude-plugins-official")
	if err != nil || n != "figma" || m != "claude-plugins-official" {
		t.Errorf("SplitID failed: %v %v %v", n, m, err)
	}
	if _, _, err := SplitID("noslash"); err == nil {
		t.Error("expected error for missing @")
	}
}

func TestManifestAddRemove(t *testing.T) {
	m := NewManifest()
	m.AddPlugin("figma@claude-plugins-official", "^2.1.0")
	if m.Plugins["figma@claude-plugins-official"] != "^2.1.0" {
		t.Error("plugin not added")
	}
	if !m.RemovePlugin("figma@claude-plugins-official") {
		t.Error("remove returned false")
	}
	if len(m.Plugins) != 0 {
		t.Error("plugin not removed")
	}
}

func TestManifestSaveLoad(t *testing.T) {
	dir := t.TempDir()
	m := NewManifest()
	m.PluginManager.Scope = "user"
	m.AddPlugin("caveman@caveman", "*")
	m.Marketplaces["caveman"] = MarketplaceSource{Source: "github", Repo: "JuliusBrussee/caveman"}

	if err := m.Save(dir); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(dir + "/apm.toml"); err != nil {
		t.Fatal("apm.toml not created")
	}

	m2, err := LoadManifest(dir)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if m2.Plugins["caveman@caveman"] != "*" {
		t.Error("plugin not round-tripped")
	}
	if m2.Marketplaces["caveman"].Repo != "JuliusBrussee/caveman" {
		t.Error("marketplace not round-tripped")
	}
}
