package generatecmd

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"go/format"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	_ "net/http/pprof"

	"github.com/a-h/templ/cmd/templ/generatecmd/proxy"
	"github.com/a-h/templ/cmd/templ/generatecmd/run"
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
	// PPROFPort is the port to run the pprof server on.
	PPROFPort int
}

var defaultWorkerCount = runtime.NumCPU()

func Run(args Arguments) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()
	if args.PPROFPort > 0 {
		go func() {
			_ = http.ListenAndServe(fmt.Sprintf("localhost:%d", args.PPROFPort), nil)
		}()
	}
	go func() {
		select {
		case <-signalChan: // First signal, cancel context.
			fmt.Println("\nCancelling...")
			cancel()
		case <-ctx.Done():
		}
		<-signalChan // Second signal, hard exit.
		os.Exit(2)
	}()
	err = runCmd(ctx, args)
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}

func runCmd(ctx context.Context, args Arguments) (err error) {
	start := time.Now()
	if args.Watch && args.FileName != "" {
		return fmt.Errorf("cannot watch a single file, remove the -f or -watch flag")
	}
	if args.FileName != "" {
		return processSingleFile(ctx, args.FileName, args.GenerateSourceMapVisualisations)
	}
	var target *url.URL
	if args.Proxy != "" {
		target, err = url.Parse(args.Proxy)
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

	var p *proxy.Handler
	if args.Proxy != "" {
		p = proxy.New(args.ProxyPort, target)
	}

	fmt.Println("Processing path:", args.Path)
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = time.Millisecond * 500
	bo.MaxInterval = time.Second * 3
	var firstRunComplete bool
	fileNameToLastModTime := make(map[string]time.Time)
	for !firstRunComplete || args.Watch {
		changesFound, errs := processChanges(ctx, fileNameToLastModTime, args.Path, args.GenerateSourceMapVisualisations, args.WorkerCount)
		if len(errs) > 0 {
			if errors.Is(errs[0], context.Canceled) {
				return errs[0]
			}
			fmt.Printf("Error processing path: %v\n", errors.Join(errs...))
		}
		if changesFound > 0 {
			fmt.Printf("Generated code for %d templates with %d errors in %s\n", changesFound, len(errs), time.Since(start))
			if args.Command != "" {
				fmt.Printf("Executing command: %s\n", args.Command)
				if _, err := run.Run(ctx, args.Path, args.Command); err != nil {
					fmt.Printf("Error starting command: %v\n", err)
				}
				// Send server-sent event.
				if p != nil {
					p.SendSSE("message", "reload")
				}
			}
			if !firstRunComplete && p != nil {
				go func() {
					fmt.Printf("Proxying from %s to target: %s\n", p.URL, p.Target.String())
					if err := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", args.ProxyPort), p); err != nil {
						fmt.Printf("Error starting proxy: %v\n", err)
					}
				}()
				go func() {
					fmt.Printf("Opening URL: %s\n", p.Target.String())
					if err := openURL(p.URL); err != nil {
						fmt.Printf("Error opening URL: %v\n", err)
					}
				}()
			}
		}
		if firstRunComplete {
			if changesFound > 0 {
				bo.Reset()
			}
			time.Sleep(bo.NextBackOff())
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

func processChanges(ctx context.Context, fileNameToLastModTime map[string]time.Time, path string, generateSourceMapVisualisations bool, maxWorkerCount int) (changesFound int, errs []error) {
	sem := make(chan struct{}, maxWorkerCount)
	var wg sync.WaitGroup

	err := filepath.WalkDir(path, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err = ctx.Err(); err != nil {
			return err
		}
		if info.IsDir() && shouldSkipDir(path) {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".templ") {
			lastModTime := fileNameToLastModTime[path]
			fileInfo, err := info.Info()
			if err != nil {
				return fmt.Errorf("failed to get file info: %w", err)
			}
			if fileInfo.ModTime().After(lastModTime) {
				fileNameToLastModTime[path] = fileInfo.ModTime()
				changesFound++

				// Start a processor, but limit to maxWorkerCount.
				sem <- struct{}{}
				wg.Add(1)
				go func() {
					defer wg.Done()
					if err := processSingleFile(ctx, path, generateSourceMapVisualisations); err != nil {
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
	backoff.InitialInterval = time.Second
	var client http.Client
	client.Timeout = 1 * time.Second
	for {
		if _, err := client.Get(url); err == nil {
			break
		}
		d := backoff.NextBackOff()
		fmt.Printf("Server not ready. Retrying in %v...\n", d)
		time.Sleep(d)
	}
	return browser.OpenURL(url)
}

func processSingleFile(ctx context.Context, fileName string, generateSourceMapVisualisations bool) error {
	start := time.Now()
	err := compile(ctx, fileName, generateSourceMapVisualisations)
	if err != nil {
		return err
	}
	fmt.Printf("Generated code for %q in %s\n", fileName, time.Since(start))
	return err
}

func compile(ctx context.Context, fileName string, generateSourceMapVisualisations bool) (err error) {
	if err = ctx.Err(); err != nil {
		return
	}

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

	if err = os.WriteFile(targetFileName, data, 0644); err != nil {
		return fmt.Errorf("%s write file error: %w", targetFileName, err)
	}

	if generateSourceMapVisualisations {
		err = generateSourceMapVisualisation(ctx, fileName, targetFileName, sourceMap)
	}
	return
}

func generateSourceMapVisualisation(ctx context.Context, templFileName, goFileName string, sourceMap *parser.SourceMap) error {
	if err := ctx.Err(); err != nil {
		return err
	}
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

	return visualize.HTML(templFileName, string(templContents), string(goContents), sourceMap).Render(ctx, b)
}
