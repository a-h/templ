package compile

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/a-h/templ/generator"
)

func Run(args []string) error {
	// Search for *.templ files and compile them.
	templates, err := getTemplates(".")
	if err != nil {
		return fmt.Errorf("error getting templates: %w", err)
	}
	if len(templates) == 0 {
		fmt.Println("No templates found.")
		return nil
	}
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
		sm, err := generator.Generate(t, w)
		if err != nil {
			fmt.Printf("  error compiling: %v\n", err)
			continue
		}
		targetSourceMapFileName := strings.TrimSuffix(sourceFileName, ".templ") + "_sourcemap.json"
		smFile, err := os.Create(targetSourceMapFileName)
		if err != nil {
			fmt.Printf("  error creating sourcemap file: %v\n", err)
			continue
		}
		d := json.NewEncoder(smFile)
		err = d.Encode(sm)
		if err != nil {
			fmt.Printf("   error writing sourcemap: %v\n", err)
			continue
		}
		fmt.Printf("  compiled in %s\n", time.Now().Sub(templateStart))
	}
	fmt.Printf("Done. Compiled %d templates in %s\n", len(templates), time.Now().Sub(start))
	return nil
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
