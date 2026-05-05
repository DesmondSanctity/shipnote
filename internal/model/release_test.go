package model_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/DesmondSanctity/shipnote/internal/model"
)

const (
	schemaRelPath  = "../../schema/release.schema.json"
	fixtureRelPath = "testdata/release_minimal.json"
)

// TestSchemaRoundTrip verifies that a known-good fixture:
//  1. Validates against schema/release.schema.json,
//  2. Unmarshals into model.Release without error,
//  3. Re-marshals, and the result still validates against the schema,
//  4. The re-marshaled JSON is semantically equal to the input
//     (object-key reordering is allowed, but no fields may be added or
//     dropped).
func TestSchemaRoundTrip(t *testing.T) {
	t.Parallel()
	fixtureBytes := mustReadFile(t, fixtureRelPath)
	schema := mustCompileSchema(t, schemaRelPath)

	mustValidate(t, schema, fixtureBytes, "fixture")

	var rel model.Release
	if err := json.Unmarshal(fixtureBytes, &rel); err != nil {
		t.Fatalf("unmarshal fixture into model.Release: %v", err)
	}
	if rel.SchemaVersion != model.SchemaVersion {
		t.Fatalf("schemaVersion mismatch: fixture=%q model.SchemaVersion=%q",
			rel.SchemaVersion, model.SchemaVersion)
	}

	roundTripped, err := json.Marshal(rel)
	if err != nil {
		t.Fatalf("re-marshal model.Release: %v", err)
	}
	mustValidate(t, schema, roundTripped, "round-tripped")

	var origDecoded any
	if err := json.Unmarshal(fixtureBytes, &origDecoded); err != nil {
		t.Fatalf("decode fixture for compare: %v", err)
	}
	origNormalized, err := json.Marshal(origDecoded)
	if err != nil {
		t.Fatalf("re-encode fixture for compare: %v", err)
	}
	var rtDecoded any
	if err := json.Unmarshal(roundTripped, &rtDecoded); err != nil {
		t.Fatalf("decode round-tripped for compare: %v", err)
	}
	rtNormalized, err := json.Marshal(rtDecoded)
	if err != nil {
		t.Fatalf("re-encode round-tripped for compare: %v", err)
	}
	if !bytes.Equal(origNormalized, rtNormalized) {
		t.Errorf("round-trip lost or added fields\n--- fixture (normalized)\n%s\n--- round-tripped (normalized)\n%s",
			origNormalized, rtNormalized)
	}
}

// TestSchemaConstants asserts the Go-side SchemaVersion is set.
func TestSchemaConstants(t *testing.T) {
	t.Parallel()
	if model.SchemaVersion == "" {
		t.Fatal("model.SchemaVersion must not be empty")
	}
}

func mustReadFile(t *testing.T, rel string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Clean(rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return b
}

func mustCompileSchema(t *testing.T, rel string) *jsonschema.Schema {
	t.Helper()
	abs, err := filepath.Abs(rel)
	if err != nil {
		t.Fatalf("abs path for %s: %v", rel, err)
	}
	schema, err := jsonschema.NewCompiler().Compile(abs)
	if err != nil {
		t.Fatalf("compile schema %s: %v", abs, err)
	}
	return schema
}

func mustValidate(t *testing.T, schema *jsonschema.Schema, raw []byte, label string) {
	t.Helper()
	var doc any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("%s: unmarshal for validation: %v", label, err)
	}
	if err := schema.Validate(doc); err != nil {
		t.Fatalf("%s: schema validation failed:\n%v", label, err)
	}
}

func sampleDetInputs() model.DeterminismInputs {
	return model.DeterminismInputs{
		Repo: model.Repo{
			Provider:      "github",
			Owner:         "acme",
			Name:          "sdk",
			URL:           "https://github.com/acme/sdk",
			DefaultBranch: "main",
		},
		Range: model.Range{
			From: model.RangePoint{Ref: "v2.3.0", SHA: "a1b2c3d", Date: "2026-04-15T00:00:00Z"},
			To:   model.RangePoint{Ref: "HEAD", SHA: "f9e8d7c", Date: "2026-05-01T11:42:08Z"},
		},
		ConfigHash: "sha256:cfg",
	}
}

// TestDeterminismKeyStable asserts the key is reproducible for identical inputs.
func TestDeterminismKeyStable(t *testing.T) {
	t.Parallel()
	in := sampleDetInputs()
	k1, err := model.ComputeDeterminismKey(in)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	k2, err := model.ComputeDeterminismKey(in)
	if err != nil {
		t.Fatalf("compute repeat: %v", err)
	}
	if k1 != k2 {
		t.Fatalf("non-deterministic: %q vs %q", k1, k2)
	}
	if !strings.HasPrefix(k1, "sha256:") || len(k1) != len("sha256:")+64 {
		t.Fatalf("key has wrong shape: %q", k1)
	}
}

// TestDeterminismKeySensitive asserts each input field changes the key.
func TestDeterminismKeySensitive(t *testing.T) {
	t.Parallel()
	baseKey, err := model.ComputeDeterminismKey(sampleDetInputs())
	if err != nil {
		t.Fatalf("compute base: %v", err)
	}
	mutations := map[string]func(*model.DeterminismInputs){
		"repo.owner":     func(in *model.DeterminismInputs) { in.Repo.Owner = "other" },
		"repo.name":      func(in *model.DeterminismInputs) { in.Repo.Name = "other" },
		"repo.branch":    func(in *model.DeterminismInputs) { in.Repo.DefaultBranch = "trunk" },
		"range.from.sha": func(in *model.DeterminismInputs) { in.Range.From.SHA = "deadbee" },
		"range.to.sha":   func(in *model.DeterminismInputs) { in.Range.To.SHA = "deadbee" },
		"range.to.date":  func(in *model.DeterminismInputs) { in.Range.To.Date = "2026-05-02T00:00:00Z" },
		"configHash":     func(in *model.DeterminismInputs) { in.ConfigHash = "sha256:other" },
	}
	for label, mutate := range mutations {
		mutated := sampleDetInputs()
		mutate(&mutated)
		got, err := model.ComputeDeterminismKey(mutated)
		if err != nil {
			t.Fatalf("compute %s: %v", label, err)
		}
		if got == baseKey {
			t.Errorf("%s: mutation did not change key", label)
		}
	}
}
