package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/qwexvf/apm/internal/claude"
	"github.com/qwexvf/apm/internal/config"
)

// fakeGH implements ghClient: TagResolver + Downloader.
type fakeGH struct {
	tags     []string
	sha      string
	files    map[string]string
	wantRepo string // if set, refuse other repos
}

func (f *fakeGH) ListTags(_ context.Context, owner, repo string) ([]string, error) {
	return f.tags, nil
}
func (f *fakeGH) ResolveRef(_ context.Context, _, _, _ string) (string, error) {
	return f.sha, nil
}
func (f *fakeGH) LatestCommitSHA(_ context.Context, _, _ string) (string, error) {
	return f.sha, nil
}
func (f *fakeGH) DownloadTarball(_ context.Context, _, _, _, destDir string) (string, error) {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}
	for rel, body := range f.files {
		target := filepath.Join(destDir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return "", err
		}
		if err := os.WriteFile(target, []byte(body), 0644); err != nil {
			return "", err
		}
	}
	return "sha256:fake", nil
}

// setupLocalScope chdir's to a tmp dir, sets --local scope, swaps newGH to
// the supplied fake, and restores everything on cleanup.
func setupLocalScope(t *testing.T, gh ghClient) string {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)

	prevScope := localScope
	localScope = true
	t.Cleanup(func() { localScope = prevScope })

	prevGH := newGH
	newGH = func() ghClient { return gh }
	t.Cleanup(func() { newGH = prevGH })

	return dir
}

// seedManifest writes an apm.toml with the given marketplaces table.
func seedManifest(t *testing.T, dir string, mkts map[string]config.MarketplaceSource) {
	t.Helper()
	m := config.NewManifest()
	m.PluginManager.Scope = "local"
	for k, v := range mkts {
		m.Marketplaces[k] = v
	}
	if err := m.Save(dir); err != nil {
		t.Fatal(err)
	}
}

func TestRunAddPlugin_HappyPath(t *testing.T) {
	gh := &fakeGH{
		tags:  []string{"v2.1.0", "v2.0.0"},
		sha:   "deadbeefcafe1234567890abcdef1234567890ab",
		files: map[string]string{".claude-plugin/plugin.json": `{"name":"figma"}`},
	}
	dir := setupLocalScope(t, gh)
	seedManifest(t, dir, map[string]config.MarketplaceSource{
		"official": {Source: "github", Repo: "anthropics/claude-plugins-official"},
	})

	if err := runAddPlugin("figma@official@^2.1.0"); err != nil {
		t.Fatalf("runAddPlugin: %v", err)
	}

	m, err := config.LoadManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m.Plugins["figma@official"] != "^2.1.0" {
		t.Errorf("manifest plugin constraint = %q, want ^2.1.0", m.Plugins["figma@official"])
	}

	lock, err := config.LoadLock(dir)
	if err != nil {
		t.Fatal(err)
	}
	locked := lock.Get("figma@official")
	if locked == nil {
		t.Fatal("plugin not in lockfile")
	}
	if locked.CommitSHA != gh.sha {
		t.Errorf("CommitSHA = %q, want %q", locked.CommitSHA, gh.sha)
	}
	if locked.Version != "v2.1.0" {
		t.Errorf("Version = %q, want v2.1.0", locked.Version)
	}

	claudeDir := claude.Dir("local")
	want := claude.PluginInstallPath(claudeDir, "official", "figma", "v2.1.0")
	if _, err := os.Stat(filepath.Join(want, ".claude-plugin", "plugin.json")); err != nil {
		t.Errorf("plugin file not at install path: %v", err)
	}

	reg, err := claude.LoadRegistry(claudeDir)
	if err != nil {
		t.Fatal(err)
	}
	if entries := reg.Get("figma@official"); len(entries) == 0 {
		t.Error("plugin not in registry")
	}
}

func TestRunAddPlugin_UnknownMarketplace(t *testing.T) {
	gh := &fakeGH{}
	setupLocalScope(t, gh)

	err := runAddPlugin("figma@nope@^2.0.0")
	if err == nil || !strings.Contains(err.Error(), "unknown marketplace") {
		t.Errorf("expected unknown marketplace error, got: %v", err)
	}
}

func TestRunAddSkill_HappyPath(t *testing.T) {
	gh := &fakeGH{
		tags:  []string{"v1.0.0"},
		sha:   "abcdef0123456789abcdef0123456789abcdef01",
		files: map[string]string{"SKILL.md": "---\nname: frontend\n---"},
	}
	dir := setupLocalScope(t, gh)

	if err := runAddSkill("frontend@vercel-labs/skills@*"); err != nil {
		t.Fatalf("runAddSkill: %v", err)
	}

	m, err := config.LoadManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m.Skills["frontend@vercel-labs/skills"] != "*" {
		t.Errorf("skill not in manifest: %+v", m.Skills)
	}

	lock, err := config.LoadLock(dir)
	if err != nil {
		t.Fatal(err)
	}
	if lock.GetSkill("frontend@vercel-labs/skills") == nil {
		t.Error("skill not in lockfile")
	}

	claudeDir := claude.Dir("local")
	want := claude.SkillInstallPath(claudeDir, "frontend")
	if _, err := os.Stat(filepath.Join(want, "SKILL.md")); err != nil {
		t.Errorf("SKILL.md not at install path: %v", err)
	}
}

