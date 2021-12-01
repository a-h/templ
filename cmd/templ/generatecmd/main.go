package generatecmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/a-h/templ/cmd/templ/processor"
	"github.com/a-h/templ/generator"
	"github.com/a-h/templ/parser"
	"github.com/hashicorp/go-multierror"
)

const workerCount = 4

func Run(args []string) (err error) {
	start := time.Now()
	results := make(chan processor.Result)
	go processor.Process(".", compile, workerCount, results)
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
	fmt.Printf("Generated code for %d templates with %d errors in %s\n", successCount+errorCount, errorCount, time.Since(start))
	return
}

func compile(fileName string) (err error) {
	t, err := parser.Parse(fileName)
	if err != nil {
		return fmt.Errorf("%s parsing error: %w", fileName, err)
	}
	targetFileName := strings.TrimSuffix(fileName, ".templ") + "_templ.go"
	w, err := os.Create(targetFileName)
	if err != nil {
		return fmt.Errorf("%s compilation error: %w", fileName, err)
	}
	_, err = generator.Generate(t, w)
	if err != nil {
		return fmt.Errorf("%s generation error: %w", fileName, err)
	}
	return
}
