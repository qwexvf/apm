---
title: apm add
description: Add a plugin or skill to the manifest and install it.
sidebar:
  label: add
  order: 10
---

Add a plugin or skill to `apm.toml`, resolve its version, download it, and update `apm.lock`.

`apm add` handles two kinds of dependency and picks the right one automatically:

- **Plugins** come from a registered marketplace: `name@marketplace`.
- **Skills** come from any GitHub repo: `name@owner/repo[:subpath]`.

The second `@`-segment disambiguates: if it contains a `/` it's a repo (skill), otherwise it's a marketplace ID (plugin).

## Usage

```sh
apm add <name@marketplace[@constraint]>              # plugin
apm add <name@owner/repo[:subpath][@constraint]>     # skill
```

## Arguments

| Argument | Description |
|----------|-------------|
| `name@marketplace` | Plugin identity. `name` is the plugin name, `marketplace` is a registered marketplace ID (no `/`). |
| `name@owner/repo` | Skill identity. `owner/repo` is a GitHub repo. |
| `:subpath` | Optional. Directory inside the repo that holds `SKILL.md`, when the skill isn't at the repo root. |
| `@constraint` | Optional version constraint. Defaults to `*` (latest). Applies to both kinds. |

## Examples

### Plugins

```sh
# latest version
apm add caveman@caveman

# semver range
apm add figma@claude-plugins-official@^2.1.0

# exact version
apm add gopls-lsp@claude-plugins-official@1.0.0

# pin to a commit SHA
apm add caveman@caveman@ef6050c5e184

# pin to a branch (resolved to SHA in lockfile)
apm add myplugin@mymarketplace@main
```

### Skills

```sh
# skill in a subdirectory of a repo
apm add frontend-design@vercel-labs/skills:skills/frontend-design

# a repo whose SKILL.md sits under skills/<name>/
apm add hallmark@Nutlope/hallmark:skills/hallmark

# skill at the repo root (no subpath)
apm add my-skill@me/my-skill-repo

# pin a skill to a branch
apm add graphify@qwexvf/dotfiles:.claude/skills/graphify@main
```

## What it does

**For plugins:**

1. Parses `name@marketplace[@constraint]`
2. Looks up the marketplace repo in `known_marketplaces.json` and `apm.toml`
3. Fetches tags from GitHub, resolves the latest matching version
4. Downloads and extracts to `~/.claude/plugins/cache/<marketplace>/<name>/<version>/`
5. Updates `installed_plugins.json` and `settings.json` in Claude Code's config dir
6. Appends to the `[plugins]` table in `apm.toml`
7. Writes a pinned entry to `apm.lock`

**For skills:**

1. Parses `name@owner/repo[:subpath][@constraint]`
2. Resolves the constraint against the repo's tags (or a branch/commit SHA)
3. Downloads the repo tarball, verifies `SKILL.md` exists at `:subpath` (or root)
4. Moves the skill directory into `<claudeDir>/skills/<name>/`
5. Appends to the `[skills]` table in `apm.toml`
6. Writes a pinned entry (with integrity hash) to `apm.lock`

Repos without tags resolve to a commit SHA — that becomes the pinned version.

## Flags

| Flag | Description |
|------|-------------|
| `--global` | Install to user scope (`~/.claude/`) |
| `--local` | Install to project scope (`./.claude/`) |

## Notes

> [!NOTE]
> If no manifest exists yet, `apm add` creates one automatically before proceeding.

> [!TIP]
> Set `GITHUB_TOKEN` to avoid rate limiting when resolving versions.

> [!TIP]
> When a skill's `SKILL.md` isn't at the repo root, point `:subpath` at the directory that contains it. `apm` errors out with `no SKILL.md at <subpath>` if it can't find one.

See the [Skills guide](/guides/skills/) for the full workflow.
