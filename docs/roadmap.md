# Roadmap

shipnote is shipped in milestones, not on a calendar. Each phase has an exit gate; we move when the gate passes. The full long-form roadmap (including post-v1 cloud, chat, multi-host plans) lives in [thinkering/07-roadmap.md](../thinkering/07-roadmap.md). This file is the public-facing summary.

Status legend: ✅ shipped · 🟡 in progress · ⬜ planned

## v0.1 — alpha (current)

End-to-end usable on a single repo.

- ✅ Collectors: `git` range walk + GitHub PR fetch
- ✅ Categorization: PR labels, Conventional Commits, `changelog:` PR-body blocks
- ✅ Breaking-change detection (`!`, `BREAKING CHANGE:`, `breaking` label)
- ✅ Markdown + JSON renderers, schema v1.0.0 draft
- ✅ `init`, `generate`, `preview`, `doctor` commands
- ✅ Deterministic output, `--check` drift mode
- ✅ Optional AI summaries: OpenAI, Anthropic, Groq, Ollama, custom OpenAI-compatible — see [docs/ai.md](ai.md)
- ✅ On-disk AI cache, prompt-version pinning

**Exit gate:** clean release notes for two real OSS repos and our own; byte-deterministic across re-runs.

## v0.2 — beta (public)

Monorepo + GitHub Action.

- ⬜ Monorepo detection: pnpm/npm workspaces, Cargo workspaces, Go workspaces
- ⬜ Per-package release notes
- ⬜ GitHub Release rendering
- ⬜ GitHub Action with PR-preview comment
- ⬜ Documentation site (Astro or mkdocs)

**Exit gate:** PR-preview comment running on a real OSS repo.

## v0.3 — zero-setup AI

Deliver on the "no API key, no setup" promise.

- ⬜ llamafile provider
- ⬜ Bundled-model registry + consent download flow
- ⬜ `audiences = [...]` config to narrow which audiences are generated
- ⬜ `shipnote diff` command (changes between two refs without rendering a release)

**Exit gate:** A user with no API key, no Ollama, no config runs `shipnote generate --ai` and gets usable output after one consented download.

## v0.4 — polish & adoption

Safe to recommend.

- ⬜ Custom Markdown templates (Go templates)
- ⬜ `release` command with GitHub Release draft creation
- ⬜ Migration-notes generation from labelled before/after PR pairs
- ⬜ Better squash/merge/rebase handling
- ⬜ Robust caching layer for GitHub API responses
- ⬜ `--dry-run` everywhere meaningful

**Exit gate:** 5+ external OSS repos using shipnote in production CI; no determinism bugs reported for 4 weeks.

## v1.0 — stable

Schema and CLI surface stability commitment.

- ⬜ Schema v1.0.0 locked (additive-only afterwards)
- ⬜ CLI flag surface locked (deprecation policy starts)
- ⬜ Stable Action `@v1`
- ⬜ Plugin interface designed (not exposed publicly yet)
- ⬜ Full docs site

**Exit gate:** No breaking changes for 8 weeks; schema validated by external integrators.

---

## Beyond v1

These are tracked separately from the OSS milestones:

- **Cloud** — hosted ingestion, public changelog page per project, search across versions, diff-between-versions UI, org dashboards. Treated as a separate product that consumes `release.json`.
- **Chat / intelligence** — search-grounded Q&A over normalized releases, version-aware comparisons, per-customer migration guides, RSS / webhook delivery.
- **Beyond GitHub** — GitLab, Bitbucket, Gitea, self-hosted git hosts.

## Cross-cutting tracks

Continuous, not phase-bound:

- **Documentation** — every command, flag, env var, and config key documented.
- **ADRs** — `docs/adr/` for any decision that survives a rewrite (schema changes, source-adapter additions, AI prompt-version bumps, default-model swaps).
- **Examples** — real-world fixtures from public repos.
- **Performance** — benchmark suite from v0.2 onward.
- **Security** — supply-chain hardening, signed releases (cosign), SBOM, `SECURITY.md` kept current.
- **Community** — discussions, issue triage, monthly community calls once we have users.

## Risk-triggered pivots

We will reconsider this roadmap if any of these trigger:

- Adoption stalls in v0.4 → re-evaluate the wedge; consider sharper SDK-vendor positioning.
- AI quality from tiny local models is unusable → default `--ai` to remote providers (Groq free tier) and document local as opt-in.
- Schema needs breaking changes after v1 → bump to v2 with a migration tool; never silently break.
- Determinism bugs prove hard to eliminate → add `--strict-deterministic` and accept best-effort by default.
