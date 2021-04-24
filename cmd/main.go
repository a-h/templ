package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/a-h/templ/generator"
)

var Version = ""

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "compile":
		compileCmd(os.Args[2:])
		return
	case "lsp":
		lspCmd(os.Args[2:])
		return
	case "version":
		fmt.Println(Version)
		return
	case "--version":
		fmt.Println(Version)
		return
	}
	usage()
}

func usage() {
	fmt.Println(`usage: templ <command> [parameters]
To see help text, you can run:
  templ compile --help
  templ lsp --help
  templ version
examples:
  templ compile`)
	os.Exit(1)
}

func compileCmd(args []string) {
	cmd := flag.NewFlagSet("compile", flag.ExitOnError)
	helpFlag := cmd.Bool("help", false, "Print help and exit.")
	err := cmd.Parse(args)
	if err != nil || *helpFlag {
		cmd.PrintDefaults()
		return
	}
	// Search for *.templ files and compile them.
	templates, err := getTemplates(".")
	if err != nil {
		fmt.Printf("error getting templates: %v\n", err)
		os.Exit(1)
	}
	if len(templates) == 0 {
		fmt.Println("No templates found.")
		return
	}
	//TODO: Use goroutines to parallelise.
	start := time.Now()
	for i := 0; i < len(templates); i++ {
		templateStart := time.Now()
		sourceFileName := templates[i]
		fmt.Printf("Compiling template %s", sourceFileName)
		t, err := templ.Parse(sourceFileName)
		if err != nil {
			fmt.Printf("  error compiling: %v\n", err)
			continue
		}
		targetFileName := strings.TrimSuffix(sourceFileName, ".templ") + "_templ.go"
		w, err := os.Create(targetFileName)
		if err != nil {
			fmt.Printf("  error compiling: %v\n", err)
			continue
		}
		//TODO: Use the sourcemap to run the language server.
		_, err = generator.Generate(t, w)
		if err != nil {
			fmt.Printf("  error compiling: %v\n", err)
			continue
		}
		fmt.Printf("  compiled in %s\n", time.Now().Sub(templateStart))
	}
	fmt.Printf("Done. Compiled %d templates in %s\n", len(templates), time.Now().Sub(start))
}

func getTemplates(srcPath string) (fileNames []string, err error) {
	filepath.Walk(srcPath, func(currentPath string, info fs.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(currentPath, ".templ") {
			fileNames = append(fileNames, currentPath)
		}
		return nil
	})
	if err != nil {
		err = fmt.Errorf("failed to walk directory: %w", err)
		return
	}
	return
}

func lspCmd(args []string) {
	cmd := flag.NewFlagSet("init", flag.ExitOnError)
	helpFlag := cmd.Bool("help", false, "Print help and exit.")
	err := cmd.Parse(args)
	if err != nil || *helpFlag {
		cmd.PrintDefaults()
		return
	}
	//TODO: Run the language server.
}
