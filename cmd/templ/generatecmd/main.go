package generatecmd

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"go/format"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/a-h/templ/cmd/templ/generatecmd/run"
	"github.com/a-h/templ/cmd/templ/generatecmd/sse"
	"github.com/a-h/templ/cmd/templ/visualize"
	"github.com/a-h/templ/generator"
	"github.com/a-h/templ/parser/v2"
	"github.com/cenkalti/backoff/v4"
	"github.com/cli/browser"
)

type Arguments struct {
	FileName                        string
	Path                            string
	Watch                           bool
	Command                         string
	ProxyPort                       int
	Proxy                           string
	WorkerCount                     int
	GenerateSourceMapVisualisations bool
}

var defaultWorkerCount = runtime.NumCPU()

//go:embed script.js
var script string

func Run(args Arguments) (err error) {
	start := time.Now()
	if args.Watch && args.FileName != "" {
		return fmt.Errorf("cannot watch a single file, remove the -f or -watch flag")
	}
	if args.FileName != "" {
		return processSingleFile(args.FileName, args.GenerateSourceMapVisualisations)
	}
	var proxy *url.URL
	if args.Proxy != "" {
		proxy, err = url.Parse(args.Proxy)
		if err != nil {
			return fmt.Errorf("failed to parse proxy URL: %w", err)
		}
	}
	if args.ProxyPort == 0 {
		args.ProxyPort = 7331
	}

	if args.WorkerCount == 0 {
		args.WorkerCount = defaultWorkerCount
	}
	if !path.IsAbs(args.Path) {
		args.Path, err = filepath.Abs(args.Path)
		if err != nil {
			return
		}
	}

	var sses *sse.Server
	if args.Proxy != "" {
		sses = sse.New()
	}

	fmt.Println("Processing path:", args.Path)
	var firstRunComplete bool
	fileNameToLastModTime := make(map[string]time.Time)
	for !firstRunComplete || args.Watch {
		changesFound, errs := processChanges(fileNameToLastModTime, args.Path, args.GenerateSourceMapVisualisations, args.WorkerCount)
		if len(errs) > 0 {
			fmt.Printf("Error processing path: %v\n", errors.Join(errs...))
		}
		if changesFound > 0 {
			fmt.Printf("Generated code for %d templates with %d errors in %s\n", changesFound, len(errs), time.Since(start))
			if args.Command != "" {
				fmt.Printf("Executing command: %s\n", args.Command)
				if _, err := run.Run(context.Background(), args.Path, args.Command); err != nil {
					fmt.Printf("Error starting command: %v\n", err)
				}
				// Send server-sent event.
				if sses != nil {
					sses.Send("message", "reload")
				}
			}
			if !firstRunComplete && proxy != nil {
				proxyURL := fmt.Sprintf("http://127.0.0.1:%d", args.ProxyPort)
				scriptTag := `<script src="/_templ/reload/script.js"></script>`
				go func() {
					fmt.Printf("Proxying from %s to target: %s\n", proxyURL, proxy.String())
					proxy := httputil.NewSingleHostReverseProxy(proxy)
					proxy.ModifyResponse = func(r *http.Response) error {
						if contentType := r.Header.Get("Content-Type"); contentType != "text/html" {
							return nil
						}
						body, err := io.ReadAll(r.Body)
						if err != nil {
							return err
						}
						updated := strings.Replace(string(body), "</body>", scriptTag+"</body>", -1)
						r.Body = io.NopCloser(strings.NewReader(updated))
						r.ContentLength = int64(len(updated))
						r.Header.Set("Content-Length", strconv.Itoa(len(updated)))
						return nil
					}
					h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if r.URL.Path == "/_templ/reload/script.js" {
							// Provides a script that reloads the page.
							w.Header().Add("Content-Type", "text/javascript")
							_, err := io.WriteString(w, script)
							if err != nil {
								fmt.Printf("failed to write script: %v\n", err)
							}
							return
						}
						if r.URL.Path == "/_templ/reload/events" {
							// Provides a list of messages including a reload message.
							sses.ServeHTTP(w, r)
							return
						}
						proxy.ServeHTTP(w, r)
					})
					if err := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", args.ProxyPort), h); err != nil {
						fmt.Printf("Error starting proxy: %v\n", err)
					}
				}()
				fmt.Printf("Opening URL: %s\n", proxyURL)
				go func() {
					if err := openURL(proxyURL); err != nil {
						fmt.Printf("Error opening URL: %v\n", err)
					}
				}()
			}
		}
		if firstRunComplete {
			time.Sleep(250 * time.Millisecond)
		}
		firstRunComplete = true
		start = time.Now()
	}
	return err
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

func processChanges(fileNameToLastModTime map[string]time.Time, path string, generateSourceMapVisualisations bool, maxWorkerCount int) (changesFound int, errs []error) {
	sem := make(chan struct{}, maxWorkerCount)
	var wg sync.WaitGroup

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && shouldSkipDir(path) {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".templ") {
			lastModTime := fileNameToLastModTime[path]
			if info.ModTime().After(lastModTime) {
				fileNameToLastModTime[path] = info.ModTime()
				changesFound++

				// Start a processor, but limit to maxWorkerCount.
				sem <- struct{}{}
				wg.Add(1)
				go func() {
					defer wg.Done()
					if err := processSingleFile(path, generateSourceMapVisualisations); err != nil {
						errs = append(errs, err)
					}
					<-sem
				}()
			}
		}
		return nil
	})
	if err != nil {
		errs = append(errs, err)
	}

	wg.Wait()

	return changesFound, errs
}

func openURL(url string) error {
	backoff := backoff.NewExponentialBackOff()
	var client http.Client
	client.Timeout = 1 * time.Second
	for {
		if _, err := client.Get(url); err == nil {
			break
		}
		d := backoff.NextBackOff()
		log.Printf("Server not ready. Retrying in %v...", d)
		time.Sleep(d)
	}
	return browser.OpenURL(url)
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

func compile(fileName string, generateSourceMapVisualisations bool) (err error) {
	t, err := parser.Parse(fileName)
	if err != nil {
		return fmt.Errorf("%s parsing error: %w", fileName, err)
	}
	targetFileName := strings.TrimSuffix(fileName, ".templ") + "_templ.go"

	var b bytes.Buffer
	sourceMap, err := generator.Generate(t, &b)
	if err != nil {
		return fmt.Errorf("%s generation error: %w", fileName, err)
	}

	data, err := format.Source(b.Bytes())
	if err != nil {
		return fmt.Errorf("%s source formatting error: %w", fileName, err)
	}

	w, err := os.Create(targetFileName)
	if err != nil {
		return fmt.Errorf("%s compilation error: %w", fileName, err)
	}
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("%s compilation error: %w", fileName, err)
	}

	defer w.Close()
	if w.Sync() != nil {
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
