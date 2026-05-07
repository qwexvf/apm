---
title: ccpm add
description: Add a plugin to the manifest and install it.
sidebar:
  label: add
  order: 10
---

Add a plugin to `ccpm.toml`, resolve its version, download it, and update `ccpm.lock`.

## Usage

```sh
ccpm add <name@marketplace[@constraint]>
```

## Arguments

| Argument | Description |
|----------|-------------|
| `name@marketplace` | Plugin identity. `name` is the plugin name, `marketplace` is its source. |
| `@constraint` | Optional version constraint. Defaults to `*` (latest). |

## Examples

```sh
# latest version
ccpm add caveman@caveman

# semver range
ccpm add figma@claude-plugins-official@^2.1.0

# exact version
ccpm add gopls-lsp@claude-plugins-official@1.0.0

# pin to a commit SHA
ccpm add caveman@caveman@ef6050c5e184

# pin to a branch (resolved to SHA in lockfile)
ccpm add myplugin@mymarketplace@main
```

## What it does

1. Parses `name@marketplace[@constraint]`
2. Looks up the marketplace repo in `known_marketplaces.json` and `ccpm.toml`
3. Fetches tags from GitHub, resolves the latest matching version
4. Downloads and extracts to `~/.claude/plugins/cache/<marketplace>/<name>/<version>/`
5. Updates `installed_plugins.json` and `settings.json` in Claude Code's config dir
6. Appends to `ccpm.toml`
7. Writes a pinned entry to `ccpm.lock`

## Flags

| Flag | Description |
|------|-------------|
| `--global` | Install to user scope (`~/.claude/`) |
| `--local` | Install to project scope (`./.claude/`) |

## Notes

> [!NOTE]
> If no manifest exists yet, `ccpm add` creates one automatically before proceeding.

> [!TIP]
> Set `GITHUB_TOKEN` to avoid rate limiting when resolving versions.
