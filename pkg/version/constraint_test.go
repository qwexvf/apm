package version

import "testing"

func TestClassify(t *testing.T) {
	cases := []struct{ in string; want Kind }{
		{"", KindLatest},
		{"*", KindLatest},
		{"latest", KindLatest},
		{"1.0.0", KindSemver},
		{"2.1.30", KindSemver},
		{"^2.1.0", KindConstraint},
		{"~1.0.0", KindConstraint},
		{">=2.0.0", KindConstraint},
		{"ef6050c5e184", KindSHA},
		{"abc1234", KindSHA},
		{"main", KindBranch},
		{"v2", KindSemver}, // semver.NewVersion treats "v2" as "2.0.0"
		{"feature/x", KindBranch},
	}
	for _, c := range cases {
		if got := Classify(c.in); got != c.want {
			t.Errorf("Classify(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestMatchConstraint(t *testing.T) {
	cases := []struct {
		constraint, version string
		want                bool
	}{
		{"*", "2.1.30", true},
		{"", "1.0.0", true},
		{"^2.1.0", "2.1.30", true},
		{"^2.1.0", "2.0.0", false},
		{"^2.1.0", "3.0.0", false},
		{"~2.1.0", "2.1.5", true},
		{"~2.1.0", "2.2.0", false},
		{"1.0.0", "1.0.0", true},
		{"1.0.0", "1.0.1", false},
	}
	for _, c := range cases {
		got, err := MatchConstraint(c.constraint, c.version)
		if err != nil {
			t.Errorf("MatchConstraint(%q, %q): %v", c.constraint, c.version, err)
			continue
		}
		if got != c.want {
			t.Errorf("MatchConstraint(%q, %q) = %v, want %v", c.constraint, c.version, got, c.want)
		}
	}
}

func TestLatestMatching(t *testing.T) {
	versions := []string{"1.0.0", "2.0.0", "2.1.0", "2.1.30", "3.0.0-alpha"}
	best, err := LatestMatching("^2.1.0", versions)
	if err != nil || best != "2.1.30" {
		t.Errorf("LatestMatching(^2.1.0) = %q, %v; want 2.1.30", best, err)
	}
	best, err = LatestMatching("*", versions)
	if err != nil || best != "2.1.30" {
		t.Errorf("LatestMatching(*) = %q, %v; want 2.1.30", best, err)
	}
	if _, err := LatestMatching("^9.0.0", versions); err == nil {
		t.Error("expected error for no matching version")
	}
}
