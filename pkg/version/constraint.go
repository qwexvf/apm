package version

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
)

var shaRe = regexp.MustCompile(`^[0-9a-f]{7,40}$`)

// Kind classifies a version string.
type Kind int

const (
	KindSemver Kind = iota
	KindConstraint
	KindSHA
	KindBranch
	KindLatest
)

// Classify returns the kind of a raw version string.
func Classify(raw string) Kind {
	if raw == "" || raw == "*" || raw == "latest" {
		return KindLatest
	}
	if shaRe.MatchString(raw) {
		return KindSHA
	}
	if _, err := semver.NewVersion(raw); err == nil {
		return KindSemver
	}
	// constraint operators
	for _, op := range []string{"^", "~", ">=", "<=", ">", "<", "="} {
		if strings.HasPrefix(raw, op) {
			return KindConstraint
		}
	}
	// treat anything else as a branch/ref name
	return KindBranch
}

// MatchConstraint reports whether version v satisfies constraint c.
// c may be a semver constraint string or exact version; v must be a semver string.
func MatchConstraint(c, v string) (bool, error) {
	if c == "" || c == "*" || c == "latest" {
		return true, nil
	}
	sv, err := semver.NewVersion(v)
	if err != nil {
		return false, fmt.Errorf("invalid version %q: %w", v, err)
	}
	if _, err := semver.NewVersion(c); err == nil {
		// exact match
		ec, _ := semver.NewVersion(c)
		return sv.Equal(ec), nil
	}
	con, err := semver.NewConstraint(c)
	if err != nil {
		return false, fmt.Errorf("invalid constraint %q: %w", c, err)
	}
	return con.Check(sv), nil
}

// LatestMatching returns the highest stable version from versions that satisfies
// constraint c. Pre-release versions are skipped unless c explicitly references
// one or no stable versions match.
func LatestMatching(c string, versions []string) (string, error) {
	var best *semver.Version
	var bestPre *semver.Version // fallback if no stable matches
	for _, raw := range versions {
		v, err := semver.NewVersion(raw)
		if err != nil {
			continue
		}
		ok, err := MatchConstraint(c, raw)
		if err != nil || !ok {
			continue
		}
		if v.Prerelease() != "" {
			if bestPre == nil || v.GreaterThan(bestPre) {
				bestPre = v
			}
			continue
		}
		if best == nil || v.GreaterThan(best) {
			best = v
		}
	}
	if best != nil {
		return best.Original(), nil
	}
	if bestPre != nil {
		return bestPre.Original(), nil
	}
	return "", fmt.Errorf("no version matching %q", c)
}
