---
title: Plugin format
description: The Claude Code plugin directory structure that apm installs.
sidebar:
  label: Plugin format
  order: 30
---

apm installs plugins from GitHub repos. A valid Claude Code plugin is any repo (or directory within a marketplace) that contains a `.claude-plugin/plugin.json` file.

## Minimal plugin

```
my-plugin/
└── .claude-plugin/
    └── plugin.json
```

```json
{
  "name": "my-plugin",
  "description": "What it does",
  "author": "yourname",
  "version": "1.0.0"
}
```

## Full plugin structure

```
my-plugin/
├── .claude-plugin/
│   ├── plugin.json       # required
│   └── marketplace.json  # optional, for marketplace listing
├── skills/               # markdown skill files
│   └── my-skill/
│       └── SKILL.md
├── agents/               # agent definition files
│   └── my-agent.md
├── hooks/                # Node.js hook scripts (CommonJS)
│   ├── package.json      # {"type": "commonjs"} required
│   └── my-hook.js
├── mcp-servers/          # MCP server implementations
│   └── my-server/
└── CLAUDE.md             # shown as plugin documentation
```

## plugin.json fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | Plugin identifier |
| `description` | string | no | Short description |
| `author` | string | no | Author name |
| `version` | string | no | Version (apm uses git tags preferentially) |

## hooks/package.json

If your plugin includes hooks, you **must** include a `package.json` with `"type": "commonjs"` in the `hooks/` directory:

```json
{ "type": "commonjs" }
```

Without this, Node.js treats `.js` files as ESM and `require()` calls will fail.

## Install path

apm installs plugins to:

```
~/.claude/plugins/cache/<marketplace-id>/<plugin-name>/<version>/
```

Example:
```
~/.claude/plugins/cache/caveman/caveman/ef6050c5e184/
```
