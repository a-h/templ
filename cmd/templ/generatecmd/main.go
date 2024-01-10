package generatecmd

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"errors"
	"fmt"
	"go/format"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	_ "net/http/pprof"

	"github.com/a-h/templ"
	"github.com/a-h/templ/cmd/templ/generatecmd/proxy"
	"github.com/a-h/templ/cmd/templ/generatecmd/run"
	"github.com/a-h/templ/cmd/templ/visualize"
	"github.com/a-h/templ/generator"
	"github.com/a-h/templ/parser/v2"
	"github.com/cenkalti/backoff/v4"
	"github.com/cli/browser"
	"github.com/fatih/color"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"
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
	IncludeVersion                  bool
	IncludeTimestamp                bool
	// PPROFPort is the port to run the pprof server on.
	PPROFPort         int
	KeepOrphanedFiles bool
}

var defaultWorkerCount = runtime.NumCPU()

func Run(w io.Writer, args Arguments) (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	watchCtx, watchCancel := context.WithCancel(context.Background())

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
		watching := args.Watch
		for {
			select {
			case <-signalChan: // First signal, cancel context.
				if watching {
					fmt.Println("Stopping watch operation...")
					watchCancel()
					continue
				}

				if ctx.Err() != nil {
					fmt.Fprintln(w, "\nHARD EXIT")
					os.Exit(2) // hard exit
					continue
				}

				fmt.Fprintln(w, "\nCancelling...")
				cancel()

			case <-ctx.Done():
				break
			}
		}
	}()

	err = runCmd(ctx, watchCtx, w, args)
	if errors.Is(err, context.Canceled) {
		return nil
	}

	return err
}

func runCmd(ctx, watchCtx context.Context, w io.Writer, args Arguments) error {
	var err error

	if args.Watch && args.FileName != "" {
		return fmt.Errorf("cannot watch a single file, remove the -f or -watch flag")
	}
	var opts []generator.GenerateOpt
	if args.IncludeVersion {
		opts = append(opts, generator.WithVersion(templ.Version()))
	}
	if args.IncludeTimestamp {
		opts = append(opts, generator.WithTimestamp(time.Now()))
	}
	if args.FileName != "" {
		return processSingleFile(ctx, w, "", args.FileName, nil, args.GenerateSourceMapVisualisations, opts)
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
			return err
		}
	}

	var p *proxy.Handler
	if args.Proxy != "" {
		p = proxy.New(args.ProxyPort, target)
	}
	fmt.Fprintln(w, "Processing path:", args.Path)

	if args.Watch {
		err = generateWatched(watchCtx, w, args, opts, p)
		if err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
	}

	return generateProduction(ctx, w, args, opts, p)
}

