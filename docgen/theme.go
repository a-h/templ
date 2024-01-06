package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
	"maps"
	"testing/fstest"
)

type pageContext struct {
	Title  string
	Active string
}

func render(ctx context.Context, baseURL string, pages []*Page) (fs.FS, error) {
	files, err := toFiles(ctx, baseURL, pages, pages)
	if err != nil {
		return nil, err
	}

	files["search_content.js"] = &fstest.MapFile{Data: searchJS(pages)}

	return files, nil
}

func searchJS(pages []*Page) []byte {
	data := searchIndex(pages)

	b, _ := json.Marshal(data)

	return []byte("var index = " + string(b) + ";")
}

func searchIndex(pages []*Page) []map[string]string {
	var data []map[string]string

	for _, p := range pages {
		if p.Type() == Markdown {
			md, err := p.Markdown()
			if err != nil {
				continue
			}

			data = append(data, map[string]string{
				"title": p.Title(),
				"href":  p.Href(),
				"body":  md,
			})
		}

		if p.Type() == Section {
			data = append(data, searchIndex(p.Children())...)
		}
	}

	return data
}

func toFiles(ctx context.Context, baseURL string, pages, childPages []*Page) (fstest.MapFS, error) {
	files := fstest.MapFS{}

	for _, p := range childPages {
		switch p.Type() {
		case Section:
			subFiles, err := toFiles(ctx, baseURL, pages, p.Children())
			if err != nil {
				return nil, err
			}

			maps.Copy(files, subFiles)
		case Markdown:
			memoryFile, err := toFile(ctx, baseURL, pages, p)
			if err != nil {
				return nil, err
			}

			files[p.Href()] = memoryFile
		}
	}

	return files, nil
}

func toFile(ctx context.Context, baseURL string, pages []*Page, p *Page) (*fstest.MapFile, error) {
	mainContent, err := p.HTML()
	if err != nil {
		return nil, err
	}

	pc := &pageContext{
		Title:  p.Title(),
		Active: p.Slug(),
	}

	var htmlBuffer bytes.Buffer
	if err := HTML(pc, baseURL, pages, mainContent).Render(ctx, &htmlBuffer); err != nil {
		return nil, err
	}

	return &fstest.MapFile{Data: htmlBuffer.Bytes()}, nil
}
