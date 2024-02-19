package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/a-h/templ/docgen-2/src/render"
)

const (
	outputPath = "../public"
	inputPath  = "../../docs"
)

var inputFsys = os.DirFS(inputPath)
var staticFsys = os.DirFS("../static")

func main() {
	cmd := flag.NewFlagSet("generate", flag.ExitOnError)
	cmd.StringVar(&render.BaseUrl, "url", "https://cugu.github.io/templ/new/", "The base URL for the site.")
	helpFlag := cmd.Bool("help", false, "Print help and exit.")
	if cmd.Parse(os.Args[1:]) != nil || *helpFlag {
		cmd.PrintDefaults()
		return
	}

	err := generate(context.Background())
	if err != nil {
		panic(err)
	}

}
func generate(ctx context.Context) error {

	var pages []*render.Page

	files, err := fs.ReadDir(inputFsys, ".")
	if err != nil {
		return err
	}
	fmt.Printf("Reading from %v\n", files)

	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			return err
		}

		path := filepath.Join(".", file.Name())

		newPage, err := render.NewPage(path, info, inputFsys)
		if err != nil {
			return err
		}

		if newPage == nil {
			continue
		}
		pages = append(pages, newPage)

	}

	fmt.Printf("Created %v page structs\n", len(pages))
	docsFs, err := createMemoryFs(ctx, pages, pages)
	if err != nil {
		return err
	}

	err = writeToDisk([]fs.FS{docsFs, staticFsys})
	if err != nil {
		return err
	}

	return nil
}
