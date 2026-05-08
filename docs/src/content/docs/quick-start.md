---
title: Quick start
description: Install your first plugin with apm in under 2 minutes.
sidebar:
  order: 2
---

## 1. Initialize a manifest

```sh
apm init
```

Creates `~/.claude/apm.toml` (user scope). For a project-specific manifest:

```sh
apm init --local    # creates ./.claude/apm.toml
```

## 2. Add a plugin

```sh
apm add caveman@caveman
```

apm resolves the latest version, downloads it, and writes:
- the plugin files to `~/.claude/plugins/cache/`
- an entry in `apm.toml`
- a pinned entry in `apm.lock`
- the plugin into Claude Code's `installed_plugins.json` and `settings.json`

## 3. Add with a version constraint

```sh
apm add figma@claude-plugins-official@^2.1.0
```

Installs the latest `2.x.x` ≥ `2.1.0`. The constraint is saved in `apm.toml`; the resolved exact version goes into `apm.lock`.

## 4. Commit both files

```sh
git add apm.toml apm.lock
git commit -m "add claude code plugins"
```

## 5. Reproduce on another machine

```sh
apm install
```

Reads `apm.lock` and installs the exact same versions. No network resolution needed for already-cached plugins.

## Next

- `apm list` — see what's installed
- `apm update` — bump to latest matching constraints
- `apm search` — browse available plugins
