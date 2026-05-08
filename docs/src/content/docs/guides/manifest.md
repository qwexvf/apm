---
title: Manifest format
description: Reference for apm.toml — the declarative plugin manifest.
sidebar:
  label: Manifest (apm.toml)
  order: 10
---

`apm.toml` is the declarative manifest that lists which plugins your environment needs and at what versions. It is committed to git and shared across machines.

## File location

| Scope | Path |
|-------|------|
| User (default) | `~/.claude/apm.toml` |
| Project | `./.claude/apm.toml` |

## Full example

```toml
[plugin_manager]
scope = "user"   # "user" or "local"

[plugins]
"figma@claude-plugins-official"       = "^2.1.0"
"caveman@caveman"                     = "*"
"gopls-lsp@claude-plugins-official"  = "1.0.0"

[marketplaces]
"caveman" = { source = "github", repo = "JuliusBrussee/caveman" }
```

## `[plugin_manager]`

| Key | Values | Default | Description |
|-----|--------|---------|-------------|
| `scope` | `"user"`, `"local"` | `"user"` | Where plugins are installed. `user` = `~/.claude/`, `local` = `.claude/` in cwd. |

## `[plugins]`

Each entry is `"<name>@<marketplace>" = "<constraint>"`.

### Plugin identity

Plugin IDs follow the format `name@marketplace`:

```toml
"figma@claude-plugins-official" = "*"
#^^^^^  ^^^^^^^^^^^^^^^^^^^^^^^^
# name  marketplace ID
```

### Version constraints

| Constraint | Meaning |
|------------|---------|
| `"*"` or `"latest"` | Latest tag; falls back to HEAD if no tags |
| `"2.1.30"` | Exact semver tag |
| `"^2.1.0"` | Latest `2.x.x` ≥ `2.1.0` |
| `"~2.1.0"` | Latest `2.1.x` |
| `">=2.0.0"` | Any version ≥ `2.0.0` |
| `"ef6050c5e184"` | Exact git commit SHA |
| `"main"` | HEAD of `main` branch (resolved to SHA in lockfile) |

See [Version constraints](/guides/version-constraints/) for full details.

## `[marketplaces]`

Register additional marketplaces beyond `claude-plugins-official` (which is always available via `known_marketplaces.json`).

```toml
[marketplaces]
"my-mktplace" = { source = "github", repo = "org/repo" }
```

| Key | Description |
|-----|-------------|
| `source` | Currently only `"github"` |
| `repo` | `"owner/repo"` on GitHub |
