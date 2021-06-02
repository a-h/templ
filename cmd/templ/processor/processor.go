package processor

import (
	"io/fs"
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
	defer close(results)
	templates := make(chan string)
	go func() {
		if err := getTemplates(".", templates); err != nil {
			results <- Result{Error: err}
		}
	}()
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
					Duration: time.Now().Sub(start),
				}
			}
		}()
	}
	wg.Wait()
}

func getTemplates(srcPath string, output chan<- string) (err error) {
	defer close(output)
	return filepath.Walk(srcPath, func(currentPath string, info fs.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(currentPath, ".templ") {
			output <- currentPath
		}
		return nil
	})
}
