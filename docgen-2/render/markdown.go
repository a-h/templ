package render

import (
	"bytes"
	"io/fs"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

type MarkdownPage Page

func NewMarkdownPage(file string, inputFsys fs.FS) (*Page, error) {
	p := MarkdownPage{}
	title, err := p.renderTitle(file, inputFsys)
	if err != nil {
		return nil, err
	}
	order, err := p.renderOrder(file)
	if err != nil {
		return nil, err
	}

	md, err := p.readMarkdownFromFile(file, inputFsys)
	if err != nil {
		return nil, err
	}

	html, err := p.renderHtml(md)
	if err != nil {
		return nil, err
	}

	slug := p.renderSlug(file, inputFsys)

	page := Page{
		Path:            file,
		Type:            PageMarkdown,
		Title:           title,
		Slug:            slug,
		Href:            slug + ".html",
		Children:        nil,
		Order:           order,
		RawContent:      string(md),
		RenderedContent: html,
	}

	return &page, nil
}

func (p MarkdownPage) renderTitle(file string, inputFsys fs.FS) (string, error) {
	b, err := fs.ReadFile(inputFsys, file)
	if err != nil {
		return "", err
	}

	for _, line := range bytes.Split(b, []byte("\n")) {
		if !bytes.HasPrefix(line, []byte("# ")) {
			continue
		}

		return string(bytes.TrimPrefix(line, []byte("# "))), nil
	}

	return titleFromPath(file), nil

}

func (p MarkdownPage) renderSlug(file string, inputFsys fs.FS) string {
	if file == "index.md" {
		return "index"
	}

	noExt := strings.TrimSuffix(file, filepath.Ext(file))

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

func (p MarkdownPage) renderOrder(file string) (int, error) {
	base := filepath.Base(file)
	if base == "index.md" {
		return 0, nil
	}

	_, o := baseParts(file)
	return o, nil
}

func (p MarkdownPage) readMarkdownFromFile(file string, inputFsys fs.FS) ([]byte, error) {
	b, err := fs.ReadFile(inputFsys, file)
	if err != nil {
		return nil, err
	}

	// remove frontmatter
	if strings.HasPrefix(string(b), "---") {
		_, b, _ = bytes.Cut(b[3:], []byte("---\n"))
	}

	// replace admonitions
	re := regexp.MustCompile(`:::([a-z]+)`)
	b = re.ReplaceAll(b, []byte(":::{.$1}"))

	return b, nil
}

func (p MarkdownPage) renderHtml(raw []byte) (string, error) {
	var htmlBuffer bytes.Buffer
	err := GoldmarkDefinition.Convert(raw, &htmlBuffer)
	if err != nil {
		return "", err
	}
	return htmlBuffer.String(), nil

}
