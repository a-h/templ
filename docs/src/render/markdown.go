package render

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

func NewMarkdownPage(relativePath string, file []byte) (*Page, error) {
	order, err := getOrderFromPath(relativePath)
	if err != nil {
		return nil, err
	}

	md, err := parseMarkdown(relativePath, file)
	if err != nil {
		return nil, err
	}

	html, err := renderHtml(md)
	if err != nil {
		return nil, err
	}

	slug := renderSlug(relativePath)
	title, err := renderTitleFromFileContent(relativePath, file)
	if err != nil {
		return nil, err
	}

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

func renderTitleFromFileContent(relativePath string, file []byte) (string, error) {
	var title string
	for _, line := range bytes.Split(file, []byte("\n")) {
		if !bytes.HasPrefix(line, []byte("# ")) {
			continue
		}

		title = string(bytes.TrimPrefix(line, []byte("# ")))
		if title != "" {
			return title, nil
		}
	}
	return "", fmt.Errorf("couldn't find a # heading in %s", relativePath)

}

func parseMarkdown(relativePath string, file []byte) ([]byte, error) {
	// remove frontmatter
	if strings.HasPrefix(string(file), "---") {
		_, file, _ = bytes.Cut(file[3:], []byte("---\n"))
	}

	// replace admonitions
	re := regexp.MustCompile(`:::([a-z]+)`)
	file = re.ReplaceAll(file, []byte(":::{.$1}"))

	return file, nil
}

func renderHtml(raw []byte) (string, error) {
	var htmlBuffer bytes.Buffer
	err := GoldmarkDefinition.Convert(raw, &htmlBuffer)
	if err != nil {
		return "", err
	}
	return htmlBuffer.String(), nil

}
