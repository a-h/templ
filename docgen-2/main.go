package main

import (
	"bytes"
	"context"
	"flag"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"testing/fstest"

	"github.com/a-h/templ/docgen-2/components"
	"github.com/a-h/templ/docgen-2/render"
)

const (
	outputPath = "public"
	inputPath  = "./docs"
)

var inputFsys = os.DirFS(inputPath)

func main() {
	cmd := flag.NewFlagSet("generate", flag.ExitOnError)
	cmd.StringVar(&render.BaseUrl, "url", "https://cugu.github.io/templ/new/", "The base URL for the site.")
	helpFlag := cmd.Bool("help", false, "Print help and exit.")
	if cmd.Parse(os.Args[1:]) != nil || *helpFlag {
		cmd.PrintDefaults()
		return
	}

	var pages []*render.Page

	err := filepath.Walk(inputPath, func(path string, info fs.FileInfo, err error) error {
		page, err := render.NewPage(path, info, inputFsys)
		if err != nil {
			return err
		}
		pages = append(pages, page)
		return nil
	})

	if err != nil {
		panic(err)
	}

}

func createMemoryFs(ctx context.Context, pages, children []*render.Page) (fstest.MapFS, error) {
	files := fstest.MapFS{}

	for _, p := range children {
		switch p.Type {
		case render.PageSection:
			subFiles, err := createMemoryFs(ctx, pages, p.Children)
			if err != nil {
				return nil, err
			}

			maps.Copy(files, subFiles)
		case render.PageMarkdown:
			mainContent := p.Html

			pc := &render.PageContext{
				Title:  p.Title,
				Active: p.Slug,
			}

			var htmlBuffer bytes.Buffer
			if err := components.HTML(pc, pages, mainContent).Render(ctx, &htmlBuffer); err != nil {
				return nil, err
			}

			files[p.Href] = &fstest.MapFile{Data: htmlBuffer.Bytes()}
		}
	}

	return files, nil

}

func resetOutputPath() error {
	if err := os.RemoveAll(outputPath); err != nil {
		return err
	}
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return err
	}
	return nil
}

func writeToDisk(r []*render.Page) error {

	err := os.MkdirAll(filepath.Dir(filepath.Join(outputPath, r.File)), 0o755)
	if err != nil {
		return err
	}

	out, err := os.Create(filepath.Join(outputPath, r.File))
	defer out.Close()
	if err != nil {
		return err
	}

	return nil
}
