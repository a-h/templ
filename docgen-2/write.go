package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"testing/fstest"

	"github.com/a-h/templ/docgen-2/components"
	"github.com/a-h/templ/docgen-2/render"
)

func resetOutputPath() error {
	if err := os.RemoveAll(outputPath); err != nil {
		return err
	}
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return err
	}
	return nil
}

func writeToDisk(fsyss []fs.FS) error {
	err := resetOutputPath()
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(outputPath)
	if err == nil {
		for _, entry := range entries {
			os.RemoveAll(filepath.Join(outputPath, entry.Name()))
		}
	}

	if err := os.MkdirAll(outputPath, 0o755); err != nil {
		return err
	}

	for _, fsys := range fsyss {
		if err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			f, err := fsys.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			_ = os.MkdirAll(filepath.Dir(filepath.Join(outputPath, path)), 0o755)

			out, err := os.Create(filepath.Join(outputPath, path))
			if err != nil {
				return err
			}
			defer out.Close()
			_, err = io.Copy(out, f)

			return err
		}); err != nil {
			return err
		}
	}

	return nil
}

func createMemoryFs(ctx context.Context, pages, childPages []*render.Page) (fstest.MapFS, error) {
	files := fstest.MapFS{}

	for _, page := range pages {
		switch page.Type {
		case render.PageSection:
			subFiles, err := createMemoryFs(ctx, pages, page.Children)
			if err != nil {
				return nil, err
			}
			maps.Copy(files, subFiles)
		case render.PageMarkdown:
			pc := &render.PageContext{
				Title:  page.Title,
				Active: page.Slug,
			}

			var htmlBuffer bytes.Buffer
			err := components.HTML(pc, pages, page.RenderedContent).Render(ctx, &htmlBuffer)
			if err != nil {
				return nil, err
			}

			memoryFile := &fstest.MapFile{Data: htmlBuffer.Bytes()}
			files[page.Href] = memoryFile
		}
	}

	files["search_content.js"] = &fstest.MapFile{Data: searchJS(pages)}

	return files, nil
}

func searchJS(pages []*render.Page) []byte {
	data := searchIndex(pages)

	b, _ := json.Marshal(data)

	return []byte("var index = " + string(b) + ";")
}

func searchIndex(pages []*render.Page) []map[string]string {
	var data []map[string]string

	for _, p := range pages {
		if p.Type == render.PageMarkdown {
			data = append(data, map[string]string{
				"title": p.Title,
				"href":  p.Href,
				"body":  p.RawContent,
			})
		}

		if p.Type == render.PageSection {
			data = append(data, searchIndex(p.Children)...)
		}
	}

	return data
}
