package fmtcmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/a-h/templ/cmd/templ/processor"
	parser "github.com/a-h/templ/parser/v2"
	"github.com/natefinch/atomic"
)

const workerCount = 4

var mu sync.Mutex

type Arguments struct {
	ToStdout bool
	Files    []string
}

func Run(w io.Writer, args Arguments) (err error) {
	if len(args.Files) == 0 {
		return formatReader(w, os.Stdin)
	}

	return formatDir(args.ToStdout, w, args.Files[0])
}

func formatReader(w io.Writer, r io.Reader) (err error) {
	var bytes []byte
	bytes, err = io.ReadAll(r)
	if err != nil {
		return
	}
	t, err := parser.ParseString(string(bytes))
	if err != nil {
		return fmt.Errorf("parsing error: %w", err)
	}
	err = t.Write(w)
	if err != nil {
		return fmt.Errorf("formatting error: %w", err)
	}
	return nil
}

func formatDir(toStdout bool, w io.Writer, dir string) (err error) {
	start := time.Now()
	results := make(chan processor.Result)
	go processor.Process(dir, newformater(toStdout), workerCount, results)
	var successCount, errorCount int
	for r := range results {
		if r.Error != nil {
			err = errors.Join(err, fmt.Errorf("%s: %w", r.FileName, r.Error))
			errorCount++
			continue
		}
		if !toStdout {
			fmt.Fprintf(os.Stderr, "%s complete in %v\n", r.FileName, r.Duration)
		}
		successCount++
	}
	if !toStdout {
		fmt.Fprintf(os.Stderr, "Formatted %d templates with %d errors in %s\n", successCount+errorCount, errorCount, time.Since(start))
	}
	return
}

func newformater(toStdout bool) func(string) error {
	return func(fileName string) (err error) {
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

		if toStdout {
			mu.Lock()
			fmt.Print(w.String())
			mu.Unlock()
		} else {
			err = atomic.WriteFile(fileName, w)
			if err != nil {
				return fmt.Errorf("%s file write error: %w", fileName, err)
			}
		}
		return
	}
}