func TestRunAddSkill_OpenCodeTarget(t *testing.T) {
	gh := &fakeGH{
		tags:  []string{"v1.0.0"},
		sha:   "abcdef0123456789abcdef0123456789abcdef01",
		files: map[string]string{"SKILL.md": "---\nname: frontend\n---"},
	}
	dir := setupLocalScope(t, gh)

	// manifest pinned to the opencode target
	m := config.NewManifest()
	m.PluginManager.Scope = "local"
	m.PluginManager.Target = "opencode"
	if err := m.Save(dir); err != nil {
		t.Fatal(err)
	}

	if err := runAddSkill("frontend@vercel-labs/skills@*"); err != nil {
		t.Fatalf("runAddSkill: %v", err)
	}

	want := filepath.Join(dir, ".opencode", "skills", "frontend", "SKILL.md")
	if _, err := os.Stat(want); err != nil {
		t.Errorf("SKILL.md not at opencode install path: %v", err)
	}

	lock, err := config.LoadLock(dir)
	if err != nil {
		t.Fatal(err)
	}
	if lock.GetSkill("frontend@vercel-labs/skills") == nil {
		t.Error("skill not in lockfile")
	}
}

func TestRunAddPlugin_OpenCodeTarget_Rejected(t *testing.T) {
	gh := &fakeGH{}
	dir := setupLocalScope(t, gh)

	m := config.NewManifest()
	m.PluginManager.Scope = "local"
	m.PluginManager.Target = "opencode"
	if err := m.Save(dir); err != nil {
		t.Fatal(err)
	}

	err := runAddPlugin("figma@official@^2.1.0")
	if err == nil || !strings.Contains(err.Error(), "not supported for target") {
		t.Errorf("expected unsupported-target error, got: %v", err)
	}
}

func TestRunRemovePlugin_RoundTrip(t *testing.T) {
	gh := &fakeGH{
		tags:  []string{"v2.1.0"},
		sha:   "deadbeef",
		files: map[string]string{".claude-plugin/plugin.json": `{}`},
	}
	dir := setupLocalScope(t, gh)
	seedManifest(t, dir, map[string]config.MarketplaceSource{
		"official": {Source: "github", Repo: "anthropics/claude-plugins-official"},
	})

	if err := runAddPlugin("figma@official@^2.1.0"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := runRemovePlugin("figma@official"); err != nil {
		t.Fatalf("remove: %v", err)
	}

	m, _ := config.LoadManifest(dir)
	if _, ok := m.Plugins["figma@official"]; ok {
		t.Error("plugin still in manifest after remove")
	}
	lock, _ := config.LoadLock(dir)
	if lock.Get("figma@official") != nil {
		t.Error("plugin still in lockfile after remove")
	}

	claudeDir := claude.Dir("local")
	installed := claude.PluginInstallPath(claudeDir, "official", "figma", "v2.1.0")
	if _, err := os.Stat(installed); err == nil {
		t.Error("install dir still exists after remove")
	}
}

func TestRunRemoveSkill_RoundTrip(t *testing.T) {
	gh := &fakeGH{
		tags:  []string{"v1.0.0"},
		sha:   "abc",
		files: map[string]string{"SKILL.md": "x"},
	}
	dir := setupLocalScope(t, gh)

	if err := runAddSkill("frontend@vercel-labs/skills"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := runRemoveSkill("frontend@vercel-labs/skills"); err != nil {
		t.Fatalf("remove: %v", err)
	}

	m, _ := config.LoadManifest(dir)
	if _, ok := m.Skills["frontend@vercel-labs/skills"]; ok {
		t.Error("skill still in manifest after remove")
	}
	lock, _ := config.LoadLock(dir)
	if lock.GetSkill("frontend@vercel-labs/skills") != nil {
		t.Error("skill still in lockfile after remove")
	}

	claudeDir := claude.Dir("local")
	skillPath := claude.SkillInstallPath(claudeDir, "frontend")
	if _, err := os.Stat(skillPath); err == nil {
		t.Error("skill dir still exists after remove")
	}
}

func TestRunLockAndInstall_RoundTrip(t *testing.T) {
	gh := &fakeGH{
		tags: []string{"v2.1.0"},
		sha:  "deadbeef",
		files: map[string]string{
			".claude-plugin/plugin.json": `{}`,
			"SKILL.md":                   "x",
		},
	}
	dir := setupLocalScope(t, gh)

	// seed manifest with one plugin + one skill, no lock yet
	m := config.NewManifest()
	m.PluginManager.Scope = "local"
	m.Marketplaces["official"] = config.MarketplaceSource{Source: "github", Repo: "anthropics/claude-plugins-official"}
	m.AddPlugin("figma@official", "^2.0.0")
	m.AddSkill("frontend@vercel-labs/skills", "*")
	if err := m.Save(dir); err != nil {
		t.Fatal(err)
	}

	// `apm lock` populates the lockfile
	if err := lockCmd.RunE(lockCmd, nil); err != nil {
		t.Fatalf("lock: %v", err)
	}
	lock, _ := config.LoadLock(dir)
	if lock.Get("figma@official") == nil {
		t.Error("plugin not locked")
	}
	if lock.GetSkill("frontend@vercel-labs/skills") == nil {
		t.Error("skill not locked")
	}

	// `apm install` from lockfile
	if err := installCmd.RunE(installCmd, nil); err != nil {
		t.Fatalf("install: %v", err)
	}
	claudeDir := claude.Dir("local")
	pluginFile := filepath.Join(claude.PluginInstallPath(claudeDir, "official", "figma", "v2.1.0"), ".claude-plugin", "plugin.json")
	if _, err := os.Stat(pluginFile); err != nil {
		t.Error("plugin not installed from lockfile")
	}
	skillFile := filepath.Join(claude.SkillInstallPath(claudeDir, "frontend"), "SKILL.md")
	if _, err := os.Stat(skillFile); err != nil {
		t.Error("skill not installed from lockfile")
	}
}
