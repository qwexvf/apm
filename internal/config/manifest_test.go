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

func TestParseSkillArg(t *testing.T) {
	cases := []struct {
		in                          string
		name, repo, subpath, constr string
		wantErr                     bool
	}{
		{"frontend-design@vercel-labs/skills", "frontend-design", "vercel-labs/skills", "", "*", false},
		{"fe@vercel-labs/skills:frontend-design", "fe", "vercel-labs/skills", "frontend-design", "*", false},
		{"fe@vercel-labs/skills:skills/foo@^1.0.0", "fe", "vercel-labs/skills", "skills/foo", "^1.0.0", false},
		{"name@notarepo", "", "", "", "", true},
		{"@vercel-labs/skills", "", "", "", "", true},
	}
	for _, c := range cases {
		n, r, sp, con, err := ParseSkillArg(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("ParseSkillArg(%q): expected error", c.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseSkillArg(%q): %v", c.in, err)
			continue
		}
		if n != c.name || r != c.repo || sp != c.subpath || con != c.constr {
			t.Errorf("ParseSkillArg(%q) = (%q,%q,%q,%q), want (%q,%q,%q,%q)",
				c.in, n, r, sp, con, c.name, c.repo, c.subpath, c.constr)
		}
	}
}

func TestSplitSkillID(t *testing.T) {
	n, r, sp, err := SplitSkillID("fe@vercel-labs/skills:frontend-design")
	if err != nil || n != "fe" || r != "vercel-labs/skills" || sp != "frontend-design" {
		t.Errorf("SplitSkillID failed: %v %v %v %v", n, r, sp, err)
	}
	if _, _, _, err := SplitSkillID("bad@notarepo"); err == nil {
		t.Error("expected error for repo missing /")
	}
}

func TestSkillAddRemove(t *testing.T) {
	m := NewManifest()
	id := SkillID("frontend-design", "vercel-labs/skills", "")
	m.AddSkill(id, "*")
	if m.Skills[id] != "*" {
		t.Error("skill not added")
	}
	if !m.RemoveSkill(id) {
		t.Error("remove returned false")
	}
	if len(m.Skills) != 0 {
		t.Error("skill not removed")
	}
}

func TestManifestSaveLoad(t *testing.T) {
	dir := t.TempDir()
	m := NewManifest()
	m.PluginManager.Scope = "user"
	m.AddPlugin("caveman@caveman", "*")
	m.AddSkill("frontend-design@vercel-labs/skills", "^1.0.0")
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
	if m2.Skills["frontend-design@vercel-labs/skills"] != "^1.0.0" {
		t.Error("skill not round-tripped")
	}
	if m2.Marketplaces["caveman"].Repo != "JuliusBrussee/caveman" {
		t.Error("marketplace not round-tripped")
	}
}

func TestManifestTargetRoundTrip(t *testing.T) {
	dir := t.TempDir()
	m := NewManifest()
	m.PluginManager.Target = "opencode"
	if err := m.Save(dir); err != nil {
		t.Fatalf("Save: %v", err)
	}

	m2, err := LoadManifest(dir)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if got := m2.PluginManager.TargetOrDefault(); got != "opencode" {
		t.Errorf("target = %q, want opencode", got)
	}
}

func TestTargetOrDefault(t *testing.T) {
	var c PluginManagerConfig
	if got := c.TargetOrDefault(); got != "claude" {
		t.Errorf("empty target = %q, want claude", got)
	}
	c.Target = "opencode"
	if got := c.TargetOrDefault(); got != "opencode" {
		t.Errorf("target = %q, want opencode", got)
	}
}
