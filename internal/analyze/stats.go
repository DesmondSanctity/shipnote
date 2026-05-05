package analyze

import "github.com/DesmondSanctity/shipnote/internal/model"

// ComputeStats counts changes per category and totals the file
// footprint across the supplied changes. Excluded is set by the caller
// after Filter has run; this function only fills the per-change
// aggregates.
//
// "Internal" rolls up all developer-facing buckets (chore, refactor,
// build, ci, test, other) so the rendered Stats line stays compact.
func ComputeStats(changes []model.Change) model.Stats {
	var s model.Stats
	files := make(map[string]struct{}, len(changes)*4)
	for _, c := range changes {
		s.Changes++
		if c.Breaking {
			s.Breaking++
		}
		switch c.Type {
		case model.ChangeTypeFeat:
			s.Features++
		case model.ChangeTypeFix:
			s.Fixes++
		case model.ChangeTypeDeprecation:
			s.Deprecations++
		case model.ChangeTypePerf:
			s.Perf++
		case model.ChangeTypeSecurity:
			s.Security++
		case model.ChangeTypeDocs:
			s.Docs++
		case model.ChangeTypeChore,
			model.ChangeTypeRefactor,
			model.ChangeTypeBuild,
			model.ChangeTypeCI,
			model.ChangeTypeTest,
			model.ChangeTypeOther:
			s.Internal++
		case model.ChangeTypeBreaking:
			// Breaking is counted via the breaking bit above; the
			// "breaking" type label is rare and shouldn't double-count.
		}
		// File footprint: Files.List takes precedence (full list), then
		// Top, then we fall back to the Count when neither is populated.
		if list := c.Files.List; list != nil {
			for _, f := range list {
				files[f.Path] = struct{}{}
			}
		} else if len(c.Files.Top) > 0 {
			for _, f := range c.Files.Top {
				files[f.Path] = struct{}{}
			}
		}
	}
	if len(files) > 0 {
		s.FilesChanged = len(files)
	} else {
		// No file paths were available — sum the per-change counts as a
		// best-effort approximation.
		for _, c := range changes {
			s.FilesChanged += c.Files.Count
		}
	}
	return s
}
