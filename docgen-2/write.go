package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing/fstest"

	"github.com/a-h/templ/docgen-2/components"
	"github.com/a-h/templ/docgen-2/render"
)

func resetOutputPath() error {
	fmt.Printf("Deleteing folder %v\n", outputPath)
	if err := os.RemoveAll(outputPath); err != nil {
		return err
	}
	fmt.Printf("Creating folder %v\n", outputPath)
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

	for _, fsys := range fsyss {

		// fmt.Printf("Opening in-memory filesystem %#v\n", fsys)
		err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			// fmt.Printf("Opening %v within fsys %#v\n", path, fsys)
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
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func createMemoryFs(ctx context.Context, allPages, pagesToRender []*render.Page) (fstest.MapFS, error) {
	files := fstest.MapFS{}

	for _, page := range pagesToRender {
		if page.Type == render.PageSection {
			fmt.Printf("Creating section slug: %v \n", page.Slug)
			subFiles, err := createMemoryFs(ctx, allPages, page.Children)
			if err != nil {
				return nil, err
			}
			maps.Copy(files, subFiles)
		}
		if page.Type == render.PageMarkdown {
			fmt.Printf("Creating page slug: %v \n", page.Slug)
			pc := &render.PageContext{
				Title:  page.Title,
				Active: page.Slug,
			}

			var htmlBuffer bytes.Buffer
			err := components.HTML(pc, allPages, page.RenderedContent).Render(ctx, &htmlBuffer)
			if err != nil {
				return nil, err
			}

			memoryFile := &fstest.MapFile{Data: htmlBuffer.Bytes()}
			location := strings.TrimPrefix(page.Href, "docs/")
			files[location] = memoryFile
		}

	}

	files["search_content.js"] = &fstest.MapFile{Data: searchJS(allPages)}

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
