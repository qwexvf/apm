---
title: ccpm marketplace
description: Register and manage plugin marketplaces.
sidebar:
  label: marketplace
  order: 100
---

Manage the list of plugin marketplaces ccpm knows about.

## Subcommands

### `marketplace add`

Register a new marketplace and clone its repo locally.

```sh
ccpm marketplace add <id> <github:owner/repo>
```

**Example:**

```sh
ccpm marketplace add caveman github:JuliusBrussee/caveman
```

This:
1. Adds the marketplace to `known_marketplaces.json`
2. Adds it to `ccpm.toml` `[marketplaces]`
3. Clones the repo to `~/.claude/plugins/marketplaces/caveman/`

### `marketplace list`

List all registered marketplaces.

```sh
ccpm marketplace list
```

```
ID                        REPO                                  LOCATION
claude-plugins-official   github.com/anthropics/claude-plugins-official  ~/.claude/plugins/marketplaces/claude-plugins-official
caveman                   github.com/JuliusBrussee/caveman               ~/.claude/plugins/marketplaces/caveman
```

### `marketplace update`

Pull the latest plugin listings from all registered marketplaces.

```sh
ccpm marketplace update           # update all
ccpm marketplace update caveman   # update one
```

> [!NOTE]
> Run this periodically to see new plugins in `ccpm search`.

## How marketplaces work

A marketplace is a GitHub repo with a specific structure:

```
my-marketplace/
├── .claude-plugin/
│   └── plugin.json      # marketplace metadata
└── plugins/             # one dir per plugin
    └── myplugin/
        └── .claude-plugin/
            └── plugin.json
```

See [Creating a marketplace](/guides/marketplaces/) for details.
