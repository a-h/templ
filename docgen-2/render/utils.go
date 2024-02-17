package render

import (
	"path/filepath"
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

var BaseUrl string

func titleFromPath(path string) string {
	filename, _ := baseParts(path)

	filename = strings.ReplaceAll(filename, "-", " ")

	if len(filename) > 0 {
		filename = strings.ToUpper(filename[:1]) + filename[1:]
	}

	return filename
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

var GoldmarkDefinition = goldmark.New(
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
