# shipnote

> Release intelligence for SDK and developer-platform teams. CLI-first, CI-native, OSS.

**shipnote** turns your repository's PRs, labels, tags, and monorepo structure into deterministic, audience-targeted release notes you can trust in CI. Optional AI polish runs locally — never required.

> ⚠️ **Status: pre-alpha.** v0.1 is in active development. Schema is being stabilized; expect breaking changes until v1.0.

## Why

Most changelog tools generate from commits. Commits are the wrong unit. Pull requests, packages, and tags are the real units of change. shipnote treats them as such, and emits a structured `release.json` alongside human-readable markdown — so the same data can power a `CHANGELOG.md`, a docs page, search, or any downstream tool.

## Install

```bash
# Homebrew (coming soon)
brew install DesmondSanctity/tap/shipnote

# Go
go install github.com/DesmondSanctity/shipnote/cmd/shipnote@latest

# Binary releases: see https://github.com/DesmondSanctity/shipnote/releases
```

## 60-second quickstart

```bash
cd your-repo
shipnote init                       # writes .shipnote.toml, adds .shipnote/ to .gitignore
shipnote generate --from v1.0.0     # produces CHANGELOG.md and .shipnote/release.json
shipnote generate --check           # CI-safe: fails if regen would change output
```

## What's in the box (v0.1 scope)

- `git` + `github-prs` source adapters
- Categorization via PR labels, Conventional Commits, and `changelog:` PR-body blocks
- Breaking-change detection (CC `!`, `BREAKING CHANGE:` footer, `breaking` label)
- Markdown + JSON renderers
- Schema v1.0.0 (Apache-2.0)
- Deterministic output, `--check` mode for CI

Out of scope until later versions: monorepo, AI, GitHub Action, audience modes. See [docs/roadmap.md](docs/roadmap.md) once published.

## License

[Apache-2.0](LICENSE).
