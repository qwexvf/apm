---
title: Configuration reference
description: All ccpm.toml fields and their defaults.
sidebar:
  label: Config reference
  order: 10
---

## `ccpm.toml` fields

### `[plugin_manager]`

```toml
[plugin_manager]
scope = "user"
```

| Key | Type | Default | Values |
|-----|------|---------|--------|
| `scope` | string | `"user"` | `"user"` installs to `~/.claude/`; `"local"` installs to `./.claude/` |

### `[plugins]`

```toml
[plugins]
"name@marketplace" = "<constraint>"
```

Keys are plugin IDs (`name@marketplace`). Values are [version constraints](/guides/version-constraints/).

### `[marketplaces]`

```toml
[marketplaces]
"id" = { source = "github", repo = "owner/repo" }
```

| Key | Type | Description |
|-----|------|-------------|
| `source` | string | Always `"github"` |
| `repo` | string | GitHub repo in `owner/repo` format |

## Environment variables

| Variable | Description |
|----------|-------------|
| `GITHUB_TOKEN` | Personal access token for GitHub API. Raises rate limit from 60 to 5,000 req/hour. Read scope only — `public_repo` is sufficient. |

## Claude Code files managed by ccpm

ccpm reads and writes these files in your Claude Code config directory:

| File | Description |
|------|-------------|
| `plugins/installed_plugins.json` | Registry of installed plugins (v2 schema) |
| `settings.json` | `enabledPlugins` and `extraKnownMarketplaces` keys |
| `plugins/known_marketplaces.json` | Marketplace ID → GitHub repo mapping |
| `plugins/cache/<marketplace>/<plugin>/<version>/` | Extracted plugin files |
| `plugins/marketplaces/<marketplace>/` | Cloned marketplace git repos |

ccpm uses atomic writes (write to `.tmp`, then rename) when modifying JSON files and preserves all unrelated keys in `settings.json`.
