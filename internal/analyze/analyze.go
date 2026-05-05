package analyze

import (
	"fmt"

	gh "github.com/DesmondSanctity/shipnote/internal/collect/github"
	"github.com/DesmondSanctity/shipnote/internal/model"
)

// SourceCommit is the analyzer-side commit shape. Collectors translate
// their native types (repo.Commit, github commit nodes) into this so
// analyze does not depend on any collector package by import.
type SourceCommit struct {
	SHA      string
	ShortSHA string
	URL      string
	Message  string
}

// SourcePR pairs a github.PR with the headline commits it landed.
// Commits is empty for direct pushes that have no associated PR.
type SourcePR struct {
	PR      gh.PR
	Commits []SourceCommit
}

// Options bundles the knobs the orchestrator needs. Fields are
// optional; the zero value is valid for a single-package repo with
// default filtering.
type Options struct {
	Mapper    PackageMapper
	Merger    *IdentityMerger
	Filter    Filter
	RootPkg   string // package ID to assign when Mapper returns ""
	Generator model.Generator
}

// Analyze converts collector output into a model.Release. It runs each
// PR through categorize / breaking / impact, applies the filter,
// builds Change records, sorts them via Group, computes Stats, and
// rolls up Contributors.
//
// This function is pure: same inputs produce the same Release, with
// deterministic ordering at every level.
func Analyze(prs []SourcePR, opts Options) model.Release {
	rel := model.Release{
		SchemaVersion: model.SchemaVersion,
		Metadata:      model.Metadata{Generator: opts.Generator},
	}
	if opts.Mapper == nil {
		opts.Mapper = SinglePackageMapper{RootID: opts.RootPkg}
	}

	changes := make([]model.Change, 0, len(prs))
	excluded := make([]model.ExcludedRecord, 0)
	contribs := make([]model.Contributor, 0, len(prs))

	for _, src := range prs {
		labels := labelNames(src.PR.Labels)
		filePaths := filePaths(src.PR.Files)
		input := Input{
			Labels:     labels,
			PRTitle:    src.PR.Title,
			PRBody:     src.PR.Body,
			CommitMsg:  firstCommitMessage(src.Commits),
			FilesPaths: filePaths,
		}

		cat := Categorize(input)
		brk := DetectBreaking(input)

		ch := buildChange(src, cat, brk, opts.Mapper)

		decision := opts.Filter.Apply(ch, labels, src.PR.Author.IsBot)
		if !decision.Keep {
			excluded = append(excluded, model.ExcludedRecord{
				ID:     prID(src.PR),
				Reason: decision.Reason,
			})
			continue
		}
		changes = append(changes, ch)
		if ch.Author.Login != "" || ch.Author.Name != "" {
			contribs = append(contribs, model.Contributor{
				Login:   ch.Author.Login,
				Name:    ch.Author.Name,
				URL:     ch.Author.URL,
				Changes: 1,
			})
		}
	}

	rel.GlobalChanges = Group(changes)
	rel.Stats = ComputeStats(rel.GlobalChanges)
	rel.Stats.Excluded = len(excluded)
	rel.Excluded = excluded
	rel.Contributors = MergeContributors(opts.Merger, contribs)
	rel.Stats.Contributors = len(rel.Contributors)
	rel.BreakingChanges = collectBreaking(rel.GlobalChanges)
	return rel
}

func buildChange(src SourcePR, cat CategoryDecision, brk BreakingDecision, mapper PackageMapper) model.Change {
	pr := src.PR
	ch := model.Change{
		ID:       prID(pr),
		Type:     cat.Type,
		Breaking: brk.Breaking,
		Title:    StripCCPrefix(pr.Title),
		Source: model.ChangeSource{
			Kind:    model.SourceKindPR,
			Adapter: "github-prs",
			PR: &model.PRSource{
				Number:        pr.Number,
				URL:           pr.URL,
				MergedAt:      pr.MergedAt,
				MergeStrategy: model.MergeStrategy(pr.MergeStrategy),
				Labels:        labelNames(pr.Labels),
			},
			Commits: convertCommits(src.Commits),
		},
		Author: model.Contributor{
			Login: pr.Author.Login,
			Name:  pr.Author.Name,
			URL:   pr.Author.URL,
		},
		Packages: PackagesForFiles(mapper, filePaths(pr.Files)),
		Files:    fileSummary(pr.Files),
		Confidence: model.Confidence{
			Category: cat.Confidence,
			Breaking: brk.Confidence,
		},
	}
	if brk.Notice != "" {
		ch.Migration = &model.Migration{Summary: brk.Notice}
	}
	return ch
}

func collectBreaking(changes []model.Change) []model.BreakingChange {
	var out []model.BreakingChange
	for _, c := range changes {
		if !c.Breaking {
			continue
		}
		bc := model.BreakingChange{
			ChangeID: c.ID,
			Summary:  c.Title,
		}
		if c.Migration != nil {
			bc.Migration = *c.Migration
			bc.Impact = c.Migration.Summary
		}
		out = append(out, bc)
	}
	return out
}

func labelNames(labels []gh.Label) []string {
	if len(labels) == 0 {
		return nil
	}
	out := make([]string, len(labels))
	for i, l := range labels {
		out[i] = l.Name
	}
	return out
}

func filePaths(files []gh.File) []string {
	if len(files) == 0 {
		return nil
	}
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = f.Path
	}
	return out
}

func fileSummary(files []gh.File) model.FileSummary {
	s := model.FileSummary{Count: len(files)}
	for _, f := range files {
		s.Additions += f.Additions
		s.Deletions += f.Deletions
		s.Top = append(s.Top, model.FileDiff{
			Path:      f.Path,
			Additions: f.Additions,
			Deletions: f.Deletions,
			Status:    f.Status,
		})
	}
	return s
}

func convertCommits(cs []SourceCommit) []model.Commit {
	if len(cs) == 0 {
		return nil
	}
	out := make([]model.Commit, len(cs))
	for i, c := range cs {
		out[i] = model.Commit{SHA: c.SHA, ShortSHA: c.ShortSHA, URL: c.URL, Message: c.Message}
	}
	return out
}

func firstCommitMessage(cs []SourceCommit) string {
	if len(cs) == 0 {
		return ""
	}
	return cs[0].Message
}

func prID(pr gh.PR) string {
	if pr.Number > 0 {
		return fmt.Sprintf("github-prs:%d", pr.Number)
	}
	return ""
}
