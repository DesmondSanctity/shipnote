package cli

import "github.com/spf13/cobra"

// addAIFlags adds the optional --ai* flags to a command. AI is off by
// default everywhere so passing --ai is required to enable it. The
// flag set is shared between `generate` and `preview` so users get
// consistent behaviour across the two paths.
func addAIFlags(c *cobra.Command) {
	c.Flags().Bool("ai", false, "Enable AI summaries (off by default)")
	c.Flags().String("ai-provider", "auto", "AI provider: auto|openai|anthropic|groq|ollama|custom")
	c.Flags().String("ai-model", "", "Model id (provider-specific default if empty)")
	c.Flags().String("ai-base-url", "", "Base URL for --ai-provider=custom OpenAI-compatible servers")
	c.Flags().String("ai-key", "", "API key (rare; usually set via OPENAI_API_KEY / ANTHROPIC_API_KEY / GROQ_API_KEY)")
}

func readAIFlags(c *cobra.Command, opts *runOptions) {
	opts.AI = boolFlag(c, "ai")
	opts.AIProvider = stringFlag(c, "ai-provider")
	opts.AIModel = stringFlag(c, "ai-model")
	opts.AIBaseURL = stringFlag(c, "ai-base-url")
	opts.AIKey = stringFlag(c, "ai-key")
}
