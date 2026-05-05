package repo

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// run executes `git <args...>` in g.dir and returns stdout. Stderr is
// folded into the error message on failure. The command always runs with
// a clean environment subset to avoid the user's GIT_* leaking in and
// breaking determinism (e.g. GIT_AUTHOR_DATE).
func (g *Git) run(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, g.gitBin, args...) //nolint:gosec // args are constructed internally, never from user input
	cmd.Dir = g.dir
	cmd.Env = sanitizedEnv()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git %s: %w: %s",
			strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}

// runLines is run() split on newlines with the trailing empty entry trimmed.
func (g *Git) runLines(ctx context.Context, args ...string) ([]string, error) {
	out, err := g.run(ctx, args...)
	if err != nil {
		return nil, err
	}
	s := strings.TrimRight(string(out), "\n")
	if s == "" {
		return nil, nil
	}
	return strings.Split(s, "\n"), nil
}

// sanitizedEnv returns the minimal environment we want git to inherit.
// We strip every GIT_* variable so author/committer overrides, alternate
// object dirs, and external SSH askpass programs cannot perturb output.
// LANG=C forces stable English error strings; PATH carries the locator
// for any helper binaries (sh, ssh) git itself may invoke.
func sanitizedEnv() []string {
	keep := map[string]string{
		"PATH": getenv("PATH"),
		"HOME": getenv("HOME"), // for ~/.gitconfig — read-only by us
		"LANG": "C",
	}
	out := make([]string, 0, len(keep))
	for k, v := range keep {
		if v == "" {
			continue
		}
		out = append(out, k+"="+v)
	}
	return out
}
