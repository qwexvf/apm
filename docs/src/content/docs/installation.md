---
title: Installation
description: Install apm on Linux, macOS, or Windows.
sidebar:
  order: 1
---

## go install

```sh
go install github.com/qwexvf/apm@latest
```

Requires Go 1.21+.

## From source

```sh
git clone https://github.com/qwexvf/apm
cd apm
go build -o apm .
sudo mv apm /usr/local/bin/
```

## Verify

```sh
apm --help
```

## GitHub token (recommended)

apm uses the GitHub API to resolve plugin versions. Without a token you are rate-limited to 60 requests/hour. With a token: 5,000/hour.

```sh
export GITHUB_TOKEN=ghp_...
```

Add to your shell profile to make it permanent. The token only needs the `public_repo` read scope — no write access required.

See [GitHub token setup](/reference/github-token/) for details.
