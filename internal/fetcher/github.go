package fetcher

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v67/github"
	"golang.org/x/oauth2"
)

// GitHub fetches plugin information and archives from GitHub.
type GitHub struct {
	client *github.Client
}

// NewGitHub creates a GitHub client, using GITHUB_TOKEN if set.
func NewGitHub() *GitHub {
	token := os.Getenv("GITHUB_TOKEN")
	var hc *http.Client
	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		hc = oauth2.NewClient(context.Background(), ts)
	}
	return &GitHub{client: github.NewClient(hc)}
}

// ListTags returns all tag names for owner/repo.
func (g *GitHub) ListTags(ctx context.Context, owner, repo string) ([]string, error) {
	var tags []string
	opts := &github.ListOptions{PerPage: 100}
	for {
		ts, resp, err := g.client.Repositories.ListTags(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("list tags %s/%s: %w", owner, repo, err)
		}
		for _, t := range ts {
			tags = append(tags, t.GetName())
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return tags, nil
}

// ResolveRef returns the full commit SHA for a ref (tag, branch, or SHA).
func (g *GitHub) ResolveRef(ctx context.Context, owner, repo, ref string) (string, error) {
	commit, _, err := g.client.Repositories.GetCommit(ctx, owner, repo, ref, nil)
	if err != nil {
		return "", fmt.Errorf("resolve ref %q in %s/%s: %w", ref, owner, repo, err)
	}
	return commit.GetSHA(), nil
}

// LatestCommitSHA returns the HEAD commit SHA for the default branch.
func (g *GitHub) LatestCommitSHA(ctx context.Context, owner, repo string) (string, error) {
	r, _, err := g.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return "", fmt.Errorf("get repo %s/%s: %w", owner, repo, err)
	}
	return g.ResolveRef(ctx, owner, repo, r.GetDefaultBranch())
}

// DownloadTarball downloads the tarball for owner/repo at ref and extracts it to destDir.
// Returns the SHA256 integrity hash of the downloaded tarball.
func (g *GitHub) DownloadTarball(ctx context.Context, owner, repo, ref, destDir string) (integrity string, err error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/tarball/%s", owner, repo, ref)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := g.client.Client().Do(req)
	if err != nil {
		return "", fmt.Errorf("download tarball: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download tarball: HTTP %d", resp.StatusCode)
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}

	integrity, err = extractTarGz(resp.Body, destDir)
	if err != nil {
		return "", fmt.Errorf("extract: %w", err)
	}
	return integrity, nil
}

// extractTarGz streams a .tar.gz from r into destDir, stripping the first path component.
// Returns sha256 hex of the raw stream content.
func extractTarGz(r io.Reader, destDir string) (string, error) {
	hasher := newHashReader(r)

	gz, err := gzip.NewReader(hasher)
	if err != nil {
		return "", err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// strip leading directory component (GitHub adds "owner-repo-sha/")
		parts := strings.SplitN(hdr.Name, "/", 2)
		if len(parts) < 2 || parts[1] == "" {
			continue
		}
		rel := parts[1]

		target := filepath.Join(destDir, filepath.FromSlash(rel))
		// guard against zip-slip
		if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) {
			continue
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return "", err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return "", err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode)|0600)
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return "", err
			}
			if err := f.Close(); err != nil {
				return "", err
			}
		case tar.TypeSymlink:
			// reject absolute symlink targets and any that escape destDir
			if filepath.IsAbs(hdr.Linkname) {
				continue
			}
			linkTarget := filepath.Join(filepath.Dir(target), hdr.Linkname)
			if !strings.HasPrefix(filepath.Clean(linkTarget), filepath.Clean(destDir)+string(os.PathSeparator)) {
				continue
			}
			os.Remove(target)
			if err := os.Symlink(hdr.Linkname, target); err != nil {
				return "", err
			}
		}
	}

	return "sha256:" + hasher.HexSum(), nil
}
