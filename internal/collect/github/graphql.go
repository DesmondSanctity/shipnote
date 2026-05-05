package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// graphqlRequest is the wire shape POSTed to /graphql.
type graphqlRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

// graphqlResponse is the wire shape returned by /graphql. Data is left
// raw so callers can decode into their own typed shape.
type graphqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []graphqlError  `json:"errors,omitempty"`
}

type graphqlError struct {
	Message string   `json:"message"`
	Path    []string `json:"path,omitempty"`
	Type    string   `json:"type,omitempty"`
}

// gql executes a GraphQL query and decodes the "data" field into out.
// Network errors, HTTP non-2xx, and GraphQL-level errors all surface
// as a returned error.
//
// When the client has a Cache configured, gql is cache-first: a hit
// short-circuits the network call. On cache miss the response data is
// stored under a key derived from (query, variables). On a transport
// error after a miss, gql attempts a stale read so callers can degrade
// gracefully when GitHub is unreachable; if no entry exists, the
// transport error is returned.
func (c *Client) gql(ctx context.Context, query string, vars map[string]any, out any) error {
	varsJSON, err := json.Marshal(vars)
	if err != nil {
		return fmt.Errorf("encode variables: %w", err)
	}
	cacheKey := keyFor(query, string(varsJSON))

	if cached, err := c.cache.Get(cacheKey); err == nil {
		if out == nil {
			return nil
		}
		if jerr := json.Unmarshal(cached, out); jerr == nil {
			return nil
		}
		// fall through to network on corrupt cache entry
	}

	body, err := json.Marshal(graphqlRequest{Query: query, Variables: vars})
	if err != nil {
		return fmt.Errorf("encode request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		// Transport failure: try a (stale) cache read before giving up.
		if cached, cerr := c.cache.Get(cacheKey); cerr == nil && out != nil {
			if jerr := json.Unmarshal(cached, out); jerr == nil {
				return nil
			}
		}
		return fmt.Errorf("github request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("github http %d: %s", resp.StatusCode, truncate(string(raw), 240))
	}

	var gr graphqlResponse
	if err := json.Unmarshal(raw, &gr); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if len(gr.Errors) > 0 {
		return fmt.Errorf("graphql: %s", gr.Errors[0].Message)
	}
	// Persist successful payload regardless of out, so future calls hit.
	_ = c.cache.Put(cacheKey, gr.Data)
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(gr.Data, out); err != nil {
		return fmt.Errorf("decode data: %w", err)
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
