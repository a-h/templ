package render

import (
	"bytes"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

type MarkdownPage Page

func NewMarkdownPage(relativePath string, file []byte) (*Page, error) {
	p := MarkdownPage{}
	title, err := p.renderTitle(relativePath, file)
	if err != nil {
		return nil, err
	}
	order, err := p.renderOrder(relativePath)
	if err != nil {
		return nil, err
	}

	md, err := p.parseMarkdown(relativePath, file)
	if err != nil {
		return nil, err
	}

	html, err := p.renderHtml(md)
	if err != nil {
		return nil, err
	}

	slug := p.renderSlug(relativePath)

	page := Page{
		Path:            relativePath,
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

func (p MarkdownPage) renderTitle(relativePath string, file []byte) (string, error) {
	for _, line := range bytes.Split(file, []byte("\n")) {
		if !bytes.HasPrefix(line, []byte("# ")) {
			continue
		}

		return string(bytes.TrimPrefix(line, []byte("# "))), nil
	}

	return titleFromPath(relativePath), nil

}

func (p MarkdownPage) renderSlug(relativePath string) string {
	if relativePath == "index.md" {
		return "index"
	}

	noExt := strings.TrimSuffix(relativePath, filepath.Ext(relativePath))

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

func (p MarkdownPage) renderOrder(relativePath string) (int, error) {
	base := filepath.Base(relativePath)
	if base == "index.md" {
		return 0, nil
	}

	_, o := baseParts(relativePath)
	return o, nil
}

func (p MarkdownPage) parseMarkdown(relativePath string, file []byte) ([]byte, error) {
	// remove frontmatter
	if strings.HasPrefix(string(file), "---") {
		_, file, _ = bytes.Cut(file[3:], []byte("---\n"))
	}

	// replace admonitions
	re := regexp.MustCompile(`:::([a-z]+)`)
	file = re.ReplaceAll(file, []byte(":::{.$1}"))

	return file, nil
}

func (p MarkdownPage) renderHtml(raw []byte) (string, error) {
	var htmlBuffer bytes.Buffer
	err := GoldmarkDefinition.Convert(raw, &htmlBuffer)
	if err != nil {
		return "", err
	}
	return htmlBuffer.String(), nil

}
