package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/a-h/templ/docs/src"
	"github.com/a-h/templ/docs/src/render"
)

const (
	outputPath = "./build"
	docsPath   = "./docs"
	staticPath = "./static"
	defaultUrl = "https://templ.guide/"
)

var (
	docsFsys   = os.DirFS(docsPath)
	staticFsys = os.DirFS(staticPath)
)

func main() {
	cmd := flag.NewFlagSet("generate", flag.ExitOnError)
	localFlag := cmd.Bool("local", false, "Hosts public/ directory on http://localhost:8080")
	helpFlag := cmd.Bool("help", false, "Print help and exit.")
	if cmd.Parse(os.Args[1:]) != nil || *helpFlag {
		cmd.PrintDefaults()
		return
	}
	if *localFlag {
		render.BaseUrl = "http://localhost:8080/"
	} else {
		render.BaseUrl = defaultUrl
	}

	err := src.ResetOutputFolder(outputPath)
	if err != nil {
		log.Fatal(err)
	}

	pages, err := buildPages()
	if err != nil {
		log.Fatal(err)
	}

	docsFs, err := src.CreateMemoryFs(context.Background(), pages, pages)
	if err != nil {
		log.Fatal(err)
	}

	err = src.WriteToDisk([]fs.FS{docsFs, staticFsys}, outputPath)
	if err != nil {
		log.Fatal(err)
	}

	if *localFlag {
		fs := http.FileServer(http.Dir(outputPath))
		http.Handle("/", fs)
		fmt.Printf("Starting server at %s\n", render.BaseUrl)
		err := http.ListenAndServe("127.0.0.1:8080", nil)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func buildPages() ([]*render.Page, error) {
	var pages []*render.Page

	files, err := fs.ReadDir(docsFsys, ".")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			return nil, err
		}

		path := filepath.Join(".", info.Name())

		newPage, err := render.NewPage(path, info, docsFsys)
		if err != nil {
			return nil, err
		}

		if newPage == nil {
			continue
		}
		pages = append(pages, newPage)

	}

	slices.SortFunc(pages, func(a, b *render.Page) int {
		if a.Order == b.Order {
			return strings.Compare(a.Title, b.Title)
		}

		return a.Order - b.Order
	})

	return pages, nil
}
