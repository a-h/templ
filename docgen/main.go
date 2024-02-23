package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed static
var docGenStaticEmbed embed.FS

const (
	outputPath     = "public"
	contentPath    = "../docs"
	defaultBaseURL = "https://cugu.github.io/templ/new/"
)

func main() {
	cmd := flag.NewFlagSet("generate", flag.ExitOnError)
	baseURL := cmd.String("url", defaultBaseURL, "The base URL for the site.")
	helpFlag := cmd.Bool("help", false, "Print help and exit.")

	if cmd.Parse(os.Args[1:]) != nil || *helpFlag {
		cmd.PrintDefaults()

		return
	}

	if err := generate(context.Background(), *baseURL); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func generate(ctx context.Context, baseURL string) error {
	fsys := os.DirFS(contentPath)

	docs, err := fs.Sub(fsys, "docs")
	if err != nil {
		return err
	}

	pages, err := newPages(docs, ".")
	if err != nil {
		return err
	}

	docFS, err := render(ctx, baseURL, pages)
	if err != nil {
		return err
	}

	static, err := fs.Sub(fsys, "static")
	if err != nil {
		return err
	}

	docGenStatic, err := fs.Sub(docGenStaticEmbed, "static")
	if err != nil {
		return err
	}

	return writeToDisk([]fs.FS{static, docFS, docGenStatic}, outputPath)
}

func writeToDisk(fsyss []fs.FS, outputPath string) error {
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
