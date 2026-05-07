---
title: ccpm lock
description: Regenerate ccpm.lock from the manifest constraints.
sidebar:
  label: lock
  order: 80
---

Re-resolve all plugins in `ccpm.toml` and write a fresh `ccpm.lock`. Use this after manually editing `ccpm.toml`.

## Usage

```sh
ccpm lock
```

## What it does

For each plugin in `ccpm.toml`:
1. Fetches tags from GitHub
2. Resolves the latest version matching the constraint
3. Writes the exact version + commit SHA to `ccpm.lock`

> [!NOTE]
> `ccpm lock` does **not** download or install plugins. Run `ccpm install` after to actually apply the new lockfile.
