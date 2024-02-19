package render

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
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
var bannedFiles = []string{"readme.md"}

func NewPage(path string, info fs.FileInfo, inputFsys fs.FS) (*Page, error) {

	if slices.Contains[[]string](bannedFiles, strings.ToLower(info.Name())) {
		return nil, nil
	}

	var p *Page

	if info.IsDir() {
		fmt.Printf("Reading from folder: %v\n", info.Name())
		newPage, err := NewSectionPage(path, inputFsys)
		if err != nil {
			return nil, err
		}
		p = newPage
	}

	if filepath.Ext(info.Name()) == ".md" {
		fmt.Printf("Reading from file: %v\n", info.Name())
		file, err := fs.ReadFile(inputFsys, path)
		if err != nil {
			return nil, err
		}
		newPage, err := NewMarkdownPage(path, file)
		if err != nil {
			return nil, err
		}
		p = newPage
	}

	return p, nil

}

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
