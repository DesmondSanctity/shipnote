package analyze

import (
	"sort"

	"github.com/DesmondSanctity/shipnote/internal/model"
)

// CategoryOrder is the locked, render-friendly order for change
// categories. The renderer uses the same slice so JSON output and
// CHANGELOG sections always agree.
//
// The order matches docs/SCHEMA.md and is deliberately user-impact
// first: breaking → security → feat → fix → perf → deprecation, with
// developer-facing categories trailing.
var CategoryOrder = []model.ChangeType{
	model.ChangeTypeBreaking,
	model.ChangeTypeSecurity,
	model.ChangeTypeFeat,
	model.ChangeTypeFix,
	model.ChangeTypePerf,
	model.ChangeTypeDeprecation,
	model.ChangeTypeRefactor,
	model.ChangeTypeDocs,
	model.ChangeTypeBuild,
	model.ChangeTypeCI,
	model.ChangeTypeTest,
	model.ChangeTypeChore,
	model.ChangeTypeOther,
}

// Group sorts changes into the locked category order, then by
// title/ID within each bucket so output is deterministic. Breaking
// changes appear under ChangeTypeBreaking regardless of their
// underlying type so they surface at the top of the changelog.
//
// The input slice is not mutated. The returned slice is a stable
// reordering with the same length.
func Group(changes []model.Change) []model.Change {
	if len(changes) == 0 {
		return nil
	}
	idx := make(map[model.ChangeType]int, len(CategoryOrder))
	for i, t := range CategoryOrder {
		idx[t] = i
	}
	const lastBucket = 999

	out := append([]model.Change(nil), changes...)
	sort.SliceStable(out, func(i, j int) bool {
		ai, aj := bucketOf(out[i], idx, lastBucket), bucketOf(out[j], idx, lastBucket)
		if ai != aj {
			return ai < aj
		}
		// tie-break: title, then ID for stability across runs
		if out[i].Title != out[j].Title {
			return out[i].Title < out[j].Title
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func bucketOf(c model.Change, idx map[model.ChangeType]int, fallback int) int {
	if c.Breaking {
		return idx[model.ChangeTypeBreaking]
	}
	if i, ok := idx[c.Type]; ok {
		return i
	}
	return fallback
}
