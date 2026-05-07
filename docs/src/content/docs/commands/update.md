---
title: ccpm update
description: Update plugins to the latest version matching their constraint.
sidebar:
  label: update
  order: 30
---

Resolve the latest version satisfying each plugin's constraint, show a diff, and optionally apply.

## Usage

```sh
# update all plugins
ccpm update

# update one plugin
ccpm update caveman@caveman

# preview without applying
ccpm update --dry-run

# apply without prompt
ccpm update --yes
```

## Flags

| Flag | Description |
|------|-------------|
| `--dry-run` | Show available updates without downloading or modifying files |
| `-y, --yes` | Skip confirmation prompt |

## Example output

```
updates available:
  caveman@caveman           ef6050c5 → a1b2c3d4
  figma@claude-plugins-official  2.1.30 → 2.2.0

apply updates? [y/N]:
```

## What changes

When updates are applied:
- New plugin versions are downloaded to the cache
- `installed_plugins.json` and `settings.json` are updated
- `ccpm.lock` is updated to the new versions

`ccpm.toml` constraints are **not** changed — only the lockfile is updated to the new resolved version within those constraints.

> [!TIP]
> To change a constraint (e.g. from `^2.1.0` to `^3.0.0`), edit `ccpm.toml` manually then run `ccpm update`.
