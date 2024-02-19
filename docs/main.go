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

	"github.com/a-h/templ/docs/src"
	"github.com/a-h/templ/docs/src/render"
)

const (
	outputPath = "./public"
	docsPath   = "./docs"
	staticPath = "./static"
	defaultUrl = "https://cugu.github.io/templ/new/"
)

var docsFsys = os.DirFS(docsPath)
var staticFsys = os.DirFS(staticPath)

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

	err := resetOutputFolder()
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

	if render.BaseUrl != defaultUrl {
		err := startLocalHttp()
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
	return pages, nil

}

func resetOutputFolder() error {
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

func startLocalHttp() error {
	fs := http.FileServer(http.Dir(outputPath))
	http.Handle("/", fs)
	fmt.Println("Listening on :8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return err
	}
	return nil
}
