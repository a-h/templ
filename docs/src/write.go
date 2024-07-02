package src

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
	"testing/fstest"

	"github.com/a-h/templ/docs/src/components"
	"github.com/a-h/templ/docs/src/render"
)

func WriteToDisk(fsyss []fs.FS, outputPath string) error {
	for _, fsys := range fsyss {
		err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
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

			err = os.MkdirAll(filepath.Dir(filepath.Join(outputPath, path)), 0o755)
			if err != nil {
				return err
			}

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

func CreateMemoryFs(ctx context.Context, allPages, pagesToRender []*render.Page) (fstest.MapFS, error) {
	files := fstest.MapFS{}

	for _, page := range pagesToRender {
		if page.Type == render.PageSection {
			subFiles, err := CreateMemoryFs(ctx, allPages, page.Children)
			if err != nil {
				return nil, err
			}
			maps.Copy(files, subFiles)
		}
		if page.Type == render.PageMarkdown {
			fmt.Printf("Writing file: %v\n", page.Href)
			pc := &render.PageContext{
				Title:  page.Title,
				Active: page.Slug,
			}

			var htmlBuffer bytes.Buffer
			err := components.HTML(pc, allPages, page.RenderedContent).Render(ctx, &htmlBuffer)
			if err != nil {
				return nil, err
			}

			files[page.Href] = &fstest.MapFile{Data: htmlBuffer.Bytes()}
		}

	}

	search, err := searchJS(allPages)
	if err != nil {
		return nil, err
	}

	files["search_content.js"] = &fstest.MapFile{Data: search}

	return files, nil
}

func searchJS(pages []*render.Page) ([]byte, error) {
	data := searchIndex(pages)

	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return []byte("var AllPagesData = " + string(b) + ";"), nil
}

func searchIndex(pages []*render.Page) []map[string]string {
	var data []map[string]string
	for _, p := range pages {
		if p.Type == render.PageSection {
			data = append(data, searchIndex(p.Children)...)
		}
		if p.Type == render.PageMarkdown {
			data = append(data, map[string]string{
				"title": p.Title,
				"href":  p.Href,
				"body":  p.RawContent,
			})
		}
	}

	return data
}

func ResetOutputFolder(outputPath string) error {
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
