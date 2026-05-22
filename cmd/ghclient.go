package cmd

import (
	"github.com/qwexvf/apm/internal/fetcher"
	"github.com/qwexvf/apm/internal/installer"
	"github.com/qwexvf/apm/internal/resolver"
)

// ghClient is the combined surface that commands need from the GitHub fetcher:
// tag/ref resolution + tarball download. Tests swap newGH to inject a fake.
type ghClient interface {
	resolver.TagResolver
	installer.Downloader
}

var newGH = func() ghClient { return fetcher.NewGitHub() }
