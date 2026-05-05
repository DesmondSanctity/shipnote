// Package ai hosts the optional AI provider interface and the
// auto-detection chain. AI is off by default in CI and is never
// required for shipnote to produce output. This is one of only two
// packages permitted to make outbound network calls (the other is
// internal/collect/github).
package ai
