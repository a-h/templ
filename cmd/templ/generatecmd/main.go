package generatecmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/a-h/templ/cmd/templ/processor"
	"github.com/a-h/templ/generator"
	"github.com/a-h/templ/parser/v2"
	"github.com/hashicorp/go-multierror"
)

type Arguments struct {
	FileName                        string
	Path                            string
	WorkerCount                     int
	GenerateSourceMapVisualisations bool
}

var defaultWorkerCount = runtime.NumCPU()

func Run(args Arguments) (err error) {
	if args.FileName != "" {
		return processSingleFile(args.FileName, args.GenerateSourceMapVisualisations)
	}
	if args.WorkerCount == 0 {
		args.WorkerCount = defaultWorkerCount
	}
	return processPath(args.Path, args.GenerateSourceMapVisualisations, args.WorkerCount)
}

func processSingleFile(fileName string, generateSourceMapVisualisations bool) error {
	start := time.Now()
	err := compile(fileName, generateSourceMapVisualisations)
	if err != nil {
		return err
	}
	fmt.Printf("Generated code for %q in %s\n", fileName, time.Since(start))
	return err
}

func processPath(path string, generateSourceMapVisualisations bool, workerCount int) (err error) {
	start := time.Now()
	results := make(chan processor.Result)
	p := func(fileName string) error {
		return compile(fileName, generateSourceMapVisualisations)
	}
	go processor.Process(path, p, workerCount, results)
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

func compile(fileName string, generateSourceMapVisualisations bool) (err error) {
	t, err := parser.Parse(fileName)
	if err != nil {
		return fmt.Errorf("%s parsing error: %w", fileName, err)
	}
	targetFileName := strings.TrimSuffix(fileName, ".templ") + "_templ.go"
	w, err := os.Create(targetFileName)
	if err != nil {
		return fmt.Errorf("%s compilation error: %w", fileName, err)
	}
	defer w.Close()
	b := bufio.NewWriter(w)
	defer b.Flush()
	sourceMap, err := generator.Generate(t, b)
	if err != nil {
		return fmt.Errorf("%s generation error: %w", fileName, err)
	}
	if b.Flush() != nil {
		return fmt.Errorf("%s write file error: %w", targetFileName, err)
	}
	if generateSourceMapVisualisations {
		err = generateSourceMapVisualisation(fileName, targetFileName, sourceMap)
	}
	return
}

func generateSourceMapVisualisation(templFileName, goFileName string, sourceMap *parser.SourceMap) error {
	var templContents, goContents []byte
	var templErr, goErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		templContents, templErr = os.ReadFile(templFileName)
	}()
	go func() {
		defer wg.Done()
		goContents, goErr = os.ReadFile(goFileName)
	}()
	wg.Wait()
	if templErr != nil {
		return templErr
	}
	if goErr != nil {
		return templErr
	}
	tl := templLines{contents: string(templContents), sourceMap: sourceMap}
	gl := goLines{contents: string(goContents), sourceMap: sourceMap}
	visualisationComponent := visualisation(templFileName, tl, gl)

	targetFileName := strings.TrimSuffix(templFileName, ".templ") + "_templ_sourcemap.html"
	w, err := os.Create(targetFileName)
	if err != nil {
		return fmt.Errorf("%s sourcemap visualisation error: %w", templFileName, err)
	}
	defer w.Close()
	b := bufio.NewWriter(w)
	defer b.Flush()

	return visualisationComponent.Render(context.Background(), b)
}

type templLines struct {
	contents  string
	sourceMap *parser.SourceMap
}

func (tl templLines) Render(ctx context.Context, w io.Writer) error {
	templLines := strings.Split(tl.contents, "\n")
	for lineIndex, line := range templLines {
		for colIndex, c := range line {
			if _, m, ok := tl.sourceMap.TargetPositionFromSource(uint32(lineIndex), uint32(colIndex)); ok {
				sourceID := fmt.Sprintf("src_%d", m.Source.Range.From.Index)
				targetID := fmt.Sprintf("tgt_%d", m.Target.From.Index)
				if err := mappedCharacter(string(c), sourceID, targetID).Render(ctx, w); err != nil {
					return err
				}
			} else {
				if _, err := w.Write([]byte(string(c))); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

type goLines struct {
	contents  string
	sourceMap *parser.SourceMap
}

func (gl goLines) Render(ctx context.Context, w io.Writer) error {
	templLines := strings.Split(gl.contents, "\n")
	for lineIndex, line := range templLines {
		for colIndex, c := range line {
			if _, m, ok := gl.sourceMap.SourcePositionFromTarget(uint32(lineIndex), uint32(colIndex)); ok {
				sourceID := fmt.Sprintf("src_%d", m.Source.Range.From.Index)
				targetID := fmt.Sprintf("tgt_%d", m.Target.From.Index)
				if err := mappedCharacter(string(c), sourceID, targetID).Render(ctx, w); err != nil {
					return err
				}
			} else {
				if _, err := w.Write([]byte(string(c))); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
