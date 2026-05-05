package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/DesmondSanctity/shipnote/internal/ai"
	"github.com/DesmondSanctity/shipnote/internal/config"
	"github.com/DesmondSanctity/shipnote/internal/runner"
)

// runOptions captures the flags shared by `generate` and `preview`.
type runOptions struct {
	FromRef    string
	ToRef      string
	ConfigPath string
	Token      string
	CacheDir   string
	NoNetwork  bool

	AI         bool   // --ai
	AIProvider string // --ai-provider: auto|openai|anthropic|groq|ollama|custom
	AIModel    string // --ai-model
	AIBaseURL  string // --ai-base-url for custom OpenAI-compatible servers
	AIKey      string // --ai-key (rare; usually env)
}

func resolveRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return cwd, nil
}

func loadConfigForRoot(root, configPath string) (config.Config, string, error) {
	path := config.ResolvedPath(root, configPath)
	cfg, err := config.Load(path)
	return cfg, path, err
}

func defaultToken(flagToken string) string {
	if flagToken != "" {
		return flagToken
	}
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return t
	}
	return os.Getenv("GH_TOKEN")
}

func defaultCacheDir(root string) string {
	return filepath.Join(root, ".shipnote", "cache")
}

func executeRun(ctx context.Context, opts runOptions) (runner.Outputs, config.Config, string, error) {
	root, err := resolveRepoRoot()
	if err != nil {
		return runner.Outputs{}, config.Config{}, "", err
	}
	cfg, _, err := loadConfigForRoot(root, opts.ConfigPath)
	if err != nil {
		return runner.Outputs{}, cfg, root, err
	}
	in := runner.Inputs{
		RepoDir:   root,
		FromRef:   opts.FromRef,
		ToRef:     opts.ToRef,
		Token:     defaultToken(opts.Token),
		Cfg:       cfg,
		CacheDir:  defaultCacheDir(root),
		NoNetwork: opts.NoNetwork,
	}
	if opts.AI {
		prov := resolveAIProvider(opts)
		if prov != nil {
			in.AIProvider = prov
			in.AICache = &ai.Cache{Dir: filepath.Join(root, ".shipnote", "cache", "ai")}
			in.AILogger = func(msg string) { _, _ = fmt.Fprintln(os.Stderr, msg) }
		}
	}
	out, err := runner.Run(ctx, in)
	return out, cfg, root, err
}

// resolveAIProvider builds a Summarizer from the user flags. Returns
// nil with a warning when the requested provider has no usable key /
// endpoint; the runner treats nil as 'AI off' which keeps the
// deterministic baseline intact.
func resolveAIProvider(opts runOptions) ai.Summarizer {
	provider := strings.ToLower(opts.AIProvider)
	if provider == "" {
		provider = "auto"
	}
	switch provider {
	case "openai":
		key := firstNonEmpty(opts.AIKey, os.Getenv("OPENAI_API_KEY"))
		if key == "" {
			_, _ = fmt.Fprintln(os.Stderr, "shipnote: --ai-provider=openai requires OPENAI_API_KEY")
			return nil
		}
		return ai.NewOpenAI(key, opts.AIModel)
	case "anthropic", "claude":
		key := firstNonEmpty(opts.AIKey, os.Getenv("ANTHROPIC_API_KEY"))
		if key == "" {
			_, _ = fmt.Fprintln(os.Stderr, "shipnote: --ai-provider=anthropic requires ANTHROPIC_API_KEY")
			return nil
		}
		return ai.NewAnthropic(key, opts.AIModel)
	case "groq":
		key := firstNonEmpty(opts.AIKey, os.Getenv("GROQ_API_KEY"))
		if key == "" {
			_, _ = fmt.Fprintln(os.Stderr, "shipnote: --ai-provider=groq requires GROQ_API_KEY")
			return nil
		}
		return ai.NewGroq(key, opts.AIModel)
	case "ollama":
		model := opts.AIModel
		if model == "" {
			model = "llama3.2:1b"
		}
		return ai.NewOllama(model)
	case "custom":
		if opts.AIBaseURL == "" {
			_, _ = fmt.Fprintln(os.Stderr, "shipnote: --ai-provider=custom requires --ai-base-url")
			return nil
		}
		return ai.NewOpenAICompatible(ai.OpenAICompatibleConfig{
			Name: "custom", BaseURL: opts.AIBaseURL, APIKey: opts.AIKey, Model: opts.AIModel,
		})
	case "auto":
		// Detection chain: explicit key beats local server.
		if key := os.Getenv("OPENAI_API_KEY"); key != "" {
			return ai.NewOpenAI(key, opts.AIModel)
		}
		if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
			return ai.NewAnthropic(key, opts.AIModel)
		}
		if key := os.Getenv("GROQ_API_KEY"); key != "" {
			return ai.NewGroq(key, opts.AIModel)
		}
		model := opts.AIModel
		if model == "" {
			model = "llama3.2:1b"
		}
		return ai.NewOllama(model)
	default:
		_, _ = fmt.Fprintf(os.Stderr, "shipnote: unknown --ai-provider %q\n", opts.AIProvider)
		return nil
	}
}

func firstNonEmpty(s ...string) string {
	for _, v := range s {
		if v != "" {
			return v
		}
	}
	return ""
}

// stringFlag pulls a string flag, panicking on developer error
// (missing flag definitions). Keeps callers concise.
func stringFlag(c *cobra.Command, name string) string {
	v, _ := c.Flags().GetString(name)
	return v
}

func boolFlag(c *cobra.Command, name string) bool {
	v, _ := c.Flags().GetBool(name)
	return v
}

func writeArtifact(absPath string, body []byte) error {
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(absPath), err)
	}
	if err := os.WriteFile(absPath, body, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", absPath, err)
	}
	return nil
}

func absUnder(root, rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Join(root, rel)
}
