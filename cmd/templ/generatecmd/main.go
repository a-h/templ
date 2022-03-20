package generatecmd

import (
	"bufio"
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
	err := compile(fileName)
	fmt.Printf("Generated code for %q in %s\n", fileName, time.Since(start))
	return err
}

func processPath(path string) (err error) {
	start := time.Now()
	results := make(chan processor.Result)
	go processor.Process(path, compile, workerCount, results)
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
	return err
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
	b := bufio.NewWriter(w)
	_, err = generator.Generate(t, b)
	if err != nil {
		return fmt.Errorf("%s generation error: %w", fileName, err)
	}
	if b.Flush() != nil {
		return fmt.Errorf("%s write file error: %w", targetFileName, err)
	}
	return
}
