package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/a-h/templ/docgen-2/render"
)

const (
	outputPath     = "public"
	contentPath    = "./docs"
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
	sections := []render.SectionPage{}
	mdPages := []render.MarkdownPage{}

	fsys, err := fs.Sub(os.DirFS(contentPath), "docs")
	if err != nil {
		return err
	}

	err = filepath.Walk(
		contentPath,
		func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				sections = append(sections, render.SectionPage{Path: path, Fsys: fsys})
			}
			if filepath.Ext(info.Name()) == ".md" {
				mdPages = append(mdPages, render.MarkdownPage{Path: path, Fsys: fsys})
			}
			return nil
		})
	if err != nil {
		return err
	}

	for _, v := range sections {
		writeToDisk(v)
	}
	for _, v := range mdPages {
		writeToDisk(v)
	}

	return nil
}

func resetOutputPath() error {
	if err :=			os.RemoveAll(filepath.Join(outputPath, entry.Name())); err != nil {
		return err
	}

	if err = os.MkdirAll(outputPath, 0o755); err != nil {
		return err
	}
}

func writeToDisk(r render.Renderer) {


		_ = os.MkdirAll(filepath.Dir(filepath.Join(outputPath, path)), 0o755)

		out, err := os.Create(filepath.Join(outputPath, path))
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, f)

		return err

	return nil
}
