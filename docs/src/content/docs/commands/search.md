---
title: ccpm search
description: Search available plugins across registered marketplaces.
sidebar:
  label: search
  order: 60
---

Search plugins by name or description across all registered marketplaces.

## Usage

```sh
ccpm search [query]
```

## Examples

```sh
# list all available plugins
ccpm search

# filter by keyword
ccpm search figma
ccpm search lsp
ccpm search "code review"
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
> `ccpm search` reads from locally cloned marketplace repos. Run `ccpm marketplace update` first to get the latest plugin listings.
