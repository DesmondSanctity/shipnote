package ai

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// cacheEntry is what we serialise to disk. The PromptHash is recomputed
// at lookup time and asserted against the file's contents to detect
// silent template drift.
type cacheEntry struct {
	Text       string `json:"text"`
	Model      string `json:"model"`
	PromptHash string `json:"promptHash"`
	Audience   string `json:"audience"`
	Scope      string `json:"scope"`
}

// Cache stores AI responses keyed by (provider, model, promptVersion,
// canonicalFacts). Entries are JSON files; the cache is portable and
// can be committed to a workflow cache without leaking secrets.
type Cache struct {
	Dir string
}

// Lookup returns a hit if and only if the cache file exists and its
// PromptHash matches the supplied promptHash (i.e. prompt template
// hasn't drifted under us since the entry was written).
func (c *Cache) Lookup(providerName, model string, in SummarizeInput, promptHash string) (SummarizeOutput, bool) {
	if c == nil || c.Dir == "" {
		return SummarizeOutput{}, false
	}
	path := c.path(providerName, model, in, promptHash)
	data, err := os.ReadFile(path)
	if err != nil {
		return SummarizeOutput{}, false
	}
	var e cacheEntry
	if err := json.Unmarshal(data, &e); err != nil {
		return SummarizeOutput{}, false
	}
	if e.PromptHash != promptHash {
		return SummarizeOutput{}, false
	}
	return SummarizeOutput{Text: e.Text, Model: e.Model, PromptHash: e.PromptHash}, true
}

// Save writes an entry; any error is non-fatal (we still return the
// result to the caller, just without persistence).
func (c *Cache) Save(providerName, model string, in SummarizeInput, out SummarizeOutput) error {
	if c == nil || c.Dir == "" {
		return nil
	}
	path := c.path(providerName, model, in, out.PromptHash)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(cacheEntry{
		Text:       out.Text,
		Model:      out.Model,
		PromptHash: out.PromptHash,
		Audience:   string(in.Audience),
		Scope:      in.Scope,
	})
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// path returns the cache filename. Two requests with the same
// (provider, model, promptVersion, canonical facts, audience) produce
// the same path.
func (c *Cache) path(providerName, model string, in SummarizeInput, promptHash string) string {
	canonical, _ := canonicalJSON(in.Facts)
	keyMaterial := strings.Join([]string{
		providerName, model, PromptVersion,
		string(in.Audience), in.Scope, promptHash, canonical,
	}, "|")
	sum := sha256.Sum256([]byte(keyMaterial))
	hex := hex.EncodeToString(sum[:])
	// Two-level directory shard keeps the directory small on busy CI.
	return filepath.Join(c.Dir, hex[:2], fmt.Sprintf("%s.json", hex))
}
