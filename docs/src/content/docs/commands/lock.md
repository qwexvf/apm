---
title: apm lock
description: Regenerate apm.lock from the manifest constraints.
sidebar:
  label: lock
  order: 80
---

Re-resolve all plugins in `apm.toml` and write a fresh `apm.lock`. Use this after manually editing `apm.toml`.

## Usage

```sh
apm lock
```

## What it does

For each plugin in `apm.toml`:
1. Fetches tags from GitHub
2. Resolves the latest version matching the constraint
3. Writes the exact version + commit SHA to `apm.lock`

> [!NOTE]
> `apm lock` does **not** download or install plugins. Run `apm install` after to actually apply the new lockfile.
