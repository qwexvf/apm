---
title: apm init
description: Scaffold a apm.toml manifest.
sidebar:
  label: init
  order: 5
---

Create a `apm.toml` manifest file with the official marketplace pre-configured.

## Usage

```sh
apm init           # user scope → ~/.claude/apm.toml
apm init --local   # project scope → ./.claude/apm.toml
```

## Output

Creates `apm.toml`:

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
> `apm init` fails if `apm.toml` already exists at the target path.
