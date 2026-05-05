package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommandHelp(t *testing.T) {
	t.Parallel()

	root := NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("--help should not error, got: %v", err)
	}

	out := buf.String()
	for _, want := range []string{"shipnote", "init", "generate", "preview", "doctor"} {
		if !strings.Contains(out, want) {
			t.Errorf("--help output missing %q\nfull output:\n%s", want, out)
		}
	}
}

func TestRootCommandVersion(t *testing.T) {
	t.Parallel()

	root := NewRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"--version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("--version should not error, got: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "schema:") {
		t.Errorf("--version output should mention schema version, got:\n%s", out)
	}
}
