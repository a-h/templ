package migratecmd

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/a-h/templ/cmd/templ/processor"
	v1 "github.com/a-h/templ/parser/v1"
	v2 "github.com/a-h/templ/parser/v2"
	"github.com/hashicorp/go-multierror"
	"github.com/jinzhu/copier"
	"github.com/natefinch/atomic"
)

const workerCount = 4

type Arguments struct {
	FileName string
	Path     string
}

func Run(args Arguments) (err error) {
	if args.FileName != "" {
		return processSingleFile(args.FileName)
	}
	return processPath(args.Path)
}

func processSingleFile(fileName string) error {
	start := time.Now()
	err := migrate(fileName)
	fmt.Printf("Migrated code for %q in %s\n", fileName, time.Since(start))
	return err
}

func processPath(path string) (err error) {
	start := time.Now()
	results := make(chan processor.Result)
	go processor.Process(path, migrate, workerCount, results)
	var successCount, errorCount int
	for r := range results {
		if r.Error != nil {
			err = multierror.Append(err, fmt.Errorf("%s: %w", r.FileName, r.Error))
			errorCount++
			continue
		}
		successCount++
		fmt.Printf("%s complete in %v\n", r.FileName, r.Duration)
	}
	fmt.Printf("Migrated code for %d templates with %d errors in %s\n", successCount+errorCount, errorCount, time.Since(start))
	return err
}

func migrate(fileName string) (err error) {
	// Check that it's actually a V1 file.
	_, err = v2.Parse(fileName)
	if err == nil {
		return fmt.Errorf("migrate: %s able to parse file as V2, are you sure this needs to be migrated?", fileName)
	}
	if err != v2.ErrLegacyFileFormat {
		return fmt.Errorf("migrate: %s unexpected error: %v", fileName, err)
	}
	// Parse.
	v1Template, err := v1.Parse(fileName)
	if err != nil {
		return fmt.Errorf("migrate: %s v1 parsing error: %w", fileName, err)
	}
	// Convert.
	var v2Template v2.TemplateFile

	// Copy everything.
	err = copier.Copy(&v2Template, &v1Template)
	if err != nil {
		return fmt.Errorf("migrate: %s v1 to v2 copy error: %w", fileName, err)
	}

	// Rework the structure.

	// Update
	var sb strings.Builder
	sb.WriteString("package " + v1Template.Package.Expression.Value)
	sb.WriteString("\n")
	for _, imp := range v1Template.Imports {
		sb.WriteString("import ")
		sb.WriteString(imp.Expression.Value)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
	v2Template.Package.Expression.Value = sb.String()

	// Write the updated file.
	w := new(bytes.Buffer)
	err = v2Template.Write(w)
	if err != nil {
		return fmt.Errorf("%s formatting error: %w", fileName, err)
	}
	err = atomic.WriteFile(fileName, w)
	if err != nil {
		return fmt.Errorf("%s file write error: %w", fileName, err)
	}
	return
}
