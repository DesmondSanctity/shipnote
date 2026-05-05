// Package analyze categorizes and groups collected changes into the
// Package analyze turns collector output (PRs + commits) into a normalized
// model.Release. Pure functions: no I/O, no globals, deterministic for
// equal inputs. Each concern (categorize, breaking, group, impact, stats)
// gets its own file; the orchestrator lives in analyze.go.
package analyze
