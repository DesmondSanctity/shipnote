// Package runner executes the end-to-end shipnote pipeline:
// repo -> github collector -> analyzer -> renderers.
//
// Both `shipnote generate` and `shipnote preview` go through here so
// they share identical resolution logic; the only difference between
// them is whether the caller writes the rendered artifacts to disk.
package runner

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/DesmondSanctity/shipnote/internal/ai"
	"github.com/DesmondSanctity/shipnote/internal/analyze"
	"github.com/DesmondSanctity/shipnote/internal/collect/github"
	"github.com/DesmondSanctity/shipnote/internal/config"
	"github.com/DesmondSanctity/shipnote/internal/model"
	"github.com/DesmondSanctity/shipnote/internal/repo"
	"github.com/DesmondSanctity/shipnote/internal/version"
)

// Inputs configures one Run.
type Inputs struct {
	RepoDir   string        // path to the working tree
	FromRef   string        // empty = latest tag, else explicit
	ToRef     string        // empty = HEAD
	TagGlob   string        // tag pattern for from-ref autodetect, default "v*"
	Token     string        // GitHub token; empty = anonymous
	Cfg       config.Config // resolved config
	CacheDir  string        // GitHub response cache; empty disables
	NoNetwork bool          // skip GitHub call (git-only fallback)
	Endpoint  string        // GraphQL endpoint override; empty = api.github.com (test hook)

	// AI is optional; when AIProvider is nil the pipeline is the
	// deterministic baseline (Markdown identical to AI-off runs).
	AIProvider  ai.Summarizer
	AICache     *ai.Cache
	AIAudiences []ai.Audience
	AILogger    func(msg string)

	// Warn receives non-fatal advisories (e.g. "you are on a non-default
	// branch"). The runner will not abort even if Warn is nil.
	Warn func(msg string)
}

// Outputs is what Run produces. Markdown / JSON are byte-identical
// across runs given the same Inputs.
type Outputs struct {
	Release  model.Release
	Markdown string
	JSON     []byte
}

