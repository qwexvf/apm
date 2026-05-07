---
title: Installation
description: Install ccpm on Linux, macOS, or Windows.
sidebar:
  order: 1
---

## go install

```sh
go install github.com/qwexvf/ccpm@latest
```

Requires Go 1.21+.

## From source

```sh
git clone https://github.com/qwexvf/ccpm
cd ccpm
go build -o ccpm .
sudo mv ccpm /usr/local/bin/
```

## Verify

```sh
ccpm --help
```

## GitHub token (recommended)

ccpm uses the GitHub API to resolve plugin versions. Without a token you are rate-limited to 60 requests/hour. With a token: 5,000/hour.

```sh
export GITHUB_TOKEN=ghp_...
```

Add to your shell profile to make it permanent. The token only needs the `public_repo` read scope — no write access required.

See [GitHub token setup](/reference/github-token/) for details.
