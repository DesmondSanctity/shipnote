package render

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/DesmondSanctity/shipnote/internal/model"
)

// JSON serializes a Release as canonical, deterministic JSON.
//
// Guarantees:
//   - Indented with two spaces, '\n' terminated.
//   - HTML escaping disabled so URLs and angle-brackets round-trip cleanly.
//   - Field order is fixed by the struct definitions in package model;
//     none of the types use maps, so output is byte-stable across runs.
func JSON(r model.Release) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(r); err != nil {
		return nil, fmt.Errorf("encode release: %w", err)
	}
	return buf.Bytes(), nil
}
