---
title: apm add
description: Add a plugin to the manifest and install it.
sidebar:
  label: add
  order: 10
---

Add a plugin to `apm.toml`, resolve its version, download it, and update `apm.lock`.

## Usage

```sh
apm add <name@marketplace[@constraint]>
```

## Arguments

| Argument | Description |
|----------|-------------|
| `name@marketplace` | Plugin identity. `name` is the plugin name, `marketplace` is its source. |
| `@constraint` | Optional version constraint. Defaults to `*` (latest). |

## Examples

```sh
# latest version
apm add caveman@caveman

# semver range
apm add figma@claude-plugins-official@^2.1.0

# exact version
apm add gopls-lsp@claude-plugins-official@1.0.0

# pin to a commit SHA
apm add caveman@caveman@ef6050c5e184

# pin to a branch (resolved to SHA in lockfile)
apm add myplugin@mymarketplace@main
```

## What it does

1. Parses `name@marketplace[@constraint]`
2. Looks up the marketplace repo in `known_marketplaces.json` and `apm.toml`
3. Fetches tags from GitHub, resolves the latest matching version
4. Downloads and extracts to `~/.claude/plugins/cache/<marketplace>/<name>/<version>/`
5. Updates `installed_plugins.json` and `settings.json` in Claude Code's config dir
6. Appends to `apm.toml`
7. Writes a pinned entry to `apm.lock`

## Flags

| Flag | Description |
|------|-------------|
| `--global` | Install to user scope (`~/.claude/`) |
| `--local` | Install to project scope (`./.claude/`) |

## Notes

> [!NOTE]
> If no manifest exists yet, `apm add` creates one automatically before proceeding.

> [!TIP]
> Set `GITHUB_TOKEN` to avoid rate limiting when resolving versions.
