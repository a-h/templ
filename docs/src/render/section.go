package render

import (
	"io/fs"
)

func NewSectionPage(folder string, inputFsys fs.FS) (*Page, error) {
	order, err := getOrderFromPath(folder)
	if err != nil {
		return nil, err
	}

	children, err := renderChildren(folder, inputFsys)
	if err != nil {
		return nil, err
	}

	slug := renderSlug(folder)
	title := getTitleFromSlug(slug)

	page := Page{
		Path:     folder,
		Title:    title,
		Slug:     slug,
		Type:     PageSection,
		Order:    order,
		Children: children,
	}
	return &page, nil

}

func renderChildren(folder string, inputFsys fs.FS) ([]*Page, error) {
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
