package site

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/DesmondSanctity/shipnote/internal/model"
)

// JSON Feed 1.1 — https://www.jsonfeed.org/version/1.1/
type jsonFeed struct {
	Version     string         `json:"version"`
	Title       string         `json:"title"`
	HomePageURL string         `json:"home_page_url,omitempty"`
	FeedURL     string         `json:"feed_url,omitempty"`
	Description string         `json:"description,omitempty"`
	Items       []jsonFeedItem `json:"items"`
}

type jsonFeedItem struct {
	ID            string             `json:"id"`
	URL           string             `json:"url,omitempty"`
	Title         string             `json:"title"`
	ContentText   string             `json:"content_text,omitempty"`
	ContentHTML   string             `json:"content_html,omitempty"`
	Summary       string             `json:"summary,omitempty"`
	DatePublished string             `json:"date_published,omitempty"`
	Tags          []string           `json:"tags,omitempty"`
	Extensions    *jsonFeedExtension `json:"_shipnote,omitempty"`
}

// jsonFeedExtension is a vendor extension consumed by the embeddable
// widget. Keeping it under "_shipnote" follows JSON Feed 1.1 namespace
// conventions and makes generic readers ignore it.
type jsonFeedExtension struct {
	Tag             string            `json:"tag"`
	Stats           model.Stats       `json:"stats"`
	AISummaries     model.AISummaries `json:"aiSummaries"`
	Counts          map[string]int    `json:"counts"`
	BreakingChanges int               `json:"breakingChanges"`
}

