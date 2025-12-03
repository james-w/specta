# specta Documentation

This directory contains the Hugo-based documentation site for specta.

## Viewing Documentation Online

The documentation is automatically deployed to GitHub Pages:

**https://james-w.github.io/specta/**

## Local Development

### Prerequisites

- [Hugo](https://gohugo.io/installation/) (Extended version, v0.128.0 or later)
- Go 1.22+

### Running Locally

1. Install Hugo modules:
   ```bash
   cd docs
   hugo mod init github.com/james-w/specta/docs
   hugo mod get
   ```

2. Start the development server:
   ```bash
   hugo server -D
   ```

3. Open http://localhost:1313 in your browser

The site will automatically reload when you make changes to content files.

### Building

To build the static site:

```bash
hugo --gc --minify
```

The built site will be in `docs/public/`.

## Structure

```
docs/
├── content/           # Markdown content
│   ├── _index.md     # Home page
│   └── docs/         # Documentation sections
│       ├── introduction/
│       ├── core-matchers/
│       ├── matchers-for-your-types/
│       ├── factories/
│       ├── factories-and-matchers/
│       ├── property-based-testing/
│       ├── advanced/
│       ├── api-reference/
│       └── examples/
├── hugo.toml         # Hugo configuration
└── go.mod           # Hugo modules

```

## Theme

We use the [Hugo Book](https://github.com/alex-shpak/hugo-book) theme via Hugo modules.

## Deployment

Documentation is automatically deployed to GitHub Pages when changes are pushed to the `main` branch:

- Workflow: `.github/workflows/docs.yml`
- Deployment: GitHub Pages (gh-pages branch)

## Adding Content

### New Section

1. Create a directory under `content/docs/`
2. Add an `_index.md` file with front matter:
   ```yaml
   ---
   title: "Section Title"
   weight: <number>
   bookCollapseSection: false
   ---
   ```

### New Page

Create a `.md` file in the appropriate section directory with front matter:

```yaml
---
title: "Page Title"
weight: <number>
---
```

## Writing Style

- Use code blocks with language hints: ` ```go `
- Keep examples short and focused
- Include both success and failure cases where relevant
- Link to related sections using `[text](/docs/section/)`
