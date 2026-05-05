package analyze

import "testing"

func TestDetectBreakingFromLabel(t *testing.T) {
	d := DetectBreaking(Input{Labels: []string{"breaking-change"}})
	if !d.Breaking || d.Source != "label" {
		t.Fatalf("got %+v", d)
	}
}

func TestDetectBreakingFromCCBang(t *testing.T) {
	d := DetectBreaking(Input{PRTitle: "feat!: drop v1 endpoints"})
	if !d.Breaking || d.Source != "cc-bang" {
		t.Fatalf("got %+v", d)
	}
}

func TestDetectBreakingFromFooter(t *testing.T) {
	body := "Reworked the auth flow.\n\nBREAKING CHANGE: tokens are now scoped per-team."
	d := DetectBreaking(Input{PRBody: body})
	if !d.Breaking || d.Source != "footer" {
		t.Fatalf("got %+v", d)
	}
	if d.Notice == "" {
		t.Errorf("expected notice text, got empty")
	}
}

func TestDetectBreakingDashedFooter(t *testing.T) {
	d := DetectBreaking(Input{PRBody: "BREAKING-CHANGE: removed deprecated flag"})
	if !d.Breaking {
		t.Fatalf("dashed footer not detected: %+v", d)
	}
}

func TestDetectBreakingDefault(t *testing.T) {
	d := DetectBreaking(Input{PRTitle: "feat: harmless addition"})
	if d.Breaking {
		t.Fatalf("false positive: %+v", d)
	}
}

func TestDetectBreakingLabelBeatsBody(t *testing.T) {
	d := DetectBreaking(Input{
		Labels:  []string{"semver:major"},
		PRBody:  "no breaking footer here",
		PRTitle: "feat: thing",
	})
	if d.Source != "label" {
		t.Errorf("expected label precedence: %+v", d)
	}
}
