package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	fences "github.com/stefanfritsch/goldmark-fences"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/anchor"
	"go.abhg.dev/goldmark/mermaid"
	"mvdan.cc/xurls/v2"
)

type PageType int

const (
	Section PageType = iota
	Markdown
)

type Page struct {
	fsys     fs.FS
	path     string
	pageType PageType

	children []*Page
}

func NewMarkdownPage(fsys fs.FS, path string) *Page {
	return &Page{
		fsys:     fsys,
		path:     path,
		pageType: Markdown,
	}
}

func NewSectionPage(fsys fs.FS, path string) *Page {
	return &Page{
		fsys:     fsys,
		path:     path,
		pageType: Section,
	}
}

func (p *Page) Type() PageType {
	return p.pageType
}

func (p *Page) Title() string {
	if p.pageType == Section {
		b, err := fs.ReadFile(p.fsys, filepath.Join(p.path, "_category_.json"))
		if err != nil {
			return titleFromPath(p.path)
		}

		var category struct {
			Label string `json:"label"`
		}

		if err := json.Unmarshal(b, &category); err != nil {
			return titleFromPath(p.path)
		}

		if category.Label != "" {
			return category.Label
		}

		return titleFromPath(p.path)
	}

	b, err := fs.ReadFile(p.fsys, p.path)
	if err != nil {
		return titleFromPath(p.path)
	}

	for _, line := range bytes.Split(b, []byte("\n")) {
		if !bytes.HasPrefix(line, []byte("# ")) {
			continue
		}

		return string(bytes.TrimPrefix(line, []byte("# ")))
	}

	return titleFromPath(p.path)
}

func titleFromPath(path string) string {
	filename, _ := baseParts(path)

	filename = strings.ReplaceAll(filename, "-", " ")

	if len(filename) > 0 {
		filename = strings.ToUpper(filename[:1]) + filename[1:]
	}

	return filename
}

func (p *Page) Order() int {
	if p.pageType == Section {
		b, err := fs.ReadFile(p.fsys, filepath.Join(p.path, "_category_.json"))
		if err != nil {
			return orderFromPath(p)
		}

		var category struct {
			Position int `json:"position"`
		}

		if err := json.Unmarshal(b, &category); err != nil {
			return orderFromPath(p)
		}

		return category.Position
	}

	base := filepath.Base(p.path)
	if base == "index.md" {
		return 0
	}

	return orderFromPath(p)
}

func orderFromPath(p *Page) int {
	_, o := baseParts(p.path)

	return o
}

func baseParts(path string) (string, int) {
	base := filepath.Base(path)
	filename := base[:len(base)-len(filepath.Ext(base))]
	prefix, suffix, hasSpace := strings.Cut(filename, "-")

	if hasSpace {
		prefix := strings.TrimPrefix(prefix, "0")
		o, err := strconv.Atoi(prefix)
		if err != nil {
			return suffix, -1
		}

		return suffix, o
	}

	return filename, -1
}

func (p *Page) Slug() string {
	if p.path == "index.md" {
		return "index"
	}

	noExt := strings.TrimSuffix(p.path, filepath.Ext(p.path))

	htmlPath := ""

	for _, r := range strings.Split(noExt, "/") {
		name, _ := baseParts(r)
		htmlPath = path.Join(htmlPath, name)
	}

	for _, r := range []string{"\\", " ", ".", "_"} {
		htmlPath = strings.ReplaceAll(htmlPath, r, "-")
	}

	htmlPath = strings.ToLower(htmlPath)
	htmlPath = strings.Trim(htmlPath, "-")

	return htmlPath
}

func (p *Page) Href() string {
	return p.Slug() + ".html"
}

func (p *Page) Markdown() (string, error) {
	if p.pageType != Markdown {
		return "", errors.New("page is not markdown")
	}

	b, err := fs.ReadFile(p.fsys, p.path)
	if err != nil {
		return "", err
	}

	// remove frontmatter
	if strings.HasPrefix(string(b), "---") {
		_, b, _ = bytes.Cut(b[3:], []byte("---\n"))
	}

	// replace admonitions
	re := regexp.MustCompile(`:::([a-z]+)`)
	b = re.ReplaceAll(b, []byte(":::{.$1}"))

	return string(b), nil
}

func (p *Page) HTML() (string, error) {
	b, err := p.Markdown()
	if err != nil {
		return "", err
	}

	var htmlBuffer bytes.Buffer

	markdown := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithXHTML(),
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			&anchor.Extender{
				Texter: anchor.Text("#"),
			},
			extension.NewLinkify(
				extension.WithLinkifyAllowedProtocols([][]byte{
					[]byte("http:"),
					[]byte("https:"),
				}),
				extension.WithLinkifyURLRegexp(
					xurls.Strict(),
				),
			),
			&mermaid.Extender{},
			&fences.Extender{},
		),
	)

	if err := markdown.Convert([]byte(b), &htmlBuffer); err != nil {
		return "", err
	}

	return htmlBuffer.String(), nil
}

func (p *Page) AddChildren(child ...*Page) {
	p.children = append(p.children, child...)
}

func (p *Page) Children() []*Page {
	return p.children
}