func writeJSONFeed(opts Options, idx *siteIndex) error {
	feed := jsonFeed{
		Version:     "https://jsonfeed.org/version/1.1",
		Title:       feedTitle(opts, idx),
		HomePageURL: opts.SiteURL,
		FeedURL:     joinURL(opts.SiteURL, "feed.json"),
		Description: feedDescription(idx),
	}
	for _, e := range idx.Entries {
		summary := e.Customer
		if summary == "" {
			summary = e.Headline
		}
		item := jsonFeedItem{
			ID:            stableItemID(opts.SiteURL, e),
			URL:           joinURL(opts.SiteURL, e.Slug+"/"),
			Title:         e.Tag,
			Summary:       summary,
			ContentText:   itemContentText(e),
			ContentHTML:   itemContentHTML(e),
			DatePublished: dateRFC3339(e.Date),
			Tags:          itemTags(e),
			Extensions: &jsonFeedExtension{
				Tag:             e.Tag,
				Stats:           e.Stats,
				AISummaries:     extractSummary(e),
				Counts:          changeCounts(e.Release),
				BreakingChanges: breakingCount(e.Release),
			},
		}
		feed.Items = append(feed.Items, item)
	}
	body, err := json.MarshalIndent(feed, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	return writeFile(filepath.Join(opts.Root, "feed.json"), body)
}

// Atom feed for traditional RSS readers.
type atomFeed struct {
	XMLName xml.Name   `xml:"feed"`
	XMLNS   string     `xml:"xmlns,attr"`
	Title   string     `xml:"title"`
	Link    []atomLink `xml:"link"`
	ID      string     `xml:"id"`
	Updated string     `xml:"updated"`
	Entries []atomEntry
}

type atomLink struct {
	Rel  string `xml:"rel,attr,omitempty"`
	Href string `xml:"href,attr"`
	Type string `xml:"type,attr,omitempty"`
}

type atomEntry struct {
	XMLName   xml.Name  `xml:"entry"`
	Title     string    `xml:"title"`
	ID        string    `xml:"id"`
	Updated   string    `xml:"updated"`
	Published string    `xml:"published,omitempty"`
	Link      atomLink  `xml:"link"`
	Summary   atomBody  `xml:"summary"`
	Content   *atomBody `xml:"content,omitempty"`
}

type atomBody struct {
	Type string `xml:"type,attr"`
	Body string `xml:",chardata"`
}

func writeAtomFeed(opts Options, idx *siteIndex) error {
	updated := dateRFC3339("")
	if len(idx.Entries) > 0 {
		updated = dateRFC3339(idx.Entries[0].Date)
	}
	feed := atomFeed{
		XMLNS:   "http://www.w3.org/2005/Atom",
		Title:   feedTitle(opts, idx),
		ID:      firstNonEmpty(opts.SiteURL, "urn:shipnote:"+feedTitleSlug(opts, idx)),
		Updated: updated,
		Link: []atomLink{
			{Rel: "alternate", Type: "text/html", Href: opts.SiteURL},
			{Rel: "self", Type: "application/atom+xml", Href: joinURL(opts.SiteURL, "feed.xml")},
		},
	}
	for _, e := range idx.Entries {
		summary := e.Customer
		if summary == "" {
			summary = e.Headline
		}
		feed.Entries = append(feed.Entries, atomEntry{
			Title:     e.Tag,
			ID:        stableItemID(opts.SiteURL, e),
			Updated:   dateRFC3339(e.Date),
			Published: dateRFC3339(e.Date),
			Link:      atomLink{Rel: "alternate", Type: "text/html", Href: joinURL(opts.SiteURL, e.Slug+"/")},
			Summary:   atomBody{Type: "text", Body: summary},
		})
	}
	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(&buf)
	enc.Indent("", "  ")
	if err := enc.Encode(feed); err != nil {
		return err
	}
	if err := enc.Flush(); err != nil {
		return err
	}
	buf.WriteByte('\n')
	return writeFile(filepath.Join(opts.Root, "feed.xml"), buf.Bytes())
}

func feedTitle(opts Options, idx *siteIndex) string {
	if opts.Title != "" {
		return opts.Title
	}
	if len(idx.Entries) > 0 && idx.Entries[0].Repo.Name != "" {
		return fmt.Sprintf("%s changelog", idx.Entries[0].Repo.Name)
	}
	return "Changelog"
}

func feedTitleSlug(opts Options, idx *siteIndex) string {
	return slugify(feedTitle(opts, idx))
}

func feedDescription(idx *siteIndex) string {
	if len(idx.Entries) == 0 {
		return ""
	}
	r := idx.Entries[0].Repo
	if r.Name == "" {
		return ""
	}
	return fmt.Sprintf("Release notes for %s/%s", r.Owner, r.Name)
}

func itemContentText(e indexEntry) string {
	var sb strings.Builder
	if e.Customer != "" {
		sb.WriteString(e.Customer)
		sb.WriteString("\n\n")
	} else if e.Headline != "" {
		sb.WriteString(e.Headline)
		sb.WriteString("\n\n")
	}
	if e.Release == nil {
		return strings.TrimSpace(sb.String())
	}
	for _, c := range e.Release.GlobalChanges {
		sb.WriteString("- ")
		sb.WriteString(c.Title)
		sb.WriteByte('\n')
	}
	return strings.TrimSpace(sb.String())
}

func itemContentHTML(e indexEntry) string {
	var sb strings.Builder
	if e.Customer != "" {
		sb.WriteString("<p>")
		sb.WriteString(escapeHTML(e.Customer))
		sb.WriteString("</p>")
	} else if e.Headline != "" {
		sb.WriteString("<p>")
		sb.WriteString(escapeHTML(e.Headline))
		sb.WriteString("</p>")
	}
	if e.Release == nil {
		return sb.String()
	}
	sb.WriteString("<ul>")
	for _, c := range e.Release.GlobalChanges {
		sb.WriteString("<li>")
		sb.WriteString(escapeHTML(c.Title))
		sb.WriteString("</li>")
	}
	sb.WriteString("</ul>")
	return sb.String()
}

func itemTags(e indexEntry) []string {
	if e.Release == nil {
		return nil
	}
	tags := make([]string, 0, 4)
	if breakingCount(e.Release) > 0 {
		tags = append(tags, "breaking")
	}
	counts := changeCounts(e.Release)
	for _, t := range []string{"feat", "fix", "perf", "security"} {
		if counts[t] > 0 {
			tags = append(tags, t)
		}
	}
	return tags
}

func breakingCount(r *model.Release) int {
	if r == nil {
		return 0
	}
	return len(r.BreakingChanges)
}

func changeCounts(r *model.Release) map[string]int {
	if r == nil {
		return map[string]int{}
	}
	counts := map[string]int{}
	for _, c := range r.GlobalChanges {
		counts[string(c.Type)]++
	}
	return counts
}

func extractSummary(e indexEntry) model.AISummaries {
	if e.Release == nil {
		return model.AISummaries{}
	}
	return e.Release.Summary.AISummaries
}

func stableItemID(siteURL string, e indexEntry) string {
	if siteURL != "" {
		return joinURL(siteURL, e.Slug+"/")
	}
	return "urn:shipnote:" + e.Slug
}

func joinURL(base, suffix string) string {
	if base == "" {
		return suffix
	}
	base = strings.TrimRight(base, "/")
	suffix = strings.TrimLeft(suffix, "/")
	return base + "/" + path.Clean(suffix)
}

func dateRFC3339(s string) string {
	if s == "" {
		return time.Time{}.UTC().Format(time.RFC3339)
	}
	// Accept either RFC3339 or YYYY-MM-DD.
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC().Format(time.RFC3339)
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t.UTC().Format(time.RFC3339)
	}
	return s
}

func escapeHTML(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "\"", "&quot;", "'", "&#39;")
	return r.Replace(s)
}
