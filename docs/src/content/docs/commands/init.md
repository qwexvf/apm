---
title: ccpm init
description: Scaffold a ccpm.toml manifest.
sidebar:
  label: init
  order: 5
---

Create a `ccpm.toml` manifest file with the official marketplace pre-configured.

## Usage

```sh
ccpm init           # user scope → ~/.claude/ccpm.toml
ccpm init --local   # project scope → ./.claude/ccpm.toml
```

## Output

Creates `ccpm.toml`:

```toml
[plugin_manager]
scope = "user"

[plugins]

[marketplaces]
"claude-plugins-official" = { source = "github", repo = "anthropics/claude-plugins-official" }
```

## Flags

| Flag | Description |
|------|-------------|
| `--local` | Create in `./.claude/` (project scope) |
| `--global` | Create in `~/.claude/` (user scope, default) |

> [!NOTE]
> `ccpm init` fails if `ccpm.toml` already exists at the target path.
