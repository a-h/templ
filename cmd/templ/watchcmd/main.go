package watchcmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/a-h/templ/cmd/templ/processor"
	"github.com/a-h/templ/cmd/templ/visualize"
	"github.com/a-h/templ/generator"
	"github.com/a-h/templ/parser/v2"
	"github.com/fsnotify/fsnotify"
)

type Arguments struct {
	Path                            string
	WorkerCount                     int
	GenerateSourceMapVisualisations bool
}

var defaultWorkerCount = runtime.NumCPU()

var (
	cache = make(map[string]struct{})
)

func Run(args Arguments) (err error) {
	if args.WorkerCount == 0 {
		args.WorkerCount = defaultWorkerCount
	}

	log.Printf("Starting watcher...")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("NewWatcher failed: ", err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		defer close(done)
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// log.Printf("%s %s\n", event.Name, event.Op)
				if event.Op == fsnotify.Write {
					if strings.HasSuffix(event.Name, ".templ") && !strings.HasSuffix(event.Name, "_templ.go") {
						log.Printf("processing: %v", event.Name)
						processSingleFile(event.Name, args.GenerateSourceMapVisualisations)
					}
				}
				if event.Op == fsnotify.Create {
					if fi, err := os.Stat(event.Name); err == nil && fi.IsDir() {
						err = watcher.Add(event.Name)
					}
				}
				if event.Op == fsnotify.Remove {
					if fi, err := os.Stat(event.Name); err == nil && fi.IsDir() {
						err = watcher.Remove(event.Name)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	filepath.Walk(args.Path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			err = watcher.Add(path)
		}
		return nil
	})

	if err != nil {
		log.Fatal("Add failed:", err)
	}
	<-done
	return
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
			err = errors.Join(err, fmt.Errorf("%s: %w", r.FileName, r.Error))
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

	targetFileName := strings.TrimSuffix(templFileName, ".templ") + "_templ_sourcemap.html"
	w, err := os.Create(targetFileName)
	if err != nil {
		return fmt.Errorf("%s sourcemap visualisation error: %w", templFileName, err)
	}
	defer w.Close()
	b := bufio.NewWriter(w)
	defer b.Flush()

	return visualize.HTML(templFileName, string(templContents), string(goContents), sourceMap).Render(context.Background(), b)
}
