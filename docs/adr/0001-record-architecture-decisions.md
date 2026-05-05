# ADR 0001 — Record architecture decisions

- **Status:** Accepted
- **Date:** 2026-05-05
- **Deciders:** @DesmondSanctity

## Context

shipnote is a multi-component system (collector, analyzer, renderer, AI, schema) and decisions made early — schema shape, determinism rules, source-adapter interfaces, AI strategy — will be costly to revisit later. We need a lightweight way to capture _why_ each significant decision was made so future contributors don't re-litigate settled questions or unknowingly violate invariants.

## Decision

We will use lightweight Architecture Decision Records (ADRs), one Markdown file per decision under `docs/adr/`, numbered sequentially starting at `0001`. The format follows Michael Nygard's original ADR template (Context / Decision / Consequences) with minor extensions (Alternatives considered, References).

Process:

1. Copy `docs/adr/0000-template.md` to `docs/adr/NNNN-<short-title>.md` with the next free number.
2. Set status to `Proposed`, open a PR.
3. On merge, change status to `Accepted`.
4. Superseding decisions update the old ADR's status to `Superseded by ADR-NNNN` and link the replacement.
5. ADRs are append-only after acceptance — corrections go in a new ADR, not edits to the old one (typo fixes excepted).

ADRs are required for:

- Schema-shape changes after v1.0.0
- New source adapters
- Changes to determinism guarantees
- New external dependencies with significant license/operational impact
- AI provider interface changes

## Consequences

### Positive

- Future contributors can read ADRs in number order to understand how the system arrived at its current shape.
- PR reviews of architectural changes have a clear deliverable (the ADR) instead of long Slack-style discussions.
- Makes it explicit which decisions are settled vs open.

### Negative / costs

- Slight overhead per architectural PR.
- Risk of ADR rot if not enforced — mitigated by `CONTRIBUTING.md` guidance and PR reviewers asking "is there an ADR?"

### Risks

- ADRs that are too verbose discourage authors. Mitigation: the template is short and we explicitly value brevity.

## Alternatives considered

### Option A — No ADRs, rely on PR descriptions

Rejected. PR descriptions are hard to discover months later and disappear into git history.

### Option B — Wiki / external doc site

Rejected. ADRs that live in the repo evolve alongside the code they describe and survive forks.

## References

- Michael Nygard, "Documenting architecture decisions" (2011)
- `docs/adr/0000-template.md`
