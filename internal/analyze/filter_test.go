package analyze

import (
	"testing"

	"github.com/DesmondSanctity/shipnote/internal/model"
)

func TestFilterIgnoresBots(t *testing.T) {
	f := Filter{IgnoreBots: true}
	d := f.Apply(model.Change{}, nil, true)
	if d.Keep || d.Reason != "bot" {
		t.Errorf("got %+v", d)
	}
}

func TestFilterIgnoresLogins(t *testing.T) {
	f := Filter{IgnoreLogins: LowerSet([]string{"Dependabot"})}
	d := f.Apply(model.Change{Author: model.Contributor{Login: "dependabot"}}, nil, false)
	if d.Keep || d.Reason != "ignored-login" {
		t.Errorf("got %+v", d)
	}
}

func TestFilterIgnoresLabels(t *testing.T) {
	f := Filter{IgnoreLabels: LowerSet([]string{"skip-changelog"})}
	d := f.Apply(model.Change{}, []string{"type:feat", "Skip-Changelog"}, false)
	if d.Keep || d.Reason != "ignored-label" {
		t.Errorf("got %+v", d)
	}
}

func TestFilterDropsChoreOnly(t *testing.T) {
	f := Filter{DropChoreOnly: true}
	d := f.Apply(model.Change{Type: model.ChangeTypeChore}, nil, false)
	if d.Keep || d.Reason != "chore-only" {
		t.Errorf("got %+v", d)
	}
}

func TestFilterKeepsBreakingChoreEvenWhenChoreOnlyDropped(t *testing.T) {
	f := Filter{DropChoreOnly: true}
	d := f.Apply(model.Change{Type: model.ChangeTypeChore, Breaking: true}, nil, false)
	if !d.Keep {
		t.Errorf("expected breaking chore to survive: %+v", d)
	}
}

func TestFilterDefaultsKeepEverything(t *testing.T) {
	d := Filter{}.Apply(model.Change{Type: model.ChangeTypeFeat}, []string{"feat"}, true)
	if !d.Keep {
		t.Errorf("expected keep with empty filter: %+v", d)
	}
}
