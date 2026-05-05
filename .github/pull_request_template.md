<!--
Thanks for contributing to shipnote.
Please fill out the changelog block below — it's how this PR will be summarized
in release notes.
-->

## Summary

<!-- One or two sentences describing what this PR does and why. -->

## Changelog

```changelog:
category: feat | fix | perf | docs | refactor | test | ci | build | chore | breaking | deprecate | security
audience: developer | customer | migration   # optional, defaults to developer
summary: <one sentence, written for the audience above>
breaking: <only if category=breaking — describe the break and migration>
```

<!--
Examples:

category: feat
audience: developer
summary: Add `--check` mode to `shipnote generate` for CI drift detection.

category: breaking
summary: Rename `output.path` to `output.markdown`.
breaking: Update your `.shipnote.toml`: `[output] path = ...` → `[output] markdown = ...`.
-->

## Checklist

- [ ] Tests added or updated
- [ ] `go test ./...` passes locally
- [ ] `golangci-lint run` passes locally
- [ ] If this changes behavior, golden fixtures regenerated and reviewed
- [ ] If this changes the schema, ADR added under `docs/adr/`