// Run drives the full pipeline. Pure with respect to the network and
// disk: callers decide whether to persist Markdown/JSON.
func Run(ctx context.Context, in Inputs) (Outputs, error) {
	g, err := repo.Open(ctx, in.RepoDir)
	if err != nil {
		return Outputs{}, fmt.Errorf("open repo: %w", err)
	}

	owner, name, err := resolveOwnerName(ctx, g, in.Cfg.Repo)
	if err != nil {
		return Outputs{}, fmt.Errorf("resolve repo identity: %w", err)
	}

	tagGlob := in.TagGlob
	if tagGlob == "" {
		tagGlob = "v*"
	}
	rng, err := g.ResolveRange(ctx, in.FromRef, in.ToRef, tagGlob)
	if err != nil {
		return Outputs{}, fmt.Errorf("resolve range: %w", err)
	}

	// Tag set used for inference of Release.Tag and PreviousReleaseID.
	allTags, err := g.Tags(ctx, "")
	if err != nil {
		return Outputs{}, fmt.Errorf("list tags: %w", err)
	}
	tagSet := make(map[string]struct{}, len(allTags))
	for _, t := range allTags {
		tagSet[t.Name] = struct{}{}
	}

	defaultBranch := g.DefaultBranch(ctx)
	if cur, err := g.CurrentBranch(ctx); err == nil && in.Warn != nil {
		// Only warn when the user asked for HEAD (implicit or explicit)
		// AND HEAD is on a non-default branch. An explicit ref/SHA/tag
		// is intentional, never noisy.
		if cur != "" && cur != "HEAD" && cur != defaultBranch &&
			(in.ToRef == "" || in.ToRef == "HEAD") {
			in.Warn(fmt.Sprintf(
				"shipnote: HEAD is on branch %q (default is %q); changelog will reflect this branch's history",
				cur, defaultBranch,
			))
		}
	}

	commits, err := g.Log(ctx, rng.From.SHA, rng.To.SHA)
	if err != nil {
		return Outputs{}, fmt.Errorf("git log: %w", err)
	}

	prs, err := collectPRs(ctx, in, owner, name, commits)
	if err != nil {
		return Outputs{}, fmt.Errorf("github collect: %w", err)
	}

	includeDirect := !in.Cfg.Ignore.IgnoresDirectPushes()
	var firstParents map[string]struct{}
	if includeDirect {
		firstParents, err = g.FirstParentSHAs(ctx, rng.From.SHA, rng.To.SHA)
		if err != nil {
			return Outputs{}, fmt.Errorf("first-parent walk: %w", err)
		}
	}

	rel := analyze.Analyze(
		buildSourcePRs(prs, commits, includeDirect, firstParents),
		analyzeOptions(in.Cfg),
	)
	rel.ID = computeReleaseID(owner, name, rng)
	rel.Repo = model.Repo{
		Provider:      "github",
		Owner:         owner,
		Name:          name,
		URL:           fmt.Sprintf("https://github.com/%s/%s", owner, name),
		DefaultBranch: defaultBranch,
	}
	rel.Range = model.Range{
		From: model.RangePoint{Ref: rng.From.Ref, SHA: rng.From.SHA, Date: rng.From.Date},
		To:   model.RangePoint{Ref: rng.To.Ref, SHA: rng.To.SHA, Date: rng.To.Date},
	}
	rel.Metadata.Generator = model.Generator{Name: "shipnote", Version: version.Version}
	rel.Metadata.GeneratedAt = rng.To.Date
	key, _ := model.ComputeDeterminismKey(model.DeterminismInputs{
		Repo:  rel.Repo,
		Range: rel.Range,
	})
	rel.Metadata.DeterminismKey = key
	rel.Metadata.Sources = []string{"github-prs", "git"}
	rel.Release.State = model.ReleaseStateUnreleased

	// Tag inference: if the user pointed --to at an existing tag, treat
	// this as a released changelog and stamp Release.Tag + publishedAt.
	// If --from is also a tag, link the previous release id so the site
	// can render a proper "previous release" pointer.
	if _, ok := tagSet[rng.To.Ref]; ok {
		rel.Release.Tag = rng.To.Ref
		rel.Release.State = model.ReleaseStateReleased
		published := rng.To.Date
		rel.Release.PublishedAt = &published
	}
	if _, ok := tagSet[rng.From.Ref]; ok {
		rel.PreviousReleaseID = "rel-" + rng.From.Ref
	}

	md := renderMarkdown(rel)
	// AI runs AFTER the deterministic Markdown is captured so that
	// disabling AI cannot alter the canonical changelog body. JSON
	// below picks up the populated AISummaries.
	if in.AIProvider != nil {
		_ = ai.Enrich(ctx, &rel, ai.Options{
			Provider:  in.AIProvider,
			Cache:     in.AICache,
			Audiences: in.AIAudiences,
			Logger:    in.AILogger,
		})
	}
	js, err := renderJSON(rel)
	if err != nil {
		return Outputs{}, fmt.Errorf("render json: %w", err)
	}
	return Outputs{Release: rel, Markdown: md, JSON: js}, nil
}

func resolveOwnerName(ctx context.Context, g *repo.Git, rcfg config.Repo) (string, string, error) {
	if rcfg.Owner != "" && rcfg.Owner != "auto" && rcfg.Name != "" && rcfg.Name != "auto" {
		return rcfg.Owner, rcfg.Name, nil
	}
	u, err := g.OriginURL(ctx)
	if err != nil {
		return "", "", err
	}
	if u == "" {
		return "", "", fmt.Errorf("no remote configured; set [repo].owner and [repo].name in .shipnote.toml")
	}
	owner, name, ok := repo.ParseGitHubRemote(u)
	if !ok {
		return "", "", fmt.Errorf("remote %q is not a github.com URL", u)
	}
	return owner, name, nil
}

func collectPRs(ctx context.Context, in Inputs, owner, name string, commits []repo.Commit) ([]github.PR, error) {
	if in.NoNetwork {
		return nil, nil
	}
	shas := make([]string, len(commits))
	for i, c := range commits {
		shas[i] = c.SHA
	}
	if len(shas) == 0 {
		return nil, nil
	}
	opts := github.Options{Token: in.Token, Endpoint: in.Endpoint}
	if in.CacheDir != "" {
		opts.Cache = github.NewFileCache(in.CacheDir)
	}
	client := github.New(opts)
	return client.PRsForCommits(ctx, owner, name, shas)
}

