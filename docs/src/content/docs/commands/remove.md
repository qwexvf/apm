---
title: ccpm remove
description: Remove a plugin from the manifest and uninstall it.
sidebar:
  label: remove
  order: 40
---

Remove a plugin from `ccpm.toml`, delete its files, and clean up Claude Code's state.

## Usage

```sh
ccpm remove <name@marketplace>
```

## Example

```sh
ccpm remove caveman@caveman
```

## What it does

1. Reads the install path from `ccpm.lock`
2. Deletes the plugin's directory under `~/.claude/plugins/cache/`
3. Removes the entry from `installed_plugins.json`
4. Removes the `enabledPlugins` entry from `settings.json`
5. Removes the plugin from `ccpm.toml`
6. Removes the entry from `ccpm.lock`
