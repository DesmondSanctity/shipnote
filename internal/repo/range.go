package repo

import (
	"context"
	"fmt"
)

// Range describes a commit range in the form (From, To]. Either side
// may be a tag, branch, or SHA; "" for From means "from the root commit".
type Range struct {
	From RangePoint
	To   RangePoint
}

// RangePoint is one end of a Range, fully resolved.
type RangePoint struct {
	Ref  string // user-supplied label (e.g. "v2.3.0", "HEAD")
	SHA  string
	Date string // RFC3339 UTC
}

// ResolveRange turns user-supplied refs into a fully resolved Range.
// fromRef may be empty to mean "most recent tag matching tagPattern, or
// the root commit if there are no matching tags". toRef defaults to
// "HEAD" when empty.
func (g *Git) ResolveRange(ctx context.Context, fromRef, toRef, tagPattern string) (Range, error) {
	if toRef == "" {
		toRef = "HEAD"
	}
	toSHA, toDate, err := g.ResolveRef(ctx, toRef)
	if err != nil {
		return Range{}, fmt.Errorf("resolve to-ref %q: %w", toRef, err)
	}

	if fromRef == "" {
		latest, ok, err := g.LatestTag(ctx, tagPattern)
		if err != nil {
			return Range{}, fmt.Errorf("autodetect from-ref: %w", err)
		}
		if ok {
			fromRef = latest.Name
		}
	}

	var from RangePoint
	if fromRef == "" {
		// No matching tag — fall back to the root commit. The half-open
		// range (root, HEAD] excludes root itself, which matches what
		// users intuitively want for "everything in the project".
		root, err := g.FirstCommit(ctx)
		if err != nil {
			return Range{}, fmt.Errorf("autodetect root: %w", err)
		}
		_, rootDate, err := g.ResolveRef(ctx, root)
		if err != nil {
			return Range{}, fmt.Errorf("resolve root commit: %w", err)
		}
		from = RangePoint{Ref: root, SHA: root, Date: rootDate}
	} else {
		sha, date, err := g.ResolveRef(ctx, fromRef)
		if err != nil {
			return Range{}, fmt.Errorf("resolve from-ref %q: %w", fromRef, err)
		}
		from = RangePoint{Ref: fromRef, SHA: sha, Date: date}
	}

	return Range{
		From: from,
		To:   RangePoint{Ref: toRef, SHA: toSHA, Date: toDate},
	}, nil
}
