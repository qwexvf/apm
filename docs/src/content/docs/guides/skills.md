---
title: Skills
description: Install, pin, and reproduce standalone SKILL.md files from any GitHub repo.
sidebar:
  label: Skills
  order: 45
---

A **skill** is a single `SKILL.md` file (plus any supporting `references/`) that
teaches an agent a focused capability. Unlike plugins, skills don't need a
marketplace — `apm` installs them straight from any GitHub repo, with the same
lockfile pinning and reproducibility you get for plugins.

## Add a skill

```sh
apm add <name>@<owner/repo>[:subpath][@constraint]
```

- `name` — the local skill name; it lands in `<claudeDir>/skills/<name>/`.
- `owner/repo` — any GitHub repo (the `/` is what tells `apm` this is a skill, not a plugin).
- `:subpath` — the directory inside the repo that contains `SKILL.md`, when it isn't at the repo root.
- `@constraint` — optional version constraint; defaults to `*` (latest).

```sh
# skill nested under skills/<name>/
apm add hallmark@Nutlope/hallmark:skills/hallmark

# skill from a shared collection repo
apm add frontend-design@vercel-labs/skills:skills/frontend-design

# skill at the repo root
apm add my-skill@me/my-skill-repo
```

## Where subpath points

`apm` looks for `SKILL.md` at the `:subpath` you give (or the repo root if you
omit it). Most repos keep skills under `skills/<name>/`, so the subpath usually
mirrors that layout:

```
Nutlope/hallmark
└── skills/
    └── hallmark/
        ├── SKILL.md          ← :subpath = skills/hallmark
        └── references/
```

If the path is wrong you'll get:

```
no SKILL.md at <subpath> (set subpath to the skill's directory)
```

## Versioning

Skills resolve the same way as plugins:

| Constraint | Resolves to |
|------------|-------------|
| `*` | latest tag, or the default-branch commit if the repo has no tags |
| `^1.2.0` | latest tag matching the semver range |
| `1.2.0` | that exact tag |
| `main` | the branch tip, pinned to its commit SHA in the lockfile |
| `aeb42fb` | that exact commit SHA |

Repos without tags (like most single-skill repos) pin to a commit SHA — that
SHA becomes the recorded version.

## Manifest and lockfile

Skills live in their own `[skills]` table in `apm.toml`:

```toml
[skills]
"hallmark@Nutlope/hallmark:skills/hallmark"                 = "*"
"frontend-design@vercel-labs/skills:skills/frontend-design" = "*"
"graphify@qwexvf/dotfiles:.claude/skills/graphify"          = "main"
```

Each gets a pinned entry in `apm.lock` with the resolved version, commit SHA,
install path, and an integrity hash of the downloaded tarball:

```toml
[[skills]]
  id = "hallmark@Nutlope/hallmark:skills/hallmark"
  version = "aeb42fb354ff"
  commit_sha = "aeb42fb354ff4efa36ab475773a082315a3af2ce"
  resolved_url = "https://github.com/Nutlope/hallmark"
  install_path = "/home/you/.claude/skills/hallmark"
  integrity = "sha256:006bfb8b..."
```

Commit `apm.toml` + `apm.lock`, and `apm install` reproduces the exact same
skills on any machine.

## Scaffold your own

```sh
apm scaffold skill my-skill   # writes ./skills/my-skill/SKILL.md
```

Push it to a repo, then anyone can `apm add my-skill@you/your-repo:skills/my-skill`.

## Remove

```sh
apm remove hallmark@Nutlope/hallmark:skills/hallmark
```

This drops the `[skills]` entry, the lockfile entry, and deletes
`<claudeDir>/skills/<name>/`.

## Scope

Like plugins, skills install to whichever scope you target:

```sh
apm --global add hallmark@Nutlope/hallmark:skills/hallmark   # ~/.claude/skills/
apm --local  add hallmark@Nutlope/hallmark:skills/hallmark   # ./.claude/skills/
```
