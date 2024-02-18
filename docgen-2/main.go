package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/a-h/templ/docgen-2/render"
)

//go:embed static
var docGenStaticEmbed embed.FS

const (
	outputPath = "public"
	inputPath  = "../docs/docs"
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
		pages = append(pages, newPage)
	}

	fmt.Printf("Created %v page structs\n", len(pages))

	// docsFs, err := createMemoryFs(context.Background(), pages, pages)
	// if err != nil {
	// 	return err
	// }

	static, err := fs.Sub(inputFsys, "static")
	if err != nil {
		return err
	}

	docGenStatic, err := fs.Sub(docGenStaticEmbed, "static")
	if err != nil {
		return err
	}

	err = writeToDisk([]fs.FS{static, docGenStatic})
	if err != nil {
		return err
	}

	return nil
}
