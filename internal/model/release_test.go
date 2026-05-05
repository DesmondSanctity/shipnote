package model_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
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
