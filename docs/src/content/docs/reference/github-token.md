---
title: GitHub token setup
description: How to create and configure a GitHub token for ccpm.
sidebar:
  label: GitHub token
  order: 20
---

ccpm uses the GitHub API to resolve plugin versions and download archives. Without a token, requests are rate-limited to **60/hour** by IP. With a token: **5,000/hour**.

## Create a token

1. Go to [github.com/settings/tokens](https://github.com/settings/tokens)
2. Click **Generate new token (classic)**
3. Give it a name — e.g. `ccpm`
4. Select scopes: **no scopes needed** for public repos (the default read access is sufficient)
5. Click **Generate token**
6. Copy the token (shown once)

## Configure

```sh
export GITHUB_TOKEN=ghp_your_token_here
```

Add to your shell profile for persistence:

```sh
# ~/.bashrc or ~/.zshrc
export GITHUB_TOKEN=ghp_your_token_here
```

## Verify

```sh
curl -s -H "Authorization: Bearer $GITHUB_TOKEN" https://api.github.com/rate_limit \
  | jq '.rate'
```

Should show `"limit": 5000`.

## Fine-grained tokens

If you prefer [fine-grained tokens](https://github.com/settings/tokens?type=beta):

- **Repository access**: Public repositories (read-only)
- **Permissions**: Contents: Read-only

No other permissions needed.

> [!NOTE]
> ccpm never writes to GitHub. The token is only used for reading tags, commit SHAs, and downloading tarballs.
