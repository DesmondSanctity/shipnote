package github_test

import (
	"testing"

	"github.com/DesmondSanctity/shipnote/internal/collect/github"
)

func TestNewDefaults(t *testing.T) {
	t.Parallel()
	c := github.New(github.Options{})
	if !c.Anonymous() {
		t.Error("zero Options should yield anonymous client")
	}

	c2 := github.New(github.Options{Token: "tok"})
	if c2.Anonymous() {
		t.Error("token should disable anonymous mode")
	}
}
