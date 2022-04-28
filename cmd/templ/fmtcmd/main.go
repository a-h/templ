package fmtcmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/a-h/templ/cmd/templ/processor"
	"github.com/a-h/templ/parser"
	"github.com/hashicorp/go-multierror"
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
	bytes, _ := ioutil.ReadAll(os.Stdin)
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
	go processor.Process(".", format, workerCount, results)
	var successCount, errorCount int
	for r := range results {
		if r.Error != nil {
			err = multierror.Append(err, fmt.Errorf("%s: %w", r.FileName, r.Error))
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
	t, err := parser.Parse(fileName)
	if err != nil {
		return fmt.Errorf("%s parsing error: %w", fileName, err)
	}
	w := new(bytes.Buffer)
	err = t.Write(w)
	if err != nil {
		return fmt.Errorf("%s formatting error: %w", fileName, err)
	}
	err = atomic.WriteFile(fileName, w)
	if err != nil {
		return fmt.Errorf("%s file write error: %w", fileName, err)
	}
	return
}
