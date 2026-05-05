package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/DesmondSanctity/shipnote/internal/config"
	"github.com/DesmondSanctity/shipnote/internal/repo"
)

func runDoctor(cmd *cobra.Command, _ []string) error {
	w := cmd.OutOrStdout()
	ctx := cmd.Context()

	root, err := resolveRepoRoot()
	if err != nil {
		return err
	}
	checkLine(w, "cwd", root, nil)

	g, err := repo.Open(ctx, root)
	checkLine(w, "git repo", root, err)
	if err != nil {
		return nil
	}

	url, err := g.OriginURL(ctx)
	switch {
	case err != nil:
		checkLine(w, "git remote", "", err)
	case url == "":
		checkLine(w, "git remote", "(none configured)", nil)
	default:
		owner, name, ok := repo.ParseGitHubRemote(url)
		if !ok {
			checkLine(w, "git remote", url+" (not a github.com URL)", nil)
		} else {
			checkLine(w, "git remote", fmt.Sprintf("%s (%s/%s)", url, owner, name), nil)
		}
	}

	rng, err := g.ResolveRange(ctx, "", "HEAD", "v*")
	if err != nil {
		checkLine(w, "range", "", err)
	} else {
		fromLabel := rng.From.Ref
		if fromLabel == rng.From.SHA {
			fromLabel = rng.From.SHA[:7] + " (root)"
		}
		checkLine(w, "range", fmt.Sprintf("%s -> %s", fromLabel, rng.To.Ref), nil)
	}

	cfgPath := config.ResolvedPath(root, stringFlag(cmd.Root(), "config"))
	cfg, err := config.Load(cfgPath)
	if err != nil {
		checkLine(w, "config", cfgPath, err)
	} else {
		if _, statErr := os.Stat(cfgPath); statErr == nil {
			checkLine(w, "config", cfgPath, nil)
		} else {
			checkLine(w, "config", "(none — using defaults)", nil)
		}
		_ = cfg
	}

	token := defaultToken("")
	if token == "" {
		checkLine(w, "github token", "(none — anonymous; rate-limited to ~60 req/h)", nil)
	} else {
		login, remaining, err := probeGitHub(ctx, token)
		if err != nil {
			checkLine(w, "github token", "set", err)
		} else {
			checkLine(w, "github token", fmt.Sprintf("set (login=%s, rate-limit-remaining=%d)", login, remaining), nil)
		}
	}
	return nil
}

func checkLine(w io.Writer, label, value string, err error) {
	mark := "ok  "
	if err != nil {
		mark = "FAIL"
		value = err.Error()
	}
	_, _ = fmt.Fprintf(w, "[%s] %-14s %s\n", mark, label, value)
}

// probeGitHub does a single REST call to /user (no scopes required for
// the bare login + rate-limit headers). Returns the authenticated
// login and the X-RateLimit-Remaining value.
func probeGitHub(ctx context.Context, token string) (login string, remaining int, err error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "shipnote-doctor")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	// Cheap "json" extraction without pulling in a struct: look for the
	// "login":"..." pair. Doctor is best-effort.
	if idx := strings.Index(string(body), `"login":"`); idx >= 0 {
		rest := string(body)[idx+len(`"login":"`):]
		if end := strings.IndexByte(rest, '"'); end > 0 {
			login = rest[:end]
		}
	}
	if h := resp.Header.Get("X-RateLimit-Remaining"); h != "" {
		_, _ = fmt.Sscanf(h, "%d", &remaining)
	}
	return login, remaining, nil
}
