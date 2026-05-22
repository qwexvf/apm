# apm — Agent Plugin Manager

Lockfile-based plugin manager for Claude Code (and AI agent tools).

Claude Code's built-in plugin system has no lockfile, no version constraints, no reproducibility. `apm` fixes that.

```sh
apm add figma@claude-plugins-official@^2.1.0
apm add caveman@caveman
apm add frontend-design@vercel-labs/skills:skills/frontend-design   # a SKILL.md
git add apm.toml apm.lock
```

On a new machine:

```sh
apm install   # exact same versions, every time
```

## Why

I built this after hitting three problems with Claude Code's native plugin system:

**1. No per-project version control.** I work across multiple projects that need different plugin versions. Claude Code installs plugins globally with no way to pin versions per repo. One project upgrading figma would silently break another.

**2. No visibility into what's installed or where.** I had no idea which version of a plugin was active, or where it was installed on disk. `apm info <plugin>` shows you the exact version, commit SHA, and install path. `apm list` shows all plugins and their locked versions at a glance.

**3. Plugins aren't reproducible across machines.** Onboarding a new dev or spinning up CI meant manually reinstalling the same plugins and hoping you picked the right versions. With `apm`, you commit `apm.toml` + `apm.lock` and run `apm install` — same versions, every time, everywhere.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/qwexvf/apm/main/install.sh | sh
```

Or with Go:

```sh
go install github.com/qwexvf/apm@latest
```

## Commands

| Command | Description |
|---------|-------------|
| `apm init` | Scaffold `apm.toml` |
| `apm add <plugin[@constraint]>` | Add + install + lock (plugin or skill) |
| `apm remove <plugin>` | Remove + uninstall |
| `apm scaffold skill <name>` | Write a starter `SKILL.md` in `./skills/<name>/` |
| `apm install` | Deterministic install from lockfile |
| `apm update [plugin]` | Bump to latest matching constraint |
| `apm list` | Show installed plugins and locked versions |
| `apm search [query]` | Search marketplace |
| `apm info <plugin>` | Plugin details, version, and install path |
| `apm lock` | Regenerate lockfile from manifest |
| `apm sync` | Repair Claude Code state from lockfile |
| `apm marketplace add/list/update` | Manage marketplaces |

## Version constraints

```toml
# apm.toml
[plugins]
"figma@claude-plugins-official"     = "^2.1.0"   # semver range
"caveman@caveman"                   = "*"          # latest
"gopls-lsp@claude-plugins-official" = "1.0.0"     # exact
"myplugin@mymarket"                 = "abc123"     # commit SHA

[skills]
"frontend-design@vercel-labs/skills:skills/frontend-design" = "*"
"graphify@qwexvf/dotfiles:.claude/skills/graphify"          = "main"
```

## Skills

`apm` installs standalone `SKILL.md` files alongside full plugins. Source is
any GitHub repo; an optional `:subpath` selects a subdirectory inside the
repo when the skill lives there.

```sh
# install from any GitHub repo
apm add frontend-design@vercel-labs/skills:skills/frontend-design

# scaffold a new skill in this repo
apm scaffold skill my-skill   # writes ./skills/my-skill/SKILL.md

# remove
apm remove frontend-design@vercel-labs/skills:skills/frontend-design
```

Skills land in `<claudeDir>/skills/<name>/`. Version constraints, lockfile
pinning, and `apm install` behave the same as for plugins.

## GitHub token

Without a token: 60 API requests/hour. With one: 5,000.

```sh
export GITHUB_TOKEN=ghp_...
```

Only needs public repo read access.

## License

MIT
