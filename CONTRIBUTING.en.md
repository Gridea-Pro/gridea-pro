# Contributing to Gridea Pro

First off, thank you for considering contributing to Gridea Pro! Every contribution helps make this project better for the entire community.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Workflow](#workflow)
- [Commit Messages](#commit-messages)
- [Pull Request Guidelines](#pull-request-guidelines)
- [Theme Development](#theme-development)
- [Reporting Bugs](#reporting-bugs)
- [Suggesting Features](#suggesting-features)
- [Community](#community)

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). By participating, you are expected to uphold this code. Please report unacceptable behavior via [Issues](https://github.com/Gridea-Pro/gridea-pro/issues).

## How Can I Contribute?

There are many ways to contribute beyond writing code:

- **Report bugs** — Found something broken? [Open a bug report](https://github.com/Gridea-Pro/gridea-pro/issues/new?template=bug_report.yml).
- **Suggest features** — Have an idea? [Open a feature request](https://github.com/Gridea-Pro/gridea-pro/issues/new?template=feature_request.yml).
- **Create themes** — Design a theme and share it with the community.
- **Improve documentation** — Fix typos, clarify instructions, add examples.
- **Translate** — Help make Gridea Pro accessible in more languages.
- **Answer questions** — Help others in [Discussions](https://github.com/Gridea-Pro/gridea-pro/discussions).
- **Spread the word** — Star the repo, write a blog post, tell a friend.

## Development Setup

### Prerequisites

- [Go](https://go.dev/dl/) 1.22+
- [Node.js](https://nodejs.org/) 18+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) v2
- [Git](https://git-scm.com/)

### Getting Started

```bash
# 1. Fork the repository on GitHub, then clone your fork
git clone https://github.com/<your-username>/gridea-pro.git
cd gridea-pro

# 2. Install frontend dependencies
cd frontend && npm install && cd ..

# 3. Run in development mode
wails dev
```

The app will launch with hot-reload enabled for the frontend. Changes to Go code require a restart.

### Building

```bash
# Build for your current platform
wails build
```

## Project Structure

```
gridea-pro/
├── main.go              # Application entry point
├── app.go               # Wails application logic & bindings
├── frontend/            # Vue 3 + Vite + Tailwind CSS
│   ├── src/
│   │   ├── components/  # Vue components
│   │   ├── views/       # Page views
│   │   └── ...
│   └── package.json
├── build/               # Build resources (icons, installer configs)
├── themes/              # Built-in themes
└── .github/
    └── workflows/       # CI/CD (GitHub Actions)
```

## Workflow

We use **GitHub Flow** — a simple branch-based workflow:

1. **Create a branch** from `main`:
   ```bash
   git checkout -b feat/my-feature
   # or
   git checkout -b fix/some-bug
   ```

2. **Make your changes** with clear, focused commits.

3. **Push** your branch:
   ```bash
   git push origin feat/my-feature
   ```

4. **Open a Pull Request** against `main`.

### Branch Naming

| Prefix | Purpose | Example |
|--------|---------|---------|
| `feat/` | New feature | `feat/dark-mode` |
| `fix/` | Bug fix | `fix/image-upload-crash` |
| `docs/` | Documentation | `docs/theme-api` |
| `refactor/` | Code refactoring | `refactor/renderer` |
| `chore/` | Tooling, CI, deps | `chore/update-deps` |

## Commit Messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <short description>

[optional body]
```

**Types:** `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`, `ci`

**Examples:**

```
feat(editor): add image drag-and-drop support
fix(renderer): resolve Jinja2 loop variable scoping issue
docs(readme): update installation instructions
chore(ci): add Linux ARM build to release workflow
```

## Pull Request Guidelines

- **One concern per PR.** Keep pull requests focused — a single feature or bug fix.
- **Describe what and why.** Include context, screenshots, or GIFs where helpful.
- **Reference related issues.** Use `Closes #123` or `Fixes #456` in the PR description.
- **Ensure it builds.** Run `wails build` locally before submitting.
- **Be patient.** We review PRs as quickly as we can. Feedback is meant to improve the code, not criticize.

### PR Title Format

Follow the same convention as commit messages:

```
feat(editor): add image drag-and-drop support
```

## Theme Development

Gridea Pro supports three template engines: **Jinja2 (Pongo2)**, **EJS**, and **Go Templates**. Creating a theme is one of the best ways to contribute.

### Quick Start

1. Check out the [Theme Development Documentation](https://gridea.pro/docs/themes) for the full API reference.
2. Use an existing theme (like `flavor`) as a starting point.
3. A theme consists of templates, styles, and a `theme.toml` config file.

### Theme Structure

```
my-theme/
├── theme.toml           # Theme metadata & configuration
├── templates/
│   ├── index.html       # Homepage
│   ├── post.html        # Single post
│   ├── tag.html         # Tag archive
│   └── ...
├── assets/
│   ├── styles/
│   └── scripts/
└── preview.png          # Theme preview image (800×600)
```

### Submitting a Theme

Once your theme is ready, share it with the community:

1. Publish it as a standalone GitHub repository.
2. Open a Discussion in the [Show and Tell](https://github.com/Gridea-Pro/gridea-pro/discussions/categories/show-and-tell) category.
3. We may feature popular community themes on the official website.

## Reporting Bugs

Use the [Bug Report template](https://github.com/Gridea-Pro/gridea-pro/issues/new?template=bug_report.yml). Good bug reports include:

- Gridea Pro version and OS
- Steps to reproduce
- Expected vs. actual behavior
- Screenshots or error logs if available

## Suggesting Features

Use the [Feature Request template](https://github.com/Gridea-Pro/gridea-pro/issues/new?template=feature_request.yml). Helpful feature requests include:

- The problem you're trying to solve
- Your proposed solution
- Alternatives you've considered

For early-stage ideas, start a thread in [Discussions](https://github.com/Gridea-Pro/gridea-pro/discussions) first — it's a great way to get feedback before opening a formal request.

## Community

- 💬 [GitHub Discussions](https://github.com/Gridea-Pro/gridea-pro/discussions) — Questions, ideas, and conversation
- 🐛 [Issue Tracker](https://github.com/Gridea-Pro/gridea-pro/issues) — Bug reports and feature requests
- 🌐 [Official Website](https://gridea.pro) — Downloads, docs, and themes

---

## Pre-submission Checklist (enforced by CI)

Automated checks run on every PR — **the PR cannot be merged unless all of them pass**. Please verify locally before submitting:

- Go code formatted with `gofmt -w` (CI rejects unformatted code)
- `go vet ./...` and `go test ./backend/...` pass
- Frontend `npm run lint` reports no errors
- Any added/changed UI strings are translated into **all 12 locales** (`frontend/src/locales/`);
  run `python3 scripts/check_i18n.py frontend/src/locales` to verify; no Chinese left in en.json
- Changes tested locally via `wails dev`

Once all checks are green, the PR awaits maintainer review and approval before merging.

---

Thank you for helping make Gridea Pro better! 🎉
