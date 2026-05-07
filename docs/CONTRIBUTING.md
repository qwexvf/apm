# Contributing

Thanks for considering a contribution.

## Scope

This is a **template**, not a framework. The goal is to stay small, hackable, and free of magic. Good contributions:

- Fix bugs in the template itself.
- Improve the build / DX (faster cold start, better defaults, clearer config).
- Improve docs and examples.

Out of scope:

- Heavy theme variants — fork instead.
- Locking us into a specific deploy target.
- Plugins for niche markdown extensions — add them in your own fork.

## Dev loop

```sh
bun install
bun dev
bun run build    # must pass before opening a PR
```

CI runs `bun run build` on every PR.

## PRs

- One concern per PR.
- Update the README if you change behavior or add a feature.
- Match the existing code style (Prettier defaults, no extra config).

## Reporting bugs

Open an issue with:

- What you ran (`bun create …`, OS, Node version)
- What happened vs. what you expected
- A minimal repro if possible
