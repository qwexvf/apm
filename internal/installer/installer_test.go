package installer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeDownloader writes a pre-defined fixture tree into destDir instead of
// hitting the network. Each entry is "relative/path" → "file contents". A
// path ending in "/" creates an empty directory.
type fakeDownloader struct {
	files map[string]string
}

func (f *fakeDownloader) DownloadTarball(_ context.Context, _, _, _, destDir string) (string, error) {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}
	for rel, body := range f.files {
		target := filepath.Join(destDir, filepath.FromSlash(rel))
		if strings.HasSuffix(rel, "/") {
			if err := os.MkdirAll(target, 0755); err != nil {
				return "", err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return "", err
		}
		if err := os.WriteFile(target, []byte(body), 0644); err != nil {
			return "", err
		}
	}
	return "sha256:fake", nil
}

func TestInstallSkill_HappyPath_RepoRoot(t *testing.T) {
	claudeDir := t.TempDir()
	gh := &fakeDownloader{files: map[string]string{
		"SKILL.md":  "---\nname: foo\n---\nbody",
		"README.md": "# foo",
	}}

	res, err := InstallSkill(context.Background(), gh, claudeDir, "foo", "owner/repo", "v1.0.0", "")
	if err != nil {
		t.Fatalf("InstallSkill: %v", err)
	}

	wantPath := filepath.Join(claudeDir, "skills", "foo")
	if res.InstallPath != wantPath {
		t.Errorf("InstallPath = %q, want %q", res.InstallPath, wantPath)
	}
	if _, err := os.Stat(filepath.Join(wantPath, "SKILL.md")); err != nil {
		t.Error("SKILL.md not placed at install path")
	}
	if res.Integrity != "sha256:fake" {
		t.Errorf("Integrity = %q, want fake hash", res.Integrity)
	}
}

func TestInstallSkill_HappyPath_Subpath(t *testing.T) {
	claudeDir := t.TempDir()
	gh := &fakeDownloader{files: map[string]string{
		"skills/foo/SKILL.md":  "---\nname: foo\n---",
		"skills/bar/SKILL.md":  "---\nname: bar\n---",
		"README.md":            "# repo",
	}}

	res, err := InstallSkill(context.Background(), gh, claudeDir, "foo", "owner/repo", "main", "skills/foo")
	if err != nil {
		t.Fatalf("InstallSkill: %v", err)
	}
	if _, err := os.Stat(filepath.Join(res.InstallPath, "SKILL.md")); err != nil {
		t.Error("SKILL.md not placed at install path from subpath")
	}
	// sibling subpath should NOT have leaked into install dir
	if _, err := os.Stat(filepath.Join(res.InstallPath, "..", "bar")); err == nil {
		t.Error("sibling subpath leaked alongside install dir")
	}
}

func TestInstallSkill_RejectTraversalSubpath(t *testing.T) {
	claudeDir := t.TempDir()
	gh := &fakeDownloader{files: map[string]string{"SKILL.md": "x"}}

	bad := []string{"../etc", "foo/../../etc", "/abs/path"}
	for _, sp := range bad {
		_, err := InstallSkill(context.Background(), gh, claudeDir, "foo", "owner/repo", "v1", sp)
		if err == nil {
			t.Errorf("subpath %q: expected rejection, got nil", sp)
		}
	}
}

func TestInstallSkill_MissingSKILLMD(t *testing.T) {
	claudeDir := t.TempDir()
	gh := &fakeDownloader{files: map[string]string{
		"README.md": "no skill here",
	}}

	_, err := InstallSkill(context.Background(), gh, claudeDir, "foo", "owner/repo", "v1", "")
	if err == nil {
		t.Fatal("expected error when SKILL.md absent")
	}
	// final dir should not exist
	if _, err := os.Stat(filepath.Join(claudeDir, "skills", "foo")); err == nil {
		t.Error("install dir created despite missing SKILL.md")
	}
}

