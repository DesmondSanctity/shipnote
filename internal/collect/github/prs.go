package github

import (
	"context"
	"fmt"
	"strings"
)

// PRsForCommits returns one PR per input SHA in the same order. If a
// commit is not associated with any PR (direct push) the corresponding
// slot is the zero PR with Number == 0; callers should filter those out
// or treat them as commit-only changes.
//
// Internally we issue one GraphQL request that aliases each commit
// lookup as c0, c1, … cN. GitHub's GraphQL has a node-cost budget so
// callers should batch in pages of <= 50 SHAs; PRsForCommitsPaged
// handles paging when caller-supplied lists grow large.
func (c *Client) PRsForCommits(ctx context.Context, owner, name string, shas []string) ([]PR, error) {
	if owner == "" || name == "" {
		return nil, fmt.Errorf("github: owner and name are required")
	}
	if len(shas) == 0 {
		return nil, nil
	}

	var b strings.Builder
	b.WriteString("query($owner:String!,$name:String!){repository(owner:$owner,name:$name){")
	for i, sha := range shas {
		fmt.Fprintf(&b, "c%d: object(oid: \"%s\") { ... on Commit { %s } } ", i, sha, commitFragment)
	}
	b.WriteString("}}")
	vars := map[string]any{"owner": owner, "name": name}

	raw := struct {
		Repository map[string]*commitNode `json:"repository"`
	}{}
	if err := c.gql(ctx, b.String(), vars, &raw); err != nil {
		return nil, err
	}

	out := make([]PR, len(shas))
	for i := range shas {
		node := raw.Repository[fmt.Sprintf("c%d", i)]
		if node == nil || len(node.AssociatedPullRequests.Nodes) == 0 {
			continue
		}
		out[i] = node.AssociatedPullRequests.Nodes[0].toPR()
	}
	return out, nil
}

// commitFragment is the GraphQL selection set we want from each Commit
// node. We pick a single associated PR (the most recent merged one) per
// commit; squash-merges always produce exactly one, and that's the
// dominant case in modern repos.
const commitFragment = `oid associatedPullRequests(first:1, orderBy:{field:UPDATED_AT, direction:DESC}) {
	nodes {
		number title body url state mergedAt baseRefName headRefName
		mergeCommit { oid }
		author { login url avatarUrl __typename ... on User { name } }
		labels(first: 50) { nodes { name color } }
		files(first: 100) {
			nodes { path additions deletions changeType }
		}
	}
}`

type commitNode struct {
	OID                    string `json:"oid"`
	AssociatedPullRequests struct {
		Nodes []prNode `json:"nodes"`
	} `json:"associatedPullRequests"`
}

type prNode struct {
	Number      int    `json:"number"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	URL         string `json:"url"`
	State       string `json:"state"`
	MergedAt    string `json:"mergedAt"`
	BaseRefName string `json:"baseRefName"`
	HeadRefName string `json:"headRefName"`
	MergeCommit *struct {
		OID string `json:"oid"`
	} `json:"mergeCommit"`
	Author *struct {
		Login     string `json:"login"`
		Name      string `json:"name"`
		URL       string `json:"url"`
		AvatarURL string `json:"avatarUrl"`
		Typename  string `json:"__typename"`
	} `json:"author"`
	Labels struct {
		Nodes []Label `json:"nodes"`
	} `json:"labels"`
	Files struct {
		Nodes []prFileNode `json:"nodes"`
	} `json:"files"`
}

type prFileNode struct {
	Path       string `json:"path"`
	Additions  int    `json:"additions"`
	Deletions  int    `json:"deletions"`
	ChangeType string `json:"changeType"`
}

func (n prNode) toPR() PR {
	pr := PR{
		Number:   n.Number,
		Title:    n.Title,
		Body:     n.Body,
		URL:      n.URL,
		State:    n.State,
		MergedAt: n.MergedAt,
		BaseRef:  n.BaseRefName,
		HeadRef:  n.HeadRefName,
		Labels:   append([]Label(nil), n.Labels.Nodes...),
		Files:    make([]File, 0, len(n.Files.Nodes)),
	}
	if n.MergeCommit != nil {
		pr.MergeCommit = n.MergeCommit.OID
	}
	if n.Author != nil {
		pr.Author = User{
			Login:     n.Author.Login,
			Name:      n.Author.Name,
			URL:       n.Author.URL,
			AvatarURL: n.Author.AvatarURL,
			IsBot:     n.Author.Typename == "Bot",
		}
	}
	for _, f := range n.Files.Nodes {
		pr.Files = append(pr.Files, File{
			Path:      f.Path,
			Additions: f.Additions,
			Deletions: f.Deletions,
			Status:    strings.ToLower(f.ChangeType),
		})
	}
	return pr
}
