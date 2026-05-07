---
title: ccpm sync
description: Repair Claude Code state from ccpm.lock without re-downloading.
sidebar:
  label: sync
  order: 90
---

Write plugin entries into Claude Code's `installed_plugins.json` and `settings.json` based on `ccpm.lock`, without downloading anything.

## Usage

```sh
ccpm sync
```

## When to use

- Claude Code lost track of installed plugins (registry got corrupted or reset)
- You copied plugin files manually and need Claude Code to recognize them
- After a Claude Code update that reset `settings.json`

## What it does

For each entry in `ccpm.lock`:
1. Checks that the `install_path` exists on disk
2. Writes the entry to `installed_plugins.json`
3. Sets `enabledPlugins[id] = true` in `settings.json`

If an `install_path` is missing, it prints a warning and skips that plugin — run `ccpm install` to re-download those.

> [!TIP]
> `ccpm sync` is a repair tool, not a substitute for `ccpm install`. Use `install` when you need the files; use `sync` when the files exist but Claude Code doesn't know about them.
