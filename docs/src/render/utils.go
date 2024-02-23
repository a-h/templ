package render

import (
	"fmt"
	"io/fs"
	"path"
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

func getOrderFromPath(path string) (int, error) {
	base := filepath.Base(path)

	if base == "index.md" {
		return 1, nil
	}

	index := strings.Split(base, "-")[0]

	order, err := strconv.Atoi(index)
	if err != nil {
		return -1, fmt.Errorf("%v is not a valid file or folder, it must follow convention: 01-folder/ or 01-file-name.md or be index.md", path)
	}
	return order, nil
}

func renderSlug(relativePath string) string {
	if relativePath == "index.md" {
		return "index"
	}

	noExt := strings.TrimSuffix(relativePath, filepath.Ext(relativePath))

	htmlPath := ""

	for _, r := range strings.Split(noExt, "/") {
		base := filepath.Base(r)
		filename := base[:len(base)-len(filepath.Ext(base))]
		_, partialSlug, _ := strings.Cut(filename, "-")
		htmlPath = path.Join(htmlPath, partialSlug)
	}

	for _, r := range []string{"\\", " ", ".", "_"} {
		htmlPath = strings.ReplaceAll(htmlPath, r, "-")
	}

	htmlPath = strings.ToLower(htmlPath)
	htmlPath = strings.Trim(htmlPath, "-")

	return htmlPath
}

func getTitleFromSlug(slug string) string {
	base := strings.LastIndex(slug, "/")
	if base > 0 {
		slug = slug[base+1:]
	}

	title := strings.ReplaceAll(slug, "-", " ")
	words := strings.Fields(title)
	for i, word := range words {
		if len(word) > 1 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(string(word[1:]))
		} else {
			words[i] = strings.ToUpper(word)
		}
	}

	title = strings.Join(words, " ")
	return title

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
