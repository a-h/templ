package render

import (
	"bytes"
	"io/fs"
	"regexp"
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

type MarkdownPage struct {
	Fsys fs.FS
	Path string
}

func (p MarkdownPage) GetPath() string {
	return p.Path
}

func (p MarkdownPage) Title() string {
	return "asdf"
}

func (p MarkdownPage) Href() string {
	return "asdf"
}

func (p MarkdownPage) Render() (string, error) {
	b, err := fs.ReadFile(p.Fsys, p.Path)
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
