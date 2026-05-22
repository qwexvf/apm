package cmd

import "testing"

func TestIsSkillArg(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		// plugins — marketplace ID never contains "/"
		{"figma@claude-plugins-official", false},
		{"figma@claude-plugins-official@^2.1.0", false},
		{"caveman@caveman", false},
		{"caveman@caveman@*", false},

		// skills — source is owner/repo, contains "/"
		{"frontend-design@vercel-labs/skills", true},
		{"fe@vercel-labs/skills:skills/frontend-design", true},
		{"fe@vercel-labs/skills:skills/frontend-design@^1.0.0", true},
		{"graphify@qwexvf/dotfiles:.claude/skills/graphify", true},

		// malformed — no @ at all
		{"noslash", false},

		// edge: subpath has no effect on detection (split on : first)
		{"name@owner/repo:weird:path", true},
	}
	for _, c := range cases {
		got := isSkillArg(c.in)
		if got != c.want {
			t.Errorf("isSkillArg(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
