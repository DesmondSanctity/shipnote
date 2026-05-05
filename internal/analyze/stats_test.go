package analyze

import (
	"testing"

	"github.com/DesmondSanctity/shipnote/internal/model"
)

func TestComputeStatsCountsByCategoryAndBreaking(t *testing.T) {
	in := []model.Change{
		{Type: model.ChangeTypeFeat},
		{Type: model.ChangeTypeFeat, Breaking: true},
		{Type: model.ChangeTypeFix},
		{Type: model.ChangeTypeChore},
		{Type: model.ChangeTypeRefactor},
		{Type: model.ChangeTypeSecurity},
	}
	s := ComputeStats(in)
	if s.Changes != 6 {
		t.Errorf("Changes = %d, want 6", s.Changes)
	}
	if s.Features != 2 {
		t.Errorf("Features = %d, want 2", s.Features)
	}
	if s.Fixes != 1 {
		t.Errorf("Fixes = %d, want 1", s.Fixes)
	}
	if s.Breaking != 1 {
		t.Errorf("Breaking = %d, want 1", s.Breaking)
	}
	if s.Security != 1 {
		t.Errorf("Security = %d, want 1", s.Security)
	}
	if s.Internal != 2 { // chore + refactor
		t.Errorf("Internal = %d, want 2", s.Internal)
	}
}

func TestComputeStatsFileFootprintFromList(t *testing.T) {
	in := []model.Change{
		{Files: model.FileSummary{List: []model.FileDiff{{Path: "a.go"}, {Path: "b.go"}}}},
		{Files: model.FileSummary{List: []model.FileDiff{{Path: "b.go"}, {Path: "c.go"}}}},
	}
	s := ComputeStats(in)
	if s.FilesChanged != 3 {
		t.Errorf("FilesChanged = %d, want 3 (deduplicated)", s.FilesChanged)
	}
}

func TestComputeStatsFileFootprintFromTopFallsBackToCount(t *testing.T) {
	in := []model.Change{
		{Files: model.FileSummary{Count: 5}},
		{Files: model.FileSummary{Count: 3}},
	}
	s := ComputeStats(in)
	if s.FilesChanged != 8 {
		t.Errorf("FilesChanged = %d, want 8 (sum of Count)", s.FilesChanged)
	}
}
