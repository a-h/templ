package render

import (
	"encoding/json"
	"io/fs"
	"path"
	"path/filepath"
	"strings"
)

type SectionPage Page

func NewSectionPage(folder string, inputFsys fs.FS) (*Page, error) {
	p := SectionPage{}
	order, err := p.renderOrder(folder, inputFsys)
	if err != nil {
		return nil, err
	}

	children, err := p.renderChildren(folder, inputFsys)
	if err != nil {
		return nil, err
	}

	page := Page{
		Path:     folder,
		Type:     PageSection,
		Title:    p.renderTitle(folder, inputFsys),
		Slug:     p.renderSlug(folder),
		Href:     p.renderSlug(folder) + ".html",
		Order:    order,
		Children: children,
	}
	return &page, nil

}

func (p SectionPage) renderTitle(folder string, inputFsys fs.FS) string {
	b, err := fs.ReadFile(inputFsys, filepath.Join(folder, "_category_.json"))
	if err != nil {
		return titleFromPath(folder)
	}

	var category struct {
		Label string `json:"label"`
	}

	if err := json.Unmarshal(b, &category); err != nil {
		return titleFromPath(folder)
	}

	if category.Label != "" {
		return category.Label
	}

	return titleFromPath(folder)
}

func (p SectionPage) renderSlug(folder string) string {
	htmlPath := ""

	for _, r := range strings.Split(htmlPath, "/") {
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

func (p SectionPage) renderOrder(folder string, inputFsys fs.FS) (int, error) {
	b, err := fs.ReadFile(inputFsys, filepath.Join(folder, "_category_.json"))
	if err != nil {
		_, o := baseParts(folder)
		return o, nil
	}

	var category struct {
		Position int `json:"position"`
	}

	if err := json.Unmarshal(b, &category); err != nil {
		_, o := baseParts(folder)
		return o, nil
	}

	return category.Position, nil
}

func (p SectionPage) renderChildren(folder string, inputFsys fs.FS) ([]*Page, error) {
	var pages []*Page
	entries, err := fs.ReadDir(inputFsys, folder)

	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}

		p, err := NewPage(folder+"/"+entry.Name(), info, inputFsys)
		if err != nil {
			return nil, err
		}
		if p == nil {
			continue
		}
		pages = append(pages, p)
	}
	return pages, nil
}
