package processor

import (
	"io/fs"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Result struct {
	FileName string
	Duration time.Duration
	Error    error
}

func Process(dir string, f func(fileName string) error, workerCount int, results chan<- Result) {
	templates := make(chan string)
	go func() {
		defer close(templates)
		if err := FindTemplates(dir, templates); err != nil {
			results <- Result{Error: err}
		}
	}()
	ProcessChannel(templates, dir, f, workerCount, results)
}

func shouldSkipDir(dir string) bool {
	if dir == "." {
		return false
	}
	if dir == "vendor" || dir == "node_modules" {
		return true
	}
	_, name := path.Split(dir)
	// These directories are ignored by the Go tool.
	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
		return true
	}
	return false
}

func FindTemplates(srcPath string, output chan<- string) (err error) {
	return filepath.Walk(srcPath, func(currentPath string, info fs.FileInfo, err error) error {
		if info.IsDir() && shouldSkipDir(currentPath) {
			return filepath.SkipDir
		}
		if !info.IsDir() && strings.HasSuffix(currentPath, ".templ") {
			output <- currentPath
		}
		return nil
	})
}

func ProcessChannel(templates <-chan string, dir string, f func(fileName string) error, workerCount int, results chan<- Result) {
	defer close(results)
	var wg sync.WaitGroup
	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			for sourceFileName := range templates {
				start := time.Now()
				results <- Result{
					FileName: sourceFileName,
					Error:    f(sourceFileName),
					Duration: time.Since(start),
				}
			}
		}()
	}
	wg.Wait()
}
