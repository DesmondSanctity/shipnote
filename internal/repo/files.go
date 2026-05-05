package repo

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// FileChange is one file touched by a commit.
type FileChange struct {
	Path      string
	OldPath   string // set for renames/copies; empty otherwise
	Status    string // git status letter: A, M, D, R, C, T
	Additions int
	Deletions int
	Binary    bool
}

// CommitFiles returns the files changed by a single commit. For merge
// commits we diff against the first parent, matching what `git log -p`
// shows by default. The root commit's diff is taken against the empty
// tree so the very first commit's added files are reported correctly.
func (g *Git) CommitFiles(ctx context.Context, sha string) ([]FileChange, error) {
	if sha == "" {
		return nil, fmt.Errorf("repo.CommitFiles: sha is empty")
	}
	// --numstat for additions/deletions, --name-status (combined via
	// --raw) for the status letter and rename info. We make two cheap
	// calls and merge by path; this is simpler than parsing the
	// interleaved output of a single call.
	numstat, err := g.runLines(ctx,
		"show", "--no-color", "--first-parent", "--no-renames",
		"--format=", "--numstat", sha)
	if err != nil {
		return nil, fmt.Errorf("numstat %s: %w", sha, err)
	}
	statusLines, err := g.runLines(ctx,
		"show", "--no-color", "--first-parent",
		"--format=", "--name-status", sha)
	if err != nil {
		return nil, fmt.Errorf("name-status %s: %w", sha, err)
	}

	byPath := make(map[string]*FileChange, len(numstat))
	for _, ln := range numstat {
		fc, err := parseNumstatLine(ln)
		if err != nil {
			return nil, err
		}
		byPath[fc.Path] = fc
	}
	for _, ln := range statusLines {
		path, status, oldPath, err := parseStatusLine(ln)
		if err != nil {
			return nil, err
		}
		fc, ok := byPath[path]
		if !ok {
			fc = &FileChange{Path: path}
			byPath[path] = fc
		}
		fc.Status = status
		fc.OldPath = oldPath
	}

	out := make([]FileChange, 0, len(byPath))
	for _, fc := range byPath {
		out = append(out, *fc)
	}
	return out, nil
}

// parseNumstatLine parses a single line of `git --numstat` output.
// Format: "<adds>\t<dels>\t<path>" where adds/dels are "-" for binary.
func parseNumstatLine(ln string) (*FileChange, error) {
	parts := strings.SplitN(ln, "\t", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("malformed numstat line: %q", ln)
	}
	fc := &FileChange{Path: parts[2]}
	if parts[0] == "-" || parts[1] == "-" {
		fc.Binary = true
		return fc, nil
	}
	add, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("parse additions in %q: %w", ln, err)
	}
	del, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("parse deletions in %q: %w", ln, err)
	}
	fc.Additions, fc.Deletions = add, del
	return fc, nil
}

// parseStatusLine parses a single line of `git --name-status` output.
// Most entries are "<X>\t<path>"; renames/copies are "<X><score>\t<old>\t<new>".
func parseStatusLine(ln string) (path, status, oldPath string, err error) {
	parts := strings.Split(ln, "\t")
	if len(parts) < 2 {
		return "", "", "", fmt.Errorf("malformed status line: %q", ln)
	}
	status = string(parts[0][0])
	switch status {
	case "R", "C":
		if len(parts) < 3 {
			return "", "", "", fmt.Errorf("rename/copy missing path: %q", ln)
		}
		return parts[2], status, parts[1], nil
	default:
		return parts[1], status, "", nil
	}
}
