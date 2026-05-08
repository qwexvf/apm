---
title: Creating a marketplace
description: How to publish plugins as a apm-compatible marketplace.
sidebar:
  label: Creating a marketplace
  order: 40
---

A marketplace is a public GitHub repo with a specific directory structure. Once registered, users can install your plugins with `apm add <plugin>@<your-marketplace>`.

## Directory structure

```
my-marketplace/
├── .claude-plugin/
│   └── plugin.json          # marketplace metadata
└── plugins/                 # one directory per plugin
    ├── my-first-plugin/
    │   ├── .claude-plugin/
    │   │   └── plugin.json  # plugin metadata
    │   ├── skills/
    │   │   └── my-skill/
    │   │       └── SKILL.md
    │   └── README.md
    └── my-second-plugin/
        └── ...
```

## Marketplace `plugin.json`

```json
{
  "name": "my-marketplace",
  "description": "My collection of Claude Code plugins",
  "author": "yourname"
}
```

## Plugin `plugin.json`

```json
{
  "name": "my-first-plugin",
  "description": "What this plugin does",
  "author": "yourname",
  "version": "1.0.0"
}
```

## Plugin types

Plugins can include any combination of:

| Directory | Purpose |
|-----------|---------|
| `skills/` | Markdown skill files loaded by Claude Code |
| `agents/` | Agent definition files |
| `hooks/` | Node.js hook scripts |
| `mcp-servers/` | MCP server implementations |

## Registering with users

Users register your marketplace:

```sh
apm marketplace add my-marketplace github:yourname/my-marketplace
```

Then install plugins:

```sh
apm add my-first-plugin@my-marketplace
```

## Versioning

apm supports semver tags for your plugins. To release version `1.2.0`:

```sh
git tag v1.2.0
git push origin v1.2.0
```

Users with constraint `^1.0.0` will get `1.2.0` on the next `apm update`.

If you don't tag releases, apm falls back to commit SHAs — users get HEAD of the default branch when they install with `*`.
