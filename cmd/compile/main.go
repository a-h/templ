package compile

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/a-h/templ/generator"
	"github.com/hashicorp/go-multierror"
)

func Run(args []string) (err error) {
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
	var errorCount int
	var result error
	for i := 0; i < len(templates); i++ {
		templateStart := time.Now()
		sourceFileName := templates[i]
		t, err := templ.Parse(sourceFileName)
		if err != nil {
			errorCount++
			result = multierror.Append(err, fmt.Errorf("%s parsing error: %w", sourceFileName, err))
			continue
		}
		targetFileName := strings.TrimSuffix(sourceFileName, ".templ") + "_templ.go"
		w, err := os.Create(targetFileName)
		if err != nil {
			errorCount++
			result = multierror.Append(err, fmt.Errorf("%s compilation error: %w", sourceFileName, err))
			continue
		}
		_, err = generator.Generate(t, w)
		if err != nil {
			errorCount++
			result = multierror.Append(err, fmt.Errorf("%s generation error: %w", sourceFileName, err))
			continue
		}
		fmt.Printf("%s compiled in %v\n", sourceFileName, time.Now().Sub(templateStart))
	}
	fmt.Printf("Compiled %d templates with %d errors in %s\n", len(templates), errorCount, time.Now().Sub(start))
	return result
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