func buildSourcePRs(
	prs []github.PR,
	commits []repo.Commit,
	includeDirect bool,
	firstParents map[string]struct{},
) []analyze.SourcePR {
	bySHA := make(map[string]repo.Commit, len(commits))
	for _, c := range commits {
		bySHA[c.SHA] = c
	}
	out := make([]analyze.SourcePR, 0, len(prs))
	seen := make(map[int]int, len(prs))            // PR number -> index into out
	covered := make(map[string]struct{}, len(prs)) // SHAs already represented by a PR
	// prs is aligned with the commits list passed to PRsForCommits, so
	// prs[i] corresponds to commits[i]. We mark commits[i].SHA as covered
	// when prs[i].Number > 0 (a real association exists), AND any
	// MergeCommit SHA the PR points at.
	for i, pr := range prs {
		if pr.Number == 0 {
			continue
		}
		if i < len(commits) {
			covered[commits[i].SHA] = struct{}{}
		}
		if pr.MergeCommit != "" {
			covered[pr.MergeCommit] = struct{}{}
		}
		var sc *analyze.SourceCommit
		if c, ok := bySHA[pr.MergeCommit]; ok {
			sc = &analyze.SourceCommit{
				SHA:      c.SHA,
				ShortSHA: c.ShortSHA,
				Message:  joinSubjectBody(c.Subject, c.Body),
			}
		}
		if idx, ok := seen[pr.Number]; ok {
			// PR already recorded via another commit (merge + branch
			// commits both resolve to the same PR). Append the extra
			// commit if we haven't seen this SHA on it yet.
			if sc != nil {
				dup := false
				for _, existing := range out[idx].Commits {
					if existing.SHA == sc.SHA {
						dup = true
						break
					}
				}
				if !dup {
					out[idx].Commits = append(out[idx].Commits, *sc)
				}
			}
			continue
		}
		var attached []analyze.SourceCommit
		if sc != nil {
			attached = []analyze.SourceCommit{*sc}
		}
		seen[pr.Number] = len(out)
		out = append(out, analyze.SourcePR{PR: pr, Commits: attached})
	}
	if includeDirect {
		// Synthesize PR-less entries for commits on the first-parent
		// line of `to` that no PR claimed. Walk in reverse (oldest
		// first) so the analyzer sees a stable order; Group will
		// re-sort by category afterward anyway.
		for i := len(commits) - 1; i >= 0; i-- {
			c := commits[i]
			if _, ok := firstParents[c.SHA]; !ok {
				continue
			}
			if _, ok := covered[c.SHA]; ok {
				continue
			}
			out = append(out, analyze.SourcePR{
				PR: github.PR{
					Number:   0,
					Title:    c.Subject,
					Body:     c.Body,
					MergedAt: c.Date,
					Author: github.User{
						Login: c.AuthorEmail,
						Name:  c.AuthorName,
					},
				},
				Commits: []analyze.SourceCommit{{
					SHA:      c.SHA,
					ShortSHA: c.ShortSHA,
					Message:  joinSubjectBody(c.Subject, c.Body),
				}},
			})
		}
	}
	return out
}

func analyzeOptions(cfg config.Config) analyze.Options {
	opts := analyze.Options{
		RootPkg: cfg.Packages.Root,
		Filter: analyze.Filter{
			IgnoreLogins: analyze.LowerSet(cfg.Ignore.Authors),
			IgnoreLabels: analyze.LowerSet(cfg.Ignore.Labels),
			IgnoreBots:   true,
		},
		Generator: model.Generator{Name: "shipnote", Version: version.Version},
	}
	if len(cfg.Authors) > 0 {
		rules := make([]analyze.IdentityRule, 0, len(cfg.Authors))
		for _, a := range cfg.Authors {
			rules = append(rules, analyze.IdentityRule{
				Canonical: model.Contributor{
					Login: a.Canonical,
					Name:  a.Name,
					URL:   a.URL,
				},
				Aliases: a.Aliases,
			})
		}
		opts.Merger = analyze.NewIdentityMerger(rules)
	}
	return opts
}

func joinSubjectBody(subject, body string) string {
	if body == "" {
		return subject
	}
	return subject + "\n\n" + body
}

func computeReleaseID(owner, name string, rng repo.Range) string {
	h := sha256.Sum256([]byte(strings.Join([]string{owner, name, rng.From.SHA, rng.To.SHA}, "|")))
	return "rel_" + hex.EncodeToString(h[:8])
}
