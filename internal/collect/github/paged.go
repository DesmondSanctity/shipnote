package github

import (
	"context"
	"fmt"
	"sync"
)

// PageOptions tunes batched, concurrent calls to PRsForCommitsPaged.
//
// PageSize bounds how many SHAs go into one GraphQL request; GitHub
// applies a per-request node-cost budget so values above ~50 risk
// MAX_NODE_LIMIT_EXCEEDED. Concurrency bounds how many such requests
// are in flight at once. Anonymous clients are forced to Concurrency=1
// to stay polite under the unauthenticated rate limit.
type PageOptions struct {
	PageSize    int
	Concurrency int
}

// DefaultPageOptions returns the recommended defaults: 50 commits per
// request, 8 concurrent requests.
func DefaultPageOptions() PageOptions {
	return PageOptions{PageSize: 50, Concurrency: 8}
}

// PRsForCommitsPaged splits shas into pages of opts.PageSize and runs
// up to opts.Concurrency PRsForCommits calls in parallel. Output order
// matches the input slice. The first error from any page cancels the
// remaining work and is returned; partial results are discarded.
func (c *Client) PRsForCommitsPaged(ctx context.Context, owner, name string, shas []string, opts PageOptions) ([]PR, error) {
	if owner == "" || name == "" {
		return nil, fmt.Errorf("github: owner and name are required")
	}
	if len(shas) == 0 {
		return nil, nil
	}
	page := opts.PageSize
	if page <= 0 {
		page = 50
	}
	conc := opts.Concurrency
	if conc <= 0 {
		conc = 8
	}
	if c.Anonymous() && conc > 2 {
		conc = 2
	}

	out := make([]PR, len(shas))
	type job struct {
		start int
		shas  []string
	}
	jobs := make([]job, 0, (len(shas)+page-1)/page)
	for i := 0; i < len(shas); i += page {
		end := i + page
		if end > len(shas) {
			end = len(shas)
		}
		jobs = append(jobs, job{start: i, shas: shas[i:end]})
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sem := make(chan struct{}, conc)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for _, j := range jobs {
		j := j
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			if ctx.Err() != nil {
				return
			}
			prs, err := c.PRsForCommits(ctx, owner, name, j.shas)
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
					cancel()
				}
				mu.Unlock()
				return
			}
			for k, pr := range prs {
				out[j.start+k] = pr
			}
		}()
	}
	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	return out, nil
}
