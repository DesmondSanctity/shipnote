package github

import (
	"net/http"
	"strings"
	"time"
)

// DefaultEndpoint is the public github.com GraphQL endpoint. Override
// via Options.Endpoint to point at GHE or a recorded fixture server.
const DefaultEndpoint = "https://api.github.com/graphql"

// DefaultUserAgent is sent on every request. GitHub requires a
// User-Agent and uses it for rate-limit attribution.
const DefaultUserAgent = "shipnote/0.1 (+https://github.com/DesmondSanctity/shipnote)"

// Options configures Client. The zero value is valid for anonymous use
// against api.github.com.
type Options struct {
	Endpoint   string
	Token      string // bearer token; empty = anonymous
	UserAgent  string
	HTTPClient *http.Client
	Timeout    time.Duration // per-request timeout; default 30s
}

// Client talks to a GitHub-compatible GraphQL (and REST, when needed)
// endpoint. Safe for concurrent use by multiple goroutines.
type Client struct {
	endpoint  string
	token     string
	userAgent string
	hc        *http.Client
}

// New constructs a Client from Options, applying defaults.
func New(opts Options) *Client {
	endpoint := strings.TrimSpace(opts.Endpoint)
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	ua := strings.TrimSpace(opts.UserAgent)
	if ua == "" {
		ua = DefaultUserAgent
	}
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	hc := opts.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: timeout}
	}
	return &Client{
		endpoint:  endpoint,
		token:     strings.TrimSpace(opts.Token),
		userAgent: ua,
		hc:        hc,
	}
}

// Anonymous reports whether the client is operating without a token.
func (c *Client) Anonymous() bool { return c.token == "" }
