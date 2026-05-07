---
title: ccpm install
description: Install all plugins from ccpm.lock deterministically.
sidebar:
  label: install
  order: 20
---

Install every plugin listed in `ccpm.lock` at the exact pinned version. Equivalent to `npm ci` — no resolution, no surprises.

## Usage

```sh
ccpm install
```

## What it does

1. Reads `ccpm.lock`
2. For each plugin, checks if the install path already exists
3. Skips plugins that are already cached (by path)
4. Downloads missing plugins from GitHub
5. Updates `installed_plugins.json` and `settings.json` in Claude Code

## When to use

- After cloning a repo that has `ccpm.toml` + `ccpm.lock`
- On a new machine to reproduce the exact plugin state your team uses
- In CI to verify plugins install cleanly

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | All plugins installed successfully |
| `1` | One or more plugins failed — details printed to stderr |

## Flags

| Flag | Description |
|------|-------------|
| `--global` | Read from `~/.claude/ccpm.lock` |
| `--local` | Read from `./.claude/ccpm.lock` |

> [!NOTE]
> `ccpm install` never modifies `ccpm.toml` or `ccpm.lock`. Use `ccpm add` or `ccpm update` to change versions.
