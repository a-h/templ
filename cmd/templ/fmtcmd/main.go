package fmtcmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/a-h/templ/cmd/templ/processor"
	parser "github.com/a-h/templ/parser/v2"
	"github.com/natefinch/atomic"
)

const workerCount = 4

func Run(args []string) (err error) {
	if len(args) > 0 {
		return formatDir(args[0])
	}
	return formatStdin()
}

func formatStdin() (err error) {
	var bytes []byte
	bytes, err = io.ReadAll(os.Stdin)
	if err != nil {
		return
	}
	t, err := parser.ParseString(string(bytes))
	if err != nil {
		return fmt.Errorf("parsing error: %w", err)
	}
	err = t.Write(os.Stdout)
	if err != nil {
		return fmt.Errorf("formatting error: %w", err)
	}
	return nil
}

func formatDir(dir string) (err error) {
	start := time.Now()
	results := make(chan processor.Result)
	go processor.Process(dir, format, workerCount, results)
	var successCount, errorCount int
	for r := range results {
		if r.Error != nil {
			err = errors.Join(err, fmt.Errorf("%s: %w", r.FileName, r.Error))
			errorCount++
			continue
		}
		fmt.Printf("%s complete in %v\n", r.FileName, r.Duration)
		successCount++
	}
	fmt.Printf("Formatted %d templates with %d errors in %s\n", successCount+errorCount, errorCount, time.Since(start))
	return
}

func format(fileName string) (err error) {
	contents, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("failed to read file %q: %w", fileName, err)
	}
	t, err := parser.ParseString(string(contents))
	if err != nil {
		return fmt.Errorf("%s parsing error: %w", fileName, err)
	}
	w := new(bytes.Buffer)
	err = t.Write(w)
	if err != nil {
		return fmt.Errorf("%s formatting error: %w", fileName, err)
	}
	if string(contents) == w.String() {
		return nil
	}
	err = atomic.WriteFile(fileName, w)
	if err != nil {
		return fmt.Errorf("%s file write error: %w", fileName, err)
	}
	return
}
