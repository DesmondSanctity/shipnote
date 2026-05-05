package repo

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// Commit is a single commit returned by Log.
type Commit struct {
	SHA            string
	ShortSHA       string
	AuthorName     string
	AuthorEmail    string
	CommitterName  string
	CommitterEmail string
	Date           string // committer date, RFC3339 UTC
	Subject        string
	Body           string
	Parents        []string
}

// Log returns the commits in the half-open range (from, to], i.e. the
// commits reachable from `to` but not from `from`. Pass an empty `from`
// to walk all commits reachable from `to`. Newest commit first.
func (g *Git) Log(ctx context.Context, from, to string) ([]Commit, error) {
	if to == "" {
		return nil, fmt.Errorf("repo.Log: 'to' ref is required")
	}
	rng := to
	if from != "" {
		rng = from + ".." + to
	}
	// %x1f = unit separator between fields, %x1e = record separator.
	const sep = "\x1f"
	const rec = "\x1e"
	format := strings.Join([]string{
		"%H", "%h", "%an", "%ae", "%cn", "%ce", "%cI", "%P", "%s", "%b",
	}, sep) + rec
	out, err := g.run(ctx, "log", "--format="+format, rng)
	if err != nil {
		return nil, fmt.Errorf("git log %s: %w", rng, err)
	}
	records := strings.Split(strings.TrimRight(string(out), rec+"\n"), rec)
	commits := make([]Commit, 0, len(records))
	for _, raw := range records {
		raw = strings.TrimLeft(raw, "\n")
		if raw == "" {
			continue
		}
		fields := strings.SplitN(raw, sep, 10)
		if len(fields) != 10 {
			return nil, fmt.Errorf("malformed log record (%d fields): %q",
				len(fields), raw)
		}
		var parents []string
		if p := strings.TrimSpace(fields[7]); p != "" {
			parents = strings.Split(p, " ")
		}
		commits = append(commits, Commit{
			SHA:            fields[0],
			ShortSHA:       fields[1],
			AuthorName:     fields[2],
			AuthorEmail:    fields[3],
			CommitterName:  fields[4],
			CommitterEmail: fields[5],
			Date:           fields[6],
			Parents:        parents,
			Subject:        fields[8],
			Body:           strings.TrimRight(fields[9], "\n"),
		})
	}
	return commits, nil
}

// CountReachable returns the number of commits reachable from ref. Useful
// for sanity-checking range bounds before kicking off an expensive walk.
func (g *Git) CountReachable(ctx context.Context, ref string) (int, error) {
	out, err := g.run(ctx, "rev-list", "--count", ref)
	if err != nil {
		return 0, fmt.Errorf("rev-list count %s: %w", ref, err)
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, fmt.Errorf("parse count %q: %w", out, err)
	}
	return n, nil
}
