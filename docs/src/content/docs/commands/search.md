---
title: apm search
description: Search available plugins across registered marketplaces.
sidebar:
  label: search
  order: 60
---

Search plugins by name or description across all registered marketplaces.

## Usage

```sh
apm search [query]
```

## Examples

```sh
# list all available plugins
apm search

# filter by keyword
apm search figma
apm search lsp
apm search "code review"
```

## Example output

```
PLUGIN ID                              DESCRIPTION                          AUTHOR
figma@claude-plugins-official          Figma design-to-code integration     Figma
gopls-lsp@claude-plugins-official      Go language server (gopls)           Anthropic
caveman@caveman                        Ultra-compressed communication mode  JuliusBrussee
frontend-design@claude-plugins-official  Production-grade frontend UI gen   Anthropic
```

> [!NOTE]
> `apm search` reads from locally cloned marketplace repos. Run `apm marketplace update` first to get the latest plugin listings.
