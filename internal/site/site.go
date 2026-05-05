// Package site builds and maintains a static, embeddable public
// changelog site from one or more model.Release records.
//
// Layout produced under <root>/:
//
//	<root>/
//	  index.html         (latest release + sidebar of all versions)
//	  v0.2.0/index.html
//	  v0.1.0/index.html
//	  feed.json          (JSON Feed 1.1 — source of truth for the widget)
//	  feed.xml           (Atom feed for RSS readers)
//	  releases.json      (compact index used by the site itself)
//	  assets/widget.js
//	  assets/site.css
//
// Build is idempotent: re-running with the same release set produces
// byte-identical output. Adding a new release prepends without touching
// previous version pages.
package site

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/DesmondSanctity/shipnote/internal/model"
)

// Options configures a Build.
type Options struct {
	// Root is the absolute output directory. Created if missing.
	Root string
	// SiteURL is the public URL where Root will be hosted. Required for
	// JSON Feed / Atom self-links and for the widget when it's served
	// from the same origin. May be empty for previewing locally.
	SiteURL string
	// Title overrides the default "<repo> changelog" page title.
	Title string
}

// Build writes/updates the static site for `rel`. Existing release
// entries are preserved by reading the on-disk releases.json.
func Build(opts Options, rel model.Release) error {
	if opts.Root == "" {
		return fmt.Errorf("site: Root is required")
	}
	if err := os.MkdirAll(opts.Root, 0o755); err != nil {
		return fmt.Errorf("site: mkdir %s: %w", opts.Root, err)
	}
	idx, err := loadIndex(opts.Root)
	if err != nil {
		return err
	}
	idx.upsert(rel)
	if err := writeIndex(opts.Root, idx); err != nil {
		return err
	}
	for _, e := range idx.Entries {
		if err := writeReleasePage(opts, idx, e); err != nil {
			return err
		}
	}
	if err := writeHomePage(opts, idx); err != nil {
		return err
	}
	if err := writeJSONFeed(opts, idx); err != nil {
		return err
	}
	if err := writeAtomFeed(opts, idx); err != nil {
		return err
	}
	return writeAssets(opts.Root)
}

// indexEntry is the compact, on-disk record for one published release.
// Full release.json is stored alongside so re-renders are pure.
type indexEntry struct {
	ID            string         `json:"id"`
	Tag           string         `json:"tag"`
	Date          string         `json:"date"`
	Slug          string         `json:"slug"`
	Headline      string         `json:"headline,omitempty"`
	Customer      string         `json:"customer,omitempty"`
	Repo          model.Repo     `json:"repo"`
	Stats         model.Stats    `json:"stats"`
	Release       *model.Release `json:"release"`
	ContentDigest string         `json:"contentDigest"`
}

type siteIndex struct {
	Entries []indexEntry `json:"entries"` // newest first
}

func (i *siteIndex) upsert(rel model.Release) {
	tag := strings.TrimSpace(rel.Release.Tag)
	if tag == "" {
		tag = "unreleased"
	}
	slug := slugify(tag)
	digest := contentDigest(rel)
	entry := indexEntry{
		ID:            rel.ID,
		Tag:           tag,
		Date:          firstNonEmpty(publishedAt(rel.Release.PublishedAt), rel.Range.To.Date, rel.Metadata.GeneratedAt),
		Slug:          slug,
		Headline:      rel.Summary.Headline,
		Customer:      customerText(rel.Summary.AISummaries),
		Repo:          rel.Repo,
		Stats:         rel.Stats,
		Release:       cloneRelease(rel),
		ContentDigest: digest,
	}
	for k, e := range i.Entries {
		if e.Tag == tag {
			i.Entries[k] = entry
			i.sort()
			return
		}
	}
	i.Entries = append(i.Entries, entry)
	i.sort()
}

// sort orders newest first by date desc, tag desc as tiebreaker.
func (i *siteIndex) sort() {
	sort.SliceStable(i.Entries, func(a, b int) bool {
		ai, bi := i.Entries[a], i.Entries[b]
		if ai.Date != bi.Date {
			return ai.Date > bi.Date
		}
		return ai.Tag > bi.Tag
	})
}

func loadIndex(root string) (*siteIndex, error) {
	idx := &siteIndex{}
	raw, err := os.ReadFile(filepath.Join(root, "releases.json"))
	if os.IsNotExist(err) {
		return idx, nil
	}
	if err != nil {
		return nil, fmt.Errorf("site: read releases.json: %w", err)
	}
	if err := json.Unmarshal(raw, idx); err != nil {
		return nil, fmt.Errorf("site: parse releases.json: %w", err)
	}
	return idx, nil
}

func writeIndex(root string, idx *siteIndex) error {
	body, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	return writeFile(filepath.Join(root, "releases.json"), body)
}

func writeFile(path string, body []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o644)
}

func customerText(s model.AISummaries) string {
	if s.Customer == nil {
		return ""
	}
	return strings.TrimSpace(s.Customer.Text)
}

// cloneRelease deep-copies a Release via JSON round-trip; the
// releases.json file embeds the full release so the build is
// reproducible without external state.
func cloneRelease(r model.Release) *model.Release {
	raw, err := json.Marshal(r)
	if err != nil {
		return &r
	}
	out := model.Release{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return &r
	}
	return &out
}

func contentDigest(r model.Release) string {
	raw, _ := json.Marshal(r)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func publishedAt(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func firstNonEmpty(s ...string) string {
	for _, v := range s {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
			out = append(out, c)
		case c == '.' || c == '-' || c == '_':
			out = append(out, c)
		default:
			out = append(out, '-')
		}
	}
	if len(out) == 0 {
		return "release"
	}
	return string(out)
}
