---
title: ccpm info
description: Show plugin details and install status.
sidebar:
  label: info
  order: 70
---

Show detailed information about a plugin: description, author, available version, locked version, and install path.

## Usage

```sh
ccpm info <name@marketplace>
```

## Example

```sh
ccpm info figma@claude-plugins-official
```

```
name:        figma
id:          figma@claude-plugins-official
description: Figma design-to-code integration — get_design_context, Code Connect, FigJam diagrams
author:      Figma
marketplace: claude-plugins-official  (github.com/anthropics/claude-plugins-official)
version:     2.2.0
locked:      2.1.30  (3590366424de)
installed:   2.1.30
path:        /home/user/.claude/plugins/cache/claude-plugins-official/figma/2.1.30
```
