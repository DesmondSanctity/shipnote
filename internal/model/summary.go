package model

// Summary holds the canonical human-facing copy plus AI variants.
// Headline is the canonical, non-AI-derived one-line summary.
type Summary struct {
	Headline    string      `json:"headline"`
	AISummaries AISummaries `json:"aiSummaries"`
}

// AISummaries holds audience-keyed AI-generated summaries. All three
// audience keys are required by the schema; nil values marshal as JSON null
// to indicate "not generated".
type AISummaries struct {
	Developer *AISummary `json:"developer"`
	Customer  *AISummary `json:"customer"`
	Migration *AISummary `json:"migration"`
}

// AISummary is a single AI-generated summary plus the prompt/model
// fingerprint that produced it. PromptVersion is a stable label
// (e.g. "v1") that pins the prompt template used.
type AISummary struct {
	Text          string `json:"text"`
	Model         string `json:"model"`
	PromptVersion string `json:"promptVersion"`
}
