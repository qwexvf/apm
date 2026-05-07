---
title: Quick start
description: Install your first plugin with ccpm in under 2 minutes.
sidebar:
  order: 2
---

## 1. Initialize a manifest

```sh
ccpm init
```

Creates `~/.claude/ccpm.toml` (user scope). For a project-specific manifest:

```sh
ccpm init --local    # creates ./.claude/ccpm.toml
```

## 2. Add a plugin

```sh
ccpm add caveman@caveman
```

ccpm resolves the latest version, downloads it, and writes:
- the plugin files to `~/.claude/plugins/cache/`
- an entry in `ccpm.toml`
- a pinned entry in `ccpm.lock`
- the plugin into Claude Code's `installed_plugins.json` and `settings.json`

## 3. Add with a version constraint

```sh
ccpm add figma@claude-plugins-official@^2.1.0
```

Installs the latest `2.x.x` ≥ `2.1.0`. The constraint is saved in `ccpm.toml`; the resolved exact version goes into `ccpm.lock`.

## 4. Commit both files

```sh
git add ccpm.toml ccpm.lock
git commit -m "add claude code plugins"
```

## 5. Reproduce on another machine

```sh
ccpm install
```

Reads `ccpm.lock` and installs the exact same versions. No network resolution needed for already-cached plugins.

## Next

- `ccpm list` — see what's installed
- `ccpm update` — bump to latest matching constraints
- `ccpm search` — browse available plugins
