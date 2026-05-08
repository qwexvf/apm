---
title: Version constraints
description: How apm resolves semver ranges, commit SHAs, and branch refs.
sidebar:
  label: Version constraints
  order: 30
---

apm supports four kinds of version specifiers. The kind is auto-detected from the string format.

## Semver ranges

Plugins with tagged releases support standard semver constraints:

| Constraint | Resolves to |
|------------|-------------|
| `"1.0.0"` | Exactly `1.0.0` |
| `"^2.1.0"` | Latest `2.x.x` where `x.x ≥ 1.0` |
| `"~2.1.0"` | Latest `2.1.x` |
| `">=2.0.0 <3.0.0"` | Any `2.x.x` |
| `"*"` or `"latest"` | Highest tag using semver sort |

The constraint is saved in `apm.toml`; the resolved exact version is saved in `apm.lock`.

## Git commit SHAs

Pin to an exact commit — no tag required:

```toml
"caveman@caveman" = "ef6050c5e184"   # short (7-40 hex chars)
```

The full SHA is resolved and stored in the lockfile. Useful for plugins that don't publish tags.

## Branch refs

Pin to the current HEAD of a branch:

```toml
"myplugin@mymarketplace" = "main"
```

> [!WARNING]
> Branch refs are resolved to a commit SHA at install time. The SHA is stored in `apm.lock`, so subsequent `apm install` calls are still reproducible. But `apm update` will move to the new HEAD — use a SHA if you need a harder pin.

## Latest (default)

```toml
"caveman@caveman" = "*"
```

Strategy:
1. Fetch all tags
2. Find the highest semver tag — use it
3. If no semver tags: use the first tag
4. If no tags at all: use HEAD commit SHA

## Precedence

When resolving, apm detects constraint kind in this order:

1. `*` or `latest` → latest
2. 7–40 hex chars → commit SHA
3. Semver version → exact match
4. Semver constraint operators (`^`, `~`, `>=`, etc.) → range
5. Anything else → branch/ref name