func TestInstallSkill_SubpathNotFound(t *testing.T) {
	claudeDir := t.TempDir()
	gh := &fakeDownloader{files: map[string]string{"SKILL.md": "x"}}

	_, err := InstallSkill(context.Background(), gh, claudeDir, "foo", "owner/repo", "v1", "skills/nope")
	if err == nil {
		t.Fatal("expected error for missing subpath")
	}
}

func TestInstall_Plugin_HappyPath(t *testing.T) {
	claudeDir := t.TempDir()
	gh := &fakeDownloader{files: map[string]string{
		".claude-plugin/plugin.json": `{"name":"figma"}`,
		"commands/foo.md":            "body",
	}}

	res, err := Install(context.Background(), gh, claudeDir, "official", "figma", "anthropics/figma", "v2.1.0")
	if err != nil {
		t.Fatalf("Install: %v", err)
	}

	want := filepath.Join(claudeDir, "plugins", "cache", "official", "figma", "v2.1.0")
	if res.InstallPath != want {
		t.Errorf("InstallPath = %q, want %q", res.InstallPath, want)
	}
	if _, err := os.Stat(filepath.Join(want, ".claude-plugin", "plugin.json")); err != nil {
		t.Error("plugin.json not at install path")
	}
	if res.Integrity != "sha256:fake" {
		t.Errorf("Integrity = %q, want fake hash", res.Integrity)
	}
}

func TestInstall_Plugin_CachedSkip(t *testing.T) {
	claudeDir := t.TempDir()
	gh := &fakeDownloader{files: map[string]string{
		".claude-plugin/plugin.json": `{}`,
	}}

	first, err := Install(context.Background(), gh, claudeDir, "official", "figma", "anthropics/figma", "v2.1.0")
	if err != nil {
		t.Fatal(err)
	}

	// second call should skip download — same path, empty integrity
	second, err := Install(context.Background(), gh, claudeDir, "official", "figma", "anthropics/figma", "v2.1.0")
	if err != nil {
		t.Fatal(err)
	}
	if first.InstallPath != second.InstallPath {
		t.Error("cached call returned different path")
	}
	if second.Integrity != "" {
		t.Errorf("cached call returned integrity %q, want empty", second.Integrity)
	}
}

func TestInstall_Plugin_InvalidRepo(t *testing.T) {
	gh := &fakeDownloader{files: map[string]string{".claude-plugin/plugin.json": "{}"}}
	if _, err := Install(context.Background(), gh, t.TempDir(), "m", "p", "nopath", "v1"); err == nil {
		t.Error("expected error for repo without /")
	}
}

func TestUninstall(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "to-remove")
	if err := os.MkdirAll(target, 0755); err != nil {
		t.Fatal(err)
	}
	if err := Uninstall(target); err != nil {
		t.Errorf("Uninstall existing: %v", err)
	}
	if _, err := os.Stat(target); err == nil {
		t.Error("Uninstall did not remove dir")
	}
	if err := Uninstall(target); err != nil {
		t.Errorf("Uninstall missing path: %v", err)
	}
	if err := Uninstall(""); err != nil {
		t.Errorf("Uninstall empty path: %v", err)
	}
}

func TestInstallSkill_OverwritesExisting(t *testing.T) {
	claudeDir := t.TempDir()
	final := filepath.Join(claudeDir, "skills", "foo")
	if err := os.MkdirAll(final, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(final, "stale.txt"), []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	gh := &fakeDownloader{files: map[string]string{
		"SKILL.md": "fresh",
	}}
	if _, err := InstallSkill(context.Background(), gh, claudeDir, "foo", "owner/repo", "v2", ""); err != nil {
		t.Fatalf("InstallSkill: %v", err)
	}

	if _, err := os.Stat(filepath.Join(final, "stale.txt")); err == nil {
		t.Error("stale file from previous install survived")
	}
	body, _ := os.ReadFile(filepath.Join(final, "SKILL.md"))
	if string(body) != "fresh" {
		t.Errorf("SKILL.md = %q, want %q", body, "fresh")
	}
}
