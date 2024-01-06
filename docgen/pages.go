package main

import (
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
)

func newPages(fsys fs.FS, root string) ([]*Page, error) {
	var pages []*Page

	entries, err := fs.ReadDir(fsys, root)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !(entry.IsDir() || filepath.Ext(entry.Name()) == ".md") {
			continue
		}

		path := filepath.Join(root, entry.Name())

		switch {
		case entry.IsDir():
			section := NewSectionPage(fsys, path)

			subPages, err := newPages(fsys, path)
			if err != nil {
				return nil, err
			}

			section.AddChildren(subPages...)

			pages = append(pages, section)
		case filepath.Ext(entry.Name()) == ".md":
			pages = append(pages, NewMarkdownPage(fsys, path))
		default:
			continue
		}
	}

	slices.SortFunc(pages, func(a, b *Page) int {
		if a.Order() == b.Order() {
			return strings.Compare(a.Title(), b.Title())
		}

		return a.Order() - b.Order()
	})

	return pages, nil
}
