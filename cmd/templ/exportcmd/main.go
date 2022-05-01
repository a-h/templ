package exportcmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/a-h/templ"
	"github.com/a-h/templ/cmd/templ/processor"
	"github.com/a-h/templ/parser"
	"github.com/hashicorp/go-multierror"
	"github.com/natefinch/atomic"
)

const workerCount = 4

func Run(args []string) (err error) {
	if len(args) > 0 {
		return exportDir(args[0])
	}
	return exportStdin()
}

type Export struct {
	Version  string
	Template parser.TemplateFile
}

func exportStdin() (err error) {
	var bytes []byte
	bytes, err = ioutil.ReadAll(os.Stdin)
	if err != nil {
		return
	}
	t, err := parser.ParseString(string(bytes))
	if err != nil {
		return fmt.Errorf("parsing error: %w", err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", " ")
	err = enc.Encode(Export{
		Version:  templ.Version,
		Template: t,
	})
	if err != nil {
		return fmt.Errorf("export error: %w", err)
	}
	return nil
}

func exportDir(dir string) (err error) {
	start := time.Now()
	results := make(chan processor.Result)
	go processor.Process(".", export, workerCount, results)
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
	fmt.Printf("Exported %d templates to JSON with %d errors in %s\n", successCount+errorCount, errorCount, time.Since(start))
	return
}

func export(fileName string) (err error) {
	t, err := parser.Parse(fileName)
	if err != nil {
		return fmt.Errorf("%s parsing error: %w", fileName, err)
	}
	w := new(bytes.Buffer)
	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	err = enc.Encode(Export{
		Version:  templ.Version,
		Template: t,
	})
	if err != nil {
		return fmt.Errorf("%s export error: %w", fileName, err)
	}
	err = atomic.WriteFile(fileName+".json", w)
	if err != nil {
		return fmt.Errorf("%s file write error: %w", fileName, err)
	}
	return
}
