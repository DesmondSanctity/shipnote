package model

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// DeterminismInputs is the canonicalized tuple hashed into
// metadata.determinismKey. Two shipnote runs that produce the same
// DeterminismInputs MUST produce byte-identical Release records (modulo
// fields explicitly excluded from determinism, e.g. metadata itself).
//
// ConfigHash is an opaque digest of the resolved configuration the run
// used (e.g. sha256 over the canonicalized .shipnote.toml). It is the
// caller's responsibility to compute it; passing the empty string is
// allowed but disables config sensitivity.
type DeterminismInputs struct {
	Repo       Repo   `json:"repo"`
	Range      Range  `json:"range"`
	ConfigHash string `json:"configHash"`
}

// ComputeDeterminismKey returns the sha256-prefixed determinism key for
// the given inputs. The serialization is canonical JSON (sorted keys,
// no whitespace) produced by encoding/json's default Marshal of a
// struct with stable field order.
//
// Format: "sha256:<64 hex chars>".
func ComputeDeterminismKey(in DeterminismInputs) (string, error) {
	canonical, err := json.Marshal(in)
	if err != nil {
		return "", fmt.Errorf("marshal determinism inputs: %w", err)
	}
	sum := sha256.Sum256(canonical)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
