package resolver

import (
	"context"
	"fmt"
	"strings"

	"github.com/qwexvf/apm/internal/fetcher"
	"github.com/qwexvf/apm/pkg/version"
)

// Result is the resolved concrete version for a plugin.
type Result struct {
	Version   string // semver tag or short SHA
	CommitSHA string // full 40-char SHA
}

// Resolve resolves a version constraint to a concrete version + commit SHA.
// repo is "owner/repo", constraint is e.g. "^2.1.0", "*", "latest", "abc123", "main".
func Resolve(ctx context.Context, gh *fetcher.GitHub, repo, constraint string) (*Result, error) {
	owner, name, ok := strings.Cut(repo, "/")
	if !ok {
		return nil, fmt.Errorf("invalid repo %q: expected owner/repo", repo)
	}

	kind := version.Classify(constraint)

	switch kind {
	case version.KindSHA:
		sha, err := gh.ResolveRef(ctx, owner, name, constraint)
		if err != nil {
			return nil, err
		}
		short := constraint
		if len(sha) >= 12 {
			short = sha[:12]
		}
		return &Result{Version: short, CommitSHA: sha}, nil

	case version.KindBranch:
		sha, err := gh.ResolveRef(ctx, owner, name, constraint)
		if err != nil {
			return nil, err
		}
		short := sha
		if len(sha) >= 12 {
			short = sha[:12]
		}
		return &Result{Version: short, CommitSHA: sha}, nil

	case version.KindLatest:
		tags, err := gh.ListTags(ctx, owner, name)
		if err != nil {
			return nil, err
		}
		if len(tags) > 0 {
			// try semver tags first
			best, err := version.LatestMatching("*", tags)
			if err == nil {
				sha, err := gh.ResolveRef(ctx, owner, name, best)
				if err != nil {
					return nil, err
				}
				return &Result{Version: best, CommitSHA: sha}, nil
			}
			// fall back to first tag
			sha, err := gh.ResolveRef(ctx, owner, name, tags[0])
			if err != nil {
				return nil, err
			}
			return &Result{Version: tags[0], CommitSHA: sha}, nil
		}
		// no tags: use HEAD
		sha, err := gh.LatestCommitSHA(ctx, owner, name)
		if err != nil {
			return nil, err
		}
		return &Result{Version: sha[:12], CommitSHA: sha}, nil

	case version.KindSemver, version.KindConstraint:
		tags, err := gh.ListTags(ctx, owner, name)
		if err != nil {
			return nil, err
		}
		best, err := version.LatestMatching(constraint, tags)
		if err != nil {
			return nil, fmt.Errorf("no version matching %q in %s: %w", constraint, repo, err)
		}
		sha, err := gh.ResolveRef(ctx, owner, name, best)
		if err != nil {
			return nil, err
		}
		return &Result{Version: best, CommitSHA: sha}, nil
	}

	return nil, fmt.Errorf("unhandled version kind for %q", constraint)
}
