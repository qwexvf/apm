package config

import "testing"

func TestLockUpsertGet(t *testing.T) {
	l := NewLock()
	l.Upsert(LockedPlugin{ID: "figma@claude-plugins-official", Version: "2.1.30", CommitSHA: "abc123"})

	p := l.Get("figma@claude-plugins-official")
	if p == nil || p.Version != "2.1.30" {
		t.Fatal("Get failed")
	}

	// update
	l.Upsert(LockedPlugin{ID: "figma@claude-plugins-official", Version: "2.2.0", CommitSHA: "def456"})
	if l.Get("figma@claude-plugins-official").Version != "2.2.0" {
		t.Error("Upsert did not update")
	}
	if len(l.Plugins) != 1 {
		t.Error("Upsert created duplicate")
	}
}

func TestLockRemove(t *testing.T) {
	l := NewLock()
	l.Upsert(LockedPlugin{ID: "a@b"})
	l.Upsert(LockedPlugin{ID: "c@d"})
	l.Remove("a@b")
	if l.Get("a@b") != nil {
		t.Error("Remove did not delete")
	}
	if l.Get("c@d") == nil {
		t.Error("Remove deleted wrong entry")
	}
}

func TestLockSaveLoad(t *testing.T) {
	dir := t.TempDir()
	l := NewLock()
	l.Upsert(LockedPlugin{
		ID:          "caveman@caveman",
		Version:     "ef6050c5e184",
		CommitSHA:   "ef6050c5e1848b6880ff47c32ade1a608a64f85e",
		ResolvedURL: "https://github.com/JuliusBrussee/caveman",
		InstallPath: "/tmp/caveman",
		Integrity:   "sha256:abc",
	})
	if err := l.Save(dir); err != nil {
		t.Fatal(err)
	}
	l2, err := LoadLock(dir)
	if err != nil {
		t.Fatal(err)
	}
	p := l2.Get("caveman@caveman")
	if p == nil || p.CommitSHA != "ef6050c5e1848b6880ff47c32ade1a608a64f85e" {
		t.Error("lock did not round-trip")
	}
}

func TestLoadLock_missingFile(t *testing.T) {
	l, err := LoadLock(t.TempDir())
	if err != nil {
		t.Fatalf("expected empty lock for missing file, got: %v", err)
	}
	if len(l.Plugins) != 0 {
		t.Error("expected empty plugins")
	}
}

func TestLockSkillUpsertGet(t *testing.T) {
	l := NewLock()
	id := "frontend-design@vercel-labs/skills:skills/frontend-design"
	l.UpsertSkill(LockedSkill{ID: id, Version: "1.0.0", CommitSHA: "abc"})

	s := l.GetSkill(id)
	if s == nil || s.Version != "1.0.0" {
		t.Fatal("GetSkill failed")
	}

	// update should not duplicate
	l.UpsertSkill(LockedSkill{ID: id, Version: "1.1.0", CommitSHA: "def"})
	if l.GetSkill(id).Version != "1.1.0" {
		t.Error("UpsertSkill did not update")
	}
	if len(l.Skills) != 1 {
		t.Error("UpsertSkill created duplicate")
	}
}

func TestLockSkillRemove(t *testing.T) {
	l := NewLock()
	l.UpsertSkill(LockedSkill{ID: "a@x/y"})
	l.UpsertSkill(LockedSkill{ID: "b@x/z"})
	if !l.RemoveSkill("a@x/y") {
		t.Error("RemoveSkill returned false for present id")
	}
	if l.GetSkill("a@x/y") != nil {
		t.Error("RemoveSkill did not delete")
	}
	if l.GetSkill("b@x/z") == nil {
		t.Error("RemoveSkill deleted wrong entry")
	}
	if l.RemoveSkill("not-there") {
		t.Error("RemoveSkill returned true for missing id")
	}
}

func TestLockSkillSaveLoad(t *testing.T) {
	dir := t.TempDir()
	l := NewLock()
	l.UpsertSkill(LockedSkill{
		ID:          "graphify@qwexvf/dotfiles:.claude/skills/graphify",
		Version:     "main",
		CommitSHA:   "ef6050c5e1848b6880ff47c32ade1a608a64f85e",
		ResolvedURL: "https://github.com/qwexvf/dotfiles",
		InstallPath: "/tmp/.claude/skills/graphify",
		Integrity:   "sha256:abc",
	})
	if err := l.Save(dir); err != nil {
		t.Fatal(err)
	}
	l2, err := LoadLock(dir)
	if err != nil {
		t.Fatal(err)
	}
	s := l2.GetSkill("graphify@qwexvf/dotfiles:.claude/skills/graphify")
	if s == nil || s.CommitSHA != "ef6050c5e1848b6880ff47c32ade1a608a64f85e" {
		t.Error("skill lock did not round-trip")
	}
	if s.InstallPath != "/tmp/.claude/skills/graphify" {
		t.Error("InstallPath did not round-trip")
	}
}