func generateWatched(ctx context.Context, w io.Writer, args Arguments, opts []generator.GenerateOpt, p *proxy.Handler) error {
	fmt.Fprintln(w, "Generating dev code:", args.Path)
	start := time.Now()

	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = time.Millisecond * 500
	bo.MaxInterval = time.Second * 3
	bo.MaxElapsedTime = 0

	var firstRunComplete bool
	fileNameToLastModTime := make(map[string]time.Time)
	fileNameToHash := make(map[string][sha256.Size]byte)

	for !firstRunComplete || args.Watch {
		changesFound, errs := processChanges(
			ctx, w,
			fileNameToLastModTime, fileNameToHash,
			args.Path, args.GenerateSourceMapVisualisations,
			opts, args.WorkerCount, true, args.KeepOrphanedFiles)
		if len(errs) > 0 {
			if errors.Is(errs[0], context.Canceled) {
				return errs[0]
			}
			if !args.Watch {
				return fmt.Errorf("failed to process path: %v", errors.Join(errs...))
			}
			logError(w, "Error processing path: %v\n", errors.Join(errs...))
		}
		if changesFound > 0 {
			if len(errs) > 0 {
				logError(w, "Generated code for %d templates with %d errors in %s\n", changesFound, len(errs), time.Since(start))
			} else {
				logSuccess(w, "Generated code for %d templates with %d errors in %s\n", changesFound, len(errs), time.Since(start))
			}
			if args.Command != "" {
				fmt.Fprintf(w, "Executing command: %s\n", args.Command)
				if _, err := run.Run(ctx, args.Path, args.Command); err != nil {
					fmt.Fprintf(w, "Error starting command: %v\n", err)
				}
			}
			// Send server-sent event.
			if p != nil {
				p.SendSSE("message", "reload")
			}

			if !firstRunComplete && p != nil {
				go func() {
					fmt.Fprintf(w, "Proxying from %s to target: %s\n", p.URL, p.Target.String())
					if err := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", args.ProxyPort), p); err != nil {
						fmt.Fprintf(w, "Error starting proxy: %v\n", err)
					}
				}()
				go func() {
					fmt.Fprintf(w, "Opening URL: %s\n", p.Target.String())
					if err := openURL(w, p.URL); err != nil {
						fmt.Fprintf(w, "Error opening URL: %v\n", err)
					}
				}()
			}
		}
		if err := checkTemplVersion(args.Path); err != nil {
			logWarning(w, "templ version check failed: %v\n", err)
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

	return nil
}

func generateProduction(ctx context.Context, w io.Writer, args Arguments, opts []generator.GenerateOpt, p *proxy.Handler) error {
	fmt.Fprintln(w, "Generating production code:", args.Path)
	start := time.Now()

	changesFound, errs := processChanges(
		ctx, w, nil, nil,
		args.Path, args.GenerateSourceMapVisualisations,
		opts, args.WorkerCount, false, args.KeepOrphanedFiles)
	if len(errs) > 0 {
		if errors.Is(errs[0], context.Canceled) {
			return errs[0]
		}
		logError(w, "Error processing path: %v\n", errors.Join(errs...))
	}

	if changesFound > 0 {
		if len(errs) > 0 {
			logError(w, "Generated code for %d templates with %d errors in %s\n", changesFound, len(errs), time.Since(start))
		} else {
			logSuccess(w, "Generated code for %d templates with %d errors in %s\n", changesFound, len(errs), time.Since(start))
		}
		if args.Command != "" {
			fmt.Fprintf(w, "Executing command: %s\n", args.Command)
			if _, err := run.Run(ctx, args.Path, args.Command); err != nil {
				fmt.Fprintf(w, "Error starting command: %v\n", err)
			}
		}
	}

	if err := checkTemplVersion(args.Path); err != nil {
		logWarning(w, "templ version check failed: %v\n", err)
	}

	return nil
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

func processChanges(ctx context.Context, stdout io.Writer, fileNameToLastModTime map[string]time.Time, hashes map[string][sha256.Size]byte, path string, generateSourceMapVisualisations bool, opts []generator.GenerateOpt, maxWorkerCount int, watching, keepOrphanedFiles bool) (changesFound int, errs []error) {
	sem := make(chan struct{}, maxWorkerCount)
	var wg sync.WaitGroup

	if watching {
		opts = append(opts, generator.WithExtractStrings())
	}

	if fileNameToLastModTime == nil {
		fileNameToLastModTime = make(map[string]time.Time)
	}

	err := filepath.WalkDir(path, func(fileName string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err = ctx.Err(); err != nil {
			return err
		}
		if info.IsDir() && shouldSkipDir(fileName) {
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}

		orphaned := !keepOrphanedFiles && strings.HasSuffix(fileName, "_templ.go")
		if orphaned {
			// Make sure the generated file is orphaned
			// by checking if the corresponding .templ file exists.
			if _, err := os.Stat(strings.TrimSuffix(fileName, "_templ.go") + ".templ"); err == nil {
				orphaned = false
			}
		}

		devTextFile := !watching && strings.HasSuffix(fileName, "_templ.txt")
		if orphaned || devTextFile {
			if err = os.Remove(fileName); err != nil {
				return fmt.Errorf("failed to remove file: %w", err)
			}
			logWarning(stdout, "Deleted file %q\n", fileName)
			return nil
		}

		if strings.HasSuffix(fileName, ".templ") {
			lastModTime := fileNameToLastModTime[fileName]
			fileInfo, err := info.Info()
			if err != nil {
				return fmt.Errorf("failed to get file info: %w", err)
			}
			if fileInfo.ModTime().After(lastModTime) {
				fileNameToLastModTime[fileName] = fileInfo.ModTime()
				changesFound++

				// Start a processor, but limit to maxWorkerCount.
				sem <- struct{}{}
				wg.Add(1)
				go func() {
					defer wg.Done()
					if err := processSingleFile(ctx, stdout, path, fileName, hashes, generateSourceMapVisualisations, opts); err != nil {
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

func openURL(w io.Writer, url string) error {
	backoff := backoff.NewExponentialBackOff()
	backoff.InitialInterval = time.Second
	var client http.Client
	client.Timeout = 1 * time.Second
	for {
		if _, err := client.Get(url); err == nil {
			break
		}
		d := backoff.NextBackOff()
		fmt.Fprintf(w, "Server not ready. Retrying in %v...\n", d)
		time.Sleep(d)
	}
	return browser.OpenURL(url)
}

// processSingleFile generates Go code for a single template.
// If a basePath is provided, the filename included in error messages is relative to it.
func processSingleFile(ctx context.Context, stdout io.Writer, basePath, fileName string, hashes map[string][sha256.Size]byte, generateSourceMapVisualisations bool, opts []generator.GenerateOpt) (err error) {
	start := time.Now()
	diag, err := generate(ctx, basePath, fileName, hashes, generateSourceMapVisualisations, opts)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	defer func() {
		_, _ = b.WriteTo(stdout)
	}()
	if len(diag) > 0 {
		logWarning(&b, "Generated code for %q in %s\n", fileName, time.Since(start))
		printDiagnostics(&b, fileName, diag)
		return nil
	}
	logSuccess(&b, "Generated code for %q in %s\n", fileName, time.Since(start))
	return nil
}

func printDiagnostics(w io.Writer, fileName string, diags []parser.Diagnostic) {
	for _, d := range diags {
		fmt.Fprint(w, "\t")
		logWarning(w, "%s (%d:%d)\n", d.Message, d.Range.From.Line, d.Range.From.Col)
	}
	fmt.Fprintln(w)
}

// generate Go code for a single template.
// If a basePath is provided, the filename included in error messages is relative to it.
func generate(ctx context.Context, basePath, fileName string, hashes map[string][sha256.Size]byte, generateSourceMapVisualisations bool, opts []generator.GenerateOpt) (diagnostics []parser.Diagnostic, err error) {
	if err = ctx.Err(); err != nil {
		return
	}

	if hashes == nil {
		hashes = make(map[string][sha256.Size]byte)
	}

	t, err := parser.Parse(fileName)
	if err != nil {
		return nil, fmt.Errorf("%s parsing error: %w", fileName, err)
	}
	targetFileName := strings.TrimSuffix(fileName, ".templ") + "_templ.go"

	// Only use relative filenames to the basepath for filenames in runtime error messages.
	errorMessageFileName := fileName
	if basePath != "" {
		errorMessageFileName, _ = filepath.Rel(basePath, fileName)
	}

	var b bytes.Buffer
	sourceMap, literals, err := generator.Generate(t, &b, append(opts, generator.WithFileName(errorMessageFileName))...)
	if err != nil {
		return nil, fmt.Errorf("%s generation error: %w", fileName, err)
	}

	formattedGoCode, err := format.Source(b.Bytes())
	if err != nil {
		return nil, fmt.Errorf("%s source formatting error: %w", fileName, err)
	}

	// Hash output, and write out the file if the goCodeHash has changed.
	goCodeHash := sha256.Sum256(formattedGoCode)
	if hashes[targetFileName] != goCodeHash {
		if err = os.WriteFile(targetFileName, formattedGoCode, 0o644); err != nil {
			return nil, fmt.Errorf("failed to write target file %q: %w", targetFileName, err)
		}
		hashes[targetFileName] = goCodeHash
	}

	// Add the txt file if it has changed.
	if len(literals) > 0 {
		txtFileName := strings.TrimSuffix(fileName, ".templ") + "_templ.txt"
		txtHash := sha256.Sum256([]byte(literals))
		if hashes[txtFileName] != txtHash {
			if err = os.WriteFile(txtFileName, []byte(literals), 0o644); err != nil {
				return nil, fmt.Errorf("failed to write string literal file %q: %w", txtFileName, err)
			}
			hashes[txtFileName] = txtHash
		}
	}

	if generateSourceMapVisualisations {
		err = generateSourceMapVisualisation(ctx, fileName, targetFileName, sourceMap)
	}
	return t.Diagnostics, err
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

func logError(w io.Writer, format string, a ...any) {
	logWithDecoration(w, "✗", color.FgRed, format, a...)
}

func logWarning(w io.Writer, format string, a ...any) {
	logWithDecoration(w, "!", color.FgYellow, format, a...)
}

func logSuccess(w io.Writer, format string, a ...any) {
	logWithDecoration(w, "✓", color.FgGreen, format, a...)
}

func logWithDecoration(w io.Writer, decoration string, col color.Attribute, format string, a ...any) {
	color.New(col).Fprintf(w, "(%s) ", decoration)
	fmt.Fprintf(w, format, a...)
}

// Replace "go 1.21.3" with "go 1.21" until https://github.com/golang/go/issues/61888 is fixed, see templ issue https://github.com/a-h/templ/issues/355
var goVersionRegexp = regexp.MustCompile(`\ngo (\d+\.\d+)(?:\D.+)\n`)

func patchGoVersion(moduleFileContents []byte) []byte {
	return goVersionRegexp.ReplaceAll(moduleFileContents, []byte("\ngo $1\n"))
}

func checkTemplVersion(dir string) error {
	// Walk up the directory tree, starting at dir, until we find a go.mod file.
	// If it contains a go.mod file, parse it and find the templ version.
	dir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	for {
		current := filepath.Join(dir, "go.mod")
		_, err := os.Stat(current)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to stat go.mod file: %w", err)
		}
		if os.IsNotExist(err) {
			// Move up.
			prev := dir
			dir = filepath.Dir(dir)
			if dir == prev {
				return fmt.Errorf("could not find go.mod file")
			}
			continue
		}
		// Found a go.mod file.
		// Read it and find the templ version.
		m, err := os.ReadFile(current)
		if err != nil {
			return fmt.Errorf("failed to read go.mod file: %w", err)
		}

		// Replace "go 1.21.x" with "go 1.21".
		m = patchGoVersion(m)

		mf, err := modfile.Parse(current, m, nil)
		if err != nil {
			return fmt.Errorf("failed to parse go.mod file: %w", err)
		}
		if mf.Module.Mod.Path == "github.com/a-h/templ" {
			// The go.mod file is for templ itself.
			return nil
		}
		for _, r := range mf.Require {
			if r.Mod.Path == "github.com/a-h/templ" {
				cmp := semver.Compare(r.Mod.Version, templ.Version())
				if cmp < 0 {
					return fmt.Errorf("generator %v is newer than templ version %v found in go.mod file, consider running `go get -u github.com/a-h/templ` to upgrade", templ.Version(), r.Mod.Version)
				}
				if cmp > 0 {
					return fmt.Errorf("generator %v is older than templ version %v found in go.mod file, consider upgrading templ CLI", templ.Version(), r.Mod.Version)
				}
				return nil
			}
		}
	}
}
