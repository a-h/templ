package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

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
		var newPage *render.Page
		if info.IsDir() {
			newPage, err = render.NewSectionPage(path, inputFsys)
			if err != nil {
				return err
			}
		}
		if filepath.Ext(info.Name()) == ".md" {
			newPage, err = render.NewMarkdownPage(path, inputFsys)
			if err != nil {
				return err
			}
		}
		pages = append(pages, newPage)
		return nil
	})

	if err != nil {
		panic(err)
	}



	for _, v := range pages {
		writeToDisk(v)
	}

}

func resetOutputPath() error {
	if err := os.RemoveAll(outputPath); err != nil {
		return err
	}
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return err
	}
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
