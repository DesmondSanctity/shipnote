package render

import (
	"fmt"
	"sort"
	"strings"

	"github.com/DesmondSanctity/shipnote/internal/model"
)

// Sticky-section markers wrap the rendered body so subsequent runs can
// splice the new content in without disturbing surrounding hand-written
// text in CHANGELOG.md.
const (
	StickyStart = "<!-- shipnote:start:unreleased -->"
	StickyEnd   = "<!-- shipnote:end:unreleased -->"
)

// categoryOrder is the locked rendering order. Mirrors analyze.CategoryOrder
// but kept independent so render does not depend on analyze.
var categoryOrder = []model.ChangeType{
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

// categoryHeading is the user-facing label for a category.
var categoryHeading = map[model.ChangeType]string{
	model.ChangeTypeBreaking:    "Breaking Changes",
	model.ChangeTypeSecurity:    "Security",
	model.ChangeTypeFeat:        "Added",
	model.ChangeTypeFix:         "Fixed",
	model.ChangeTypePerf:        "Performance",
	model.ChangeTypeDeprecation: "Deprecated",
	model.ChangeTypeRefactor:    "Refactored",
	model.ChangeTypeDocs:        "Documentation",
	model.ChangeTypeBuild:       "Build",
	model.ChangeTypeCI:          "CI",
	model.ChangeTypeTest:        "Tests",
	model.ChangeTypeChore:       "Chores",
	model.ChangeTypeOther:       "Other",
}

// Markdown renders the release as a Keep-a-Changelog-flavored block.
//
// Output shape:
//
//	<sticky-start>
//	## [tag or Unreleased] - YYYY-MM-DD
//	### Breaking Changes ... (only if any)
//	### Added / Fixed / ...
//	### Contributors
//	<sticky-end>
//
// The function is pure and deterministic; all sorting is stable.
func Markdown(r model.Release) string {
	var sb strings.Builder
	sb.WriteString(StickyStart)
	sb.WriteByte('\n')
	writeHeader(&sb, r)
	writeBreakingNotes(&sb, r.BreakingChanges)
	writeCategorySections(&sb, r.GlobalChanges)
	writeContributors(&sb, r.Contributors)
	sb.WriteString(StickyEnd)
	sb.WriteByte('\n')
	return sb.String()
}

func writeHeader(sb *strings.Builder, r model.Release) {
	tag := strings.TrimSpace(r.Release.Tag)
	if tag == "" {
		tag = "Unreleased"
	}
	date := r.Range.To.Date
	if date == "" {
		date = r.Metadata.GeneratedAt
	}
	date = strings.SplitN(date, "T", 2)[0]
	if date == "" {
		fmt.Fprintf(sb, "## [%s]\n\n", tag)
		return
	}
	fmt.Fprintf(sb, "## [%s] - %s\n\n", tag, date)
}

func writeBreakingNotes(sb *strings.Builder, breaking []model.BreakingChange) {
	if len(breaking) == 0 {
		return
	}
	sb.WriteString("> **Breaking changes in this release.** Review the migration notes below before upgrading.\n\n")
}

func writeCategorySections(sb *strings.Builder, changes []model.Change) {
	buckets := make(map[model.ChangeType][]model.Change, len(categoryOrder))
	for _, c := range changes {
		key := c.Type
		if c.Breaking {
			key = model.ChangeTypeBreaking
		}
		buckets[key] = append(buckets[key], c)
	}
	for _, cat := range categoryOrder {
		items := buckets[cat]
		if len(items) == 0 {
			continue
		}
		fmt.Fprintf(sb, "### %s\n\n", categoryHeading[cat])
		for _, c := range items {
			sb.WriteString(renderChangeLine(c))
			sb.WriteByte('\n')
		}
		sb.WriteByte('\n')
	}
}

func renderChangeLine(c model.Change) string {
	var sb strings.Builder
	sb.WriteString("- ")
	if c.Breaking && c.Type != model.ChangeTypeBreaking {
		sb.WriteString("**BREAKING:** ")
	}
	if c.Scope != "" {
		fmt.Fprintf(&sb, "**%s:** ", c.Scope)
	}
	sb.WriteString(strings.TrimSpace(c.Title))
	if pr := c.Source.PR; pr != nil && pr.Number > 0 {
		if pr.URL != "" {
			fmt.Fprintf(&sb, " ([#%d](%s))", pr.Number, pr.URL)
		} else {
			fmt.Fprintf(&sb, " (#%d)", pr.Number)
		}
	}
	if login := c.Author.Login; login != "" {
		if c.Author.URL != "" {
			fmt.Fprintf(&sb, " by [@%s](%s)", login, c.Author.URL)
		} else {
			fmt.Fprintf(&sb, " by @%s", login)
		}
	} else if c.Author.Name != "" {
		fmt.Fprintf(&sb, " by %s", c.Author.Name)
	}
	if c.Migration != nil && strings.TrimSpace(c.Migration.Summary) != "" {
		fmt.Fprintf(&sb, "\n  - _Migration:_ %s", strings.TrimSpace(c.Migration.Summary))
	}
	return sb.String()
}

func writeContributors(sb *strings.Builder, contribs []model.Contributor) {
	if len(contribs) == 0 {
		return
	}
	ordered := make([]model.Contributor, len(contribs))
	copy(ordered, contribs)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Changes != ordered[j].Changes {
			return ordered[i].Changes > ordered[j].Changes
		}
		return loginOrName(ordered[i]) < loginOrName(ordered[j])
	})
	sb.WriteString("### Contributors\n\n")
	for _, c := range ordered {
		sb.WriteString("- ")
		if c.Login != "" {
			if c.URL != "" {
				fmt.Fprintf(sb, "[@%s](%s)", c.Login, c.URL)
			} else {
				fmt.Fprintf(sb, "@%s", c.Login)
			}
		} else if c.Name != "" {
			sb.WriteString(c.Name)
		}
		if c.Changes > 0 {
			fmt.Fprintf(sb, " — %d change", c.Changes)
			if c.Changes != 1 {
				sb.WriteByte('s')
			}
		}
		sb.WriteByte('\n')
	}
	sb.WriteByte('\n')
}

func loginOrName(c model.Contributor) string {
	if c.Login != "" {
		return strings.ToLower(c.Login)
	}
	return strings.ToLower(c.Name)
}
