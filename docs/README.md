# astro-docs-template

A minimal, opinionated documentation site template. Astro 6 + React 19 + Tailwind v4 + shadcn/ui, with full-text search, dark/light theme, and frontmatter-driven sidebar nav.

[Use this template](https://github.com/qwexvf/astro-docs-template/generate) → click the button on GitHub, name your repo, clone, and you're shipping docs.

---

## What's in the box

- **Astro 6** static site, MDX support
- **React 19** islands for interactive bits (search, theme toggle, mobile sidebar)
- **Tailwind v4** via `@tailwindcss/vite`
- **shadcn/ui** + Radix primitives
- **Pagefind** full-text search, indexed at build time (no backend)
- **Shiki** syntax highlighting (`github-light` / `github-dark-dimmed`)
- **GitHub-style alerts** (`> [!NOTE]`, `> [!WARNING]`, …) via `remark-github-blockquote-alert`
- **Auto-linked headings** (`rehype-slug` + `rehype-autolink-headings`)
- **Scroll-spy TOC** on every doc page
- **Copy-on-click** code blocks
- **Prev / next** navigation derived from sidebar order
- **Edit on GitHub** link per page
- **GitHub Pages** workflow ready out of the box
- **Subpath deployments** via `PUBLIC_BASE`

## Quick start

```sh
# 1. Click "Use this template" on GitHub, or:
gh repo create my-docs --template qwexvf/astro-docs-template --public --clone
cd my-docs

# 2. Install + run
bun install
bun dev          # http://localhost:4321
```

> Requires **Node ≥ 22.12** (Astro 6) and **Bun** (or swap commands for npm/pnpm — the lockfile is `bun.lock` after first install).

## Project layout

```
.
├── astro.config.mjs            # site URL, base path, markdown plugins
├── components.json             # shadcn config
├── public/                     # static assets (favicon, og-image, …)
├── src/
│   ├── content/
│   │   └── docs/               # all your markdown / mdx lives here
│   │       ├── index.mdx
│   │       ├── getting-started.md
│   │       ├── guides/
│   │       ├── reference/
│   │       └── contributing/
│   ├── content.config.ts       # collection schema (frontmatter validation)
│   ├── components/
│   │   ├── Header.astro        # top bar — brand, nav links, search, theme
│   │   ├── Sidebar.astro       # desktop left sidebar
│   │   ├── MobileSidebar.tsx   # drawer for mobile
│   │   ├── Search.tsx          # cmdk-driven Pagefind search
│   │   ├── ThemeToggle.tsx
│   │   └── ui/                 # shadcn primitives
│   ├── layouts/
│   │   ├── BaseLayout.astro    # html shell, theme init, copy-button script
│   │   └── DocsLayout.astro    # sidebar + article + prev/next + TOC
│   ├── lib/
│   │   ├── nav.ts              # buildNav() — section/order from frontmatter
│   │   └── utils.ts
│   ├── pages/
│   │   ├── index.astro         # landing page
│   │   └── [...slug].astro     # renders every doc collection entry
│   └── styles/
│       ├── globals.css         # tokens, theme vars, base
│       └── prose.css           # `.prose-docs` typographic styles
└── .github/workflows/
    ├── build.yml               # PR + push build check
    └── pages.yml               # deploy to GitHub Pages
```

## Authoring

### Add a page

Drop a `.md` or `.mdx` file in `src/content/docs/`. It becomes a route automatically — `src/content/docs/guides/cookbook.md` → `/guides/cookbook/`.

```yaml
---
title: Page title             # required — shown as h1 and in <title>
description: Short summary.   # optional — shown under the title and in <meta>
sidebar:
  label: Optional label       # optional — falls back to title
  order: 1                    # optional — lower = higher in sidebar (default 999)
  hidden: false               # optional — set true to omit from nav
---
```

The file body is plain markdown / MDX. GitHub-style alerts work:

```md
> [!NOTE]
> Useful information.

> [!WARNING]
> Pay attention.
```

### Sections

Sections are derived from the directory:

| Directory                     | Section in sidebar |
| ----------------------------- | ------------------ |
| `src/content/docs/*.md`       | Start here         |
| `src/content/docs/guides/*`   | Guides             |
| `src/content/docs/reference/*`| Reference          |
| `src/content/docs/contributing/*` | Contributing  |

To add or rename sections, edit the `SECTIONS` array in [`src/lib/nav.ts`](src/lib/nav.ts).

### Ordering

Pages within a section are sorted by `sidebar.order` ascending, then alphabetically. Use sparse numbers (10, 20, 30) so you can insert later.

## Customizing

Search-and-replace these placeholders before your first commit:

| Placeholder              | Where                                              | What to set                |
| ------------------------ | -------------------------------------------------- | -------------------------- |
| `https://example.com`    | `astro.config.mjs` (`site:`)                       | Production URL             |
| `your-org/your-repo`     | `Header.astro`, `DocsLayout.astro`                 | GitHub org/repo            |
| `Docs`                   | `Header.astro` brand, `BaseLayout.astro` title     | Project name               |
| `Project`                | `src/pages/index.astro`                            | Project name + tagline     |
| Logo SVG                 | `Header.astro`                                     | Your logo                  |
| Theme tokens             | `src/styles/globals.css`                           | Brand colors               |
| `docs-theme`             | `BaseLayout.astro`, `ThemeToggle.tsx`              | localStorage key (only if it would collide with another app on the same domain) |

### Theme tokens

All colors live in `src/styles/globals.css` as CSS variables, with `[data-theme='dark']` and `[data-theme='light']` blocks. Change `--color-signal` to rebrand the accent.

### shadcn

`components.json` is configured. Add new components with:

```sh
bunx shadcn@latest add <component>
```

## Deploying

### GitHub Pages

Push to `main`. The included [`pages.yml`](.github/workflows/pages.yml) workflow:

1. Builds with `PUBLIC_BASE=/<repo-name>/` so subpath URLs work.
2. Uploads `dist/`.
3. Deploys to Pages.

Enable Pages in repo Settings → Pages → Source: **GitHub Actions**.

### Custom domain or root path

Set `site:` in `astro.config.mjs` to your URL. Don't set `PUBLIC_BASE`. Drop a `CNAME` file into `public/` if using a custom domain.

### Anywhere else (Vercel, Netlify, Cloudflare Pages, …)

Static output. Build command: `bun run build`. Publish dir: `dist`.

## Scripts

| Command         | What it does                                |
| --------------- | ------------------------------------------- |
| `bun dev`       | Dev server at `localhost:4321`              |
| `bun build`     | Production build (runs Pagefind after)      |
| `bun preview`   | Serve `dist/` locally                       |
| `bun astro …`   | Astro CLI passthrough                       |

## FAQ

**Why Astro and not Starlight / Nextra / Docusaurus?**
Starlight is great but opinionated about chrome and theming. Nextra is Next-bound. Docusaurus is heavy. This template gives you a small, hackable surface area — every component is in `src/`, no framework abstraction to fight.

**Can I use npm or pnpm?**
Yes. Delete `bun.lock` and run `npm install` or `pnpm install`. Update the workflows accordingly.

**How do I add a non-doc page (e.g. `/blog/`)?**
Add an `.astro` file under `src/pages/`. It bypasses the docs collection.

**Can I have versioned docs?**
Not built in. The simplest approach is a separate deploy per version branch, served from subdirectories.

## License

MIT — see [LICENSE](LICENSE).
