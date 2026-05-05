// Package render serializes a release model to markdown / JSON / etc.
//
// Renderers MUST be deterministic: stable sort everywhere, no map
// iteration, no time.Now(), no environment reads.
package render
