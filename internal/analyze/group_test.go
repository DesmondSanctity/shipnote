package analyze

import (
	"testing"

	"github.com/DesmondSanctity/shipnote/internal/model"
)

func TestGroupOrdersByLockedCategoryOrder(t *testing.T) {
	in := []model.Change{
		{ID: "1", Type: model.ChangeTypeChore, Title: "z"},
		{ID: "2", Type: model.ChangeTypeFix, Title: "a"},
		{ID: "3", Type: model.ChangeTypeFeat, Title: "b"},
		{ID: "4", Type: model.ChangeTypeSecurity, Title: "c"},
		{ID: "5", Type: model.ChangeTypeFeat, Title: "a"}, // tied with #3 on type, sorted by title
	}
	out := Group(in)
	wantIDs := []string{"4", "5", "3", "2", "1"}
	for i, id := range wantIDs {
		if out[i].ID != id {
			t.Errorf("position %d: got %s, want %s (full: %+v)", i, out[i].ID, id, ids(out))
		}
	}
}

func TestGroupBreakingFloatsToTop(t *testing.T) {
	in := []model.Change{
		{ID: "1", Type: model.ChangeTypeFeat, Title: "a"},
		{ID: "2", Type: model.ChangeTypeFix, Title: "b", Breaking: true},
	}
	out := Group(in)
	if out[0].ID != "2" {
		t.Errorf("breaking should be first: %+v", ids(out))
	}
}

func TestGroupEmpty(t *testing.T) {
	if got := Group(nil); got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestGroupDoesNotMutateInput(t *testing.T) {
	in := []model.Change{
		{ID: "1", Type: model.ChangeTypeChore},
		{ID: "2", Type: model.ChangeTypeFeat},
	}
	_ = Group(in)
	if in[0].ID != "1" || in[1].ID != "2" {
		t.Errorf("Group mutated input: %+v", ids(in))
	}
}

func ids(cs []model.Change) []string {
	out := make([]string, len(cs))
	for i, c := range cs {
		out[i] = c.ID
	}
	return out
}
