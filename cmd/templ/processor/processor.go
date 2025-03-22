package processor

import (
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/a-h/templ/internal/skipdir"
)

type Result struct {
	FileName    string
	Duration    time.Duration
	Error       error
	ChangesMade bool
}

func Process(dir string, f func(fileName string) (error, bool), workerCount int, results chan<- Result) {
	templates := make(chan string)
	go func() {
		defer close(templates)
		if err := FindTemplates(dir, templates); err != nil {
			results <- Result{Error: err}
		}
	}()
	ProcessChannel(templates, dir, f, workerCount, results)
}

func FindTemplates(srcPath string, output chan<- string) (err error) {
	return filepath.WalkDir(srcPath, func(currentPath string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && skipdir.ShouldSkip(currentPath) {
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasSuffix(currentPath, ".templ") {
			output <- currentPath
		}
		return nil
	})
}

func ProcessChannel(templates <-chan string, dir string, f func(fileName string) (error, bool), workerCount int, results chan<- Result) {
	defer close(results)
	var wg sync.WaitGroup
	wg.Add(workerCount)
	for range workerCount {
		go func() {
			defer wg.Done()
			for sourceFileName := range templates {
				start := time.Now()
				outErr, outChanged := f(sourceFileName)
				results <- Result{
					FileName:    sourceFileName,
					Error:       outErr,
					Duration:    time.Since(start),
					ChangesMade: outChanged,
				}
			}
		}()
	}
	wg.Wait()
}
