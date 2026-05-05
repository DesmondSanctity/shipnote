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

	commits, err := g.Log(ctx, rng.From.SHA, rng.To.SHA)
	if err != nil {
		return Outputs{}, fmt.Errorf("git log: %w", err)
	}

	prs, err := collectPRs(ctx, in, owner, name, commits)
	if err != nil {
		return Outputs{}, fmt.Errorf("github collect: %w", err)
	}

	rel := analyze.Analyze(buildSourcePRs(prs, commits), analyzeOptions(in.Cfg))
	rel.ID = computeReleaseID(owner, name, rng)
	rel.Repo = model.Repo{
		Provider:      "github",
		Owner:         owner,
		Name:          name,
		URL:           fmt.Sprintf("https://github.com/%s/%s", owner, name),
		DefaultBranch: "",
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

	md := renderMarkdown(rel)
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

func buildSourcePRs(prs []github.PR, commits []repo.Commit) []analyze.SourcePR {
	bySHA := make(map[string]repo.Commit, len(commits))
	for _, c := range commits {
		bySHA[c.SHA] = c
	}
	out := make([]analyze.SourcePR, 0, len(prs))
	for _, pr := range prs {
		if pr.Number == 0 {
			// Commit had no associated PR (direct push). Skip; PRs are
			// the source of truth in v0.1.
			continue
		}
		var attached []analyze.SourceCommit
		if c, ok := bySHA[pr.MergeCommit]; ok {
			attached = append(attached, analyze.SourceCommit{
				SHA:      c.SHA,
				ShortSHA: c.ShortSHA,
				Message:  joinSubjectBody(c.Subject, c.Body),
			})
		}
		out = append(out, analyze.SourcePR{PR: pr, Commits: attached})
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
