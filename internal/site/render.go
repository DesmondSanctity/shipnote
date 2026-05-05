package site

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"

	"github.com/DesmondSanctity/shipnote/internal/model"
)

//go:embed templates/*.tmpl assets/*
var assetsFS embed.FS

var pageTmpl = template.Must(template.New("page").Funcs(funcMap()).ParseFS(
	assetsFS, "templates/page.tmpl", "templates/release.tmpl",
))

var homeTmpl = pageTmpl // home block lives in the same file as page

func funcMap() template.FuncMap {
	return template.FuncMap{
		"categoryLabel": categoryLabel,
		"categoryOrder": func() []model.ChangeType { return categoryRenderOrder },
		"groupedChanges": func(changes []model.Change) map[model.ChangeType][]model.Change {
			out := make(map[model.ChangeType][]model.Change, len(categoryRenderOrder))
			for _, c := range changes {
				key := c.Type
				if c.Breaking {
					key = model.ChangeTypeBreaking
				}
				out[key] = append(out[key], c)
			}
			return out
		},
		"audienceText": func(s model.AISummaries, audience string) string {
			switch audience {
			case "developer":
				if s.Developer != nil {
					return s.Developer.Text
				}
			case "customer":
				if s.Customer != nil {
					return s.Customer.Text
				}
			case "migration":
				if s.Migration != nil {
					return s.Migration.Text
				}
			}
			return ""
		},
	}
}

// categoryRenderOrder mirrors the Markdown renderer for visual parity.
var categoryRenderOrder = []model.ChangeType{
	model.ChangeTypeBreaking,
	model.ChangeTypeSecurity,
	model.ChangeTypeFeat,
	model.ChangeTypeFix,
	model.ChangeTypePerf,
	model.ChangeTypeDeprecation,
	model.ChangeTypeRefactor,
	model.ChangeTypeDocs,
	model.ChangeTypeBuild,
	model.ChangeTypeCI,
	model.ChangeTypeTest,
	model.ChangeTypeChore,
	model.ChangeTypeOther,
}

func categoryLabel(t model.ChangeType) string {
	switch t {
	case model.ChangeTypeBreaking:
		return "Breaking Changes"
	case model.ChangeTypeSecurity:
		return "Security"
	case model.ChangeTypeFeat:
		return "Added"
	case model.ChangeTypeFix:
		return "Fixed"
	case model.ChangeTypePerf:
		return "Performance"
	case model.ChangeTypeDeprecation:
		return "Deprecated"
	case model.ChangeTypeRefactor:
		return "Refactored"
	case model.ChangeTypeDocs:
		return "Documentation"
	case model.ChangeTypeBuild:
		return "Build"
	case model.ChangeTypeCI:
		return "CI"
	case model.ChangeTypeTest:
		return "Tests"
	case model.ChangeTypeChore:
		return "Chores"
	}
	return "Other"
}

// pageData is the variable surface a template needs to render a single
// release page.
type pageData struct {
	Title    string
	SiteURL  string
	Repo     model.Repo
	Release  *model.Release
	Date     string
	Sidebar  []indexEntry // all releases, newest first
	Active   string       // tag of the page being rendered ("" on home)
	IsHome   bool
	Customer string
}

func writeReleasePage(opts Options, idx *siteIndex, e indexEntry) error {
	data := pageData{
		Title:    pageTitle(opts, e.Tag),
		SiteURL:  strings.TrimRight(opts.SiteURL, "/"),
		Repo:     e.Repo,
		Release:  e.Release,
		Date:     e.Date,
		Sidebar:  idx.Entries,
		Active:   e.Tag,
		Customer: e.Customer,
	}
	var buf bytes.Buffer
	if err := pageTmpl.ExecuteTemplate(&buf, "page", data); err != nil {
		return fmt.Errorf("site: render %s: %w", e.Tag, err)
	}
	out := filepath.Join(opts.Root, e.Slug, "index.html")
	return writeFile(out, buf.Bytes())
}

func writeHomePage(opts Options, idx *siteIndex) error {
	if len(idx.Entries) == 0 {
		return nil
	}
	latest := idx.Entries[0]
	data := pageData{
		Title:    pageTitleHome(opts, latest.Repo),
		SiteURL:  strings.TrimRight(opts.SiteURL, "/"),
		Repo:     latest.Repo,
		Release:  latest.Release,
		Date:     latest.Date,
		Sidebar:  idx.Entries,
		Active:   latest.Tag,
		IsHome:   true,
		Customer: latest.Customer,
	}
	var buf bytes.Buffer
	if err := homeTmpl.ExecuteTemplate(&buf, "home", data); err != nil {
		return fmt.Errorf("site: render home: %w", err)
	}
	return writeFile(filepath.Join(opts.Root, "index.html"), buf.Bytes())
}

func pageTitle(opts Options, tag string) string {
	if opts.Title != "" {
		return fmt.Sprintf("%s — %s", opts.Title, tag)
	}
	return fmt.Sprintf("Changelog %s", tag)
}

func pageTitleHome(opts Options, r model.Repo) string {
	if opts.Title != "" {
		return opts.Title
	}
	if r.Name != "" {
		return fmt.Sprintf("%s changelog", r.Name)
	}
	return "Changelog"
}

func writeAssets(root string) error {
	entries, err := assetsFS.ReadDir("assets")
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		body, err := assetsFS.ReadFile("assets/" + e.Name())
		if err != nil {
			return err
		}
		if err := writeFile(filepath.Join(root, "assets", e.Name()), body); err != nil {
			return err
		}
	}
	return nil
}
