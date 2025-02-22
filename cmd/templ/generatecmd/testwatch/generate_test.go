package testwatch

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/a-h/templ/cmd/templ/generatecmd"
	"github.com/a-h/templ/cmd/templ/generatecmd/modcheck"
	"github.com/a-h/templ/internal/htmlfind"
	"golang.org/x/net/html"
)

//go:embed testdata/*
var testdata embed.FS

func createTestProject(moduleRoot string) (dir string, err error) {
	dir, err = os.MkdirTemp("", "templ_watch_test_*")
	if err != nil {
		return dir, fmt.Errorf("failed to make test dir: %w", err)
	}
	files, err := testdata.ReadDir("testdata")
	if err != nil {
		return dir, fmt.Errorf("failed to read embedded dir: %w", err)
	}
	for _, file := range files {
		src := filepath.Join("testdata", file.Name())
		data, err := testdata.ReadFile(src)
		if err != nil {
			return dir, fmt.Errorf("failed to read file: %w", err)
		}

		target := filepath.Join(dir, file.Name())
		if file.Name() == "go.mod.embed" {
			data = bytes.ReplaceAll(data, []byte("{moduleRoot}"), []byte(moduleRoot))
			target = filepath.Join(dir, "go.mod")
		}
		err = os.WriteFile(target, data, 0660)
		if err != nil {
			return dir, fmt.Errorf("failed to copy file: %w", err)
		}
	}
	return dir, nil
}

func replaceInFile(name, src, tgt string) error {
	data, err := os.ReadFile(name)
	if err != nil {
		return err
	}
	updated := strings.Replace(string(data), src, tgt, -1)
	return os.WriteFile(name, []byte(updated), 0660)
}

func getPort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}

func getHTML(url string) (n *html.Node, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get %q: %w", url, err)
	}
	defer resp.Body.Close()
	return html.Parse(resp.Body)
}

func TestCanAccessDirect(t *testing.T) {
	if testing.Short() {
		return
	}
	args, teardown, err := Setup(false)
	if err != nil {
		t.Fatalf("failed to setup test: %v", err)
	}
	defer teardown(t)

	// Assert.
	doc, err := getHTML(args.AppURL)
	if err != nil {
		t.Fatalf("failed to read HTML: %v", err)
	}
	countElements := htmlfind.All(doc, htmlfind.Element("div", htmlfind.Attr("data-testid", "count")))
	if len(countElements) != 1 {
		t.Fatalf("expected 1 count element, got %d", len(countElements))
	}
	countText := countElements[0].FirstChild.Data
	actualCount, err := strconv.Atoi(countText)
	if err != nil {
		t.Fatalf("got count %q instead of integer", countText)
	}
	if actualCount < 1 {
		t.Errorf("expected count >= 1, got %d", actualCount)
	}
}

func TestCanAccessViaProxy(t *testing.T) {
	if testing.Short() {
		return
	}
	args, teardown, err := Setup(false)
	if err != nil {
		t.Fatalf("failed to setup test: %v", err)
	}
	defer teardown(t)

	// Assert.
	doc, err := getHTML(args.ProxyURL)
	if err != nil {
		t.Fatalf("failed to read HTML: %v", err)
	}
	countElements := htmlfind.All(doc, htmlfind.Element("div", htmlfind.Attr("data-testid", "count")))
	if len(countElements) != 1 {
		t.Fatalf("expected 1 count element, got %d", len(countElements))
	}
	countText := countElements[0].FirstChild.Data
	actualCount, err := strconv.Atoi(countText)
	if err != nil {
		t.Fatalf("got count %q instead of integer", countText)
	}
	if actualCount < 1 {
		t.Errorf("expected count >= 1, got %d", actualCount)
	}
}

type Event struct {
	Type string
	Data string
}

func readSSE(ctx context.Context, url string, sse chan<- Event) (err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Connection", "keep-alive")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	var e Event
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			sse <- e
			e = Event{}
			continue
		}
		if strings.HasPrefix(line, "event: ") {
			e.Type = line[len("event: "):]
		}
		if strings.HasPrefix(line, "data: ") {
			e.Data = line[len("data: "):]
		}
	}
	return scanner.Err()
}

func TestFileModificationsResultInSSEWithGzip(t *testing.T) {
	if testing.Short() {
		return
	}
	args, teardown, err := Setup(false)
	if err != nil {
		t.Fatalf("failed to setup test: %v", err)
	}
	defer teardown(t)

	// Start the SSE check.
	events := make(chan Event)
	var eventsErr error
	go func() {
		eventsErr = readSSE(context.Background(), fmt.Sprintf("%s/_templ/reload/events", args.ProxyURL), events)
	}()

	// Assert data is expected.
	doc, err := getHTML(args.ProxyURL)
	if err != nil {
		t.Fatalf("failed to read HTML: %v", err)
	}
	modified := htmlfind.All(doc, htmlfind.Element("div", htmlfind.Attr("data-testid", "modification")))
	if len(modified) != 1 {
		t.Fatalf("expected 1 modification element, got %d", len(modified))
	}
	if text := modified[0].FirstChild.Data; text != "Original" {
		t.Errorf("expected %q, got %q", "Original", text)
	}

	// Change file.
	templFile := filepath.Join(args.AppDir, "templates.templ")
	err = replaceInFile(templFile,
		`<div data-testid="modification">Original</div>`,
		`<div data-testid="modification">Updated</div>`)
	if err != nil {
		t.Errorf("failed to replace text in file: %v", err)
	}

	// Give the filesystem watcher a few seconds.
	var reloadCount int
loop:
	for {
		select {
		case event := <-events:
			if event.Data == "reload" {
				reloadCount++
				break loop
			}
		case <-time.After(time.Second * 5):
			break loop
		}
	}
	if reloadCount == 0 {
		t.Error("failed to receive SSE about update after 5 seconds")
	}

	// Check to see if there were any errors.
	if eventsErr != nil {
		t.Errorf("error reading events: %v", err)
	}

	// See results in browser immediately.
	doc, err = getHTML(args.ProxyURL)
	if err != nil {
		t.Fatalf("failed to read HTML: %v", err)
	}
	modified = htmlfind.All(doc, htmlfind.Element("div", htmlfind.Attr("data-testid", "modification")))
	if len(modified) != 1 {
		t.Fatalf("expected 1 modification element, got %d", len(modified))
	}
	if text := modified[0].FirstChild.Data; text != "Updated" {
		t.Errorf("expected %q, got %q", "Updated", text)
	}
}

func TestFileModificationsResultInSSE(t *testing.T) {
	if testing.Short() {
		return
	}
	args, teardown, err := Setup(false)
	if err != nil {
		t.Fatalf("failed to setup test: %v", err)
	}
	defer teardown(t)

	// Start the SSE check.
	events := make(chan Event)
	var eventsErr error
	go func() {
		eventsErr = readSSE(context.Background(), fmt.Sprintf("%s/_templ/reload/events", args.ProxyURL), events)
	}()

	// Assert data is expected.
	doc, err := getHTML(args.ProxyURL)
	if err != nil {
		t.Fatalf("failed to read HTML: %v", err)
	}
	modified := htmlfind.All(doc, htmlfind.Element("div", htmlfind.Attr("data-testid", "modification")))
	if len(modified) != 1 {
		t.Fatalf("expected 1 modification element, got %d", len(modified))
	}
	if text := modified[0].FirstChild.Data; text != "Original" {
		t.Errorf("expected %q, got %q", "Original", text)
	}

	// Change file.
	templFile := filepath.Join(args.AppDir, "templates.templ")
	err = replaceInFile(templFile,
		`<div data-testid="modification">Original</div>`,
		`<div data-testid="modification">Updated</div>`)
	if err != nil {
		t.Errorf("failed to replace text in file: %v", err)
	}

	// Give the filesystem watcher a few seconds.
	var reloadCount int
loop:
	for {
		select {
		case event := <-events:
			if event.Data == "reload" {
				reloadCount++
				break loop
			}
		case <-time.After(time.Second * 5):
			break loop
		}
	}
	if reloadCount == 0 {
		t.Error("failed to receive SSE about update after 5 seconds")
	}

	// Check to see if there were any errors.
	if eventsErr != nil {
		t.Errorf("error reading events: %v", err)
	}

	// See results in browser immediately.
	doc, err = getHTML(args.ProxyURL)
	if err != nil {
		t.Fatalf("failed to read HTML: %v", err)
	}
	modified = htmlfind.All(doc, htmlfind.Element("div", htmlfind.Attr("data-testid", "modification")))
	if len(modified) != 1 {
		t.Fatalf("expected 1 modification element, got %d", len(modified))
	}
	if text := modified[0].FirstChild.Data; text != "Updated" {
		t.Errorf("expected %q, got %q", "Updated", text)
	}
}

func NewTestArgs(modRoot, appDir string, appPort int, proxyBind string, proxyPort int) TestArgs {
	return TestArgs{
		ModRoot:   modRoot,
		AppDir:    appDir,
		AppPort:   appPort,
		AppURL:    fmt.Sprintf("http://localhost:%d", appPort),
		ProxyBind: proxyBind,
		ProxyPort: proxyPort,
		ProxyURL:  fmt.Sprintf("http://%s:%d", proxyBind, proxyPort),
	}
}

type TestArgs struct {
	ModRoot   string
	AppDir    string
	AppPort   int
	AppURL    string
	ProxyBind string
	ProxyPort int
	ProxyURL  string
}

func Setup(gzipEncoding bool) (args TestArgs, teardown func(t *testing.T), err error) {
	wd, err := os.Getwd()
	if err != nil {
		return args, teardown, fmt.Errorf("could not find working dir: %w", err)
	}
	moduleRoot, err := modcheck.WalkUp(wd)
	if err != nil {
		return args, teardown, fmt.Errorf("could not find local templ go.mod file: %v", err)
	}

	appDir, err := createTestProject(moduleRoot)
	if err != nil {
		return args, teardown, fmt.Errorf("failed to create test project: %v", err)
	}
	appPort, err := getPort()
	if err != nil {
		return args, teardown, fmt.Errorf("failed to get available port: %v", err)
	}
	proxyPort, err := getPort()
	if err != nil {
		return args, teardown, fmt.Errorf("failed to get available port: %v", err)
	}
	proxyBind := "localhost"

	args = NewTestArgs(moduleRoot, appDir, appPort, proxyBind, proxyPort)
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	var cmdErr error

	wg.Add(1)
	go func() {
		defer wg.Done()

		command := fmt.Sprintf("go run . -port %d", args.AppPort)
		if gzipEncoding {
			command += " -gzip true"
		}

		log := slog.New(slog.NewJSONHandler(io.Discard, nil))

		cmdErr = generatecmd.Run(ctx, log, generatecmd.Arguments{
			Path:                            appDir,
			Watch:                           true,
			OpenBrowser:                     false,
			Command:                         command,
			ProxyBind:                       proxyBind,
			ProxyPort:                       proxyPort,
			Proxy:                           args.AppURL,
			NotifyProxy:                     false,
			WorkerCount:                     0,
			GenerateSourceMapVisualisations: false,
			IncludeVersion:                  false,
			IncludeTimestamp:                false,
			PPROFPort:                       0,
			KeepOrphanedFiles:               false,
		})
	}()

	// Wait for server to start.
	if err = waitForURL(args.AppURL); err != nil {
		cancel()
		wg.Wait()
		return args, teardown, fmt.Errorf("failed to start app server, command error %v: %w", cmdErr, err)
	}
	if err = waitForURL(args.ProxyURL); err != nil {
		cancel()
		wg.Wait()
		return args, teardown, fmt.Errorf("failed to start proxy server, command error %v: %w", cmdErr, err)
	}

	// Wait for exit.
	teardown = func(t *testing.T) {
		cancel()
		wg.Wait()
		if cmdErr != nil {
			t.Errorf("failed to run generate cmd: %v", err)
		}

		if err = os.RemoveAll(appDir); err != nil {
			t.Fatalf("failed to remove test dir %q: %v", appDir, err)
		}
	}
	return args, teardown, err
}

func waitForURL(url string) (err error) {
	var tries int
	for {
		time.Sleep(time.Second)
		if tries > 20 {
			return err
		}
		tries++
		var resp *http.Response
		resp, err = http.Get(url)
		if err != nil {
			fmt.Printf("failed to get %q: %v\n", url, err)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("failed to get %q: %v\n", url, err)
			err = fmt.Errorf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
			continue
		}
		return nil
	}
}

func TestGenerateReturnsErrors(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not find working dir: %v", err)
	}
	moduleRoot, err := modcheck.WalkUp(wd)
	if err != nil {
		t.Fatalf("could not find local templ go.mod file: %v", err)
	}

	appDir, err := createTestProject(moduleRoot)
	if err != nil {
		t.Fatalf("failed to create test project: %v", err)
	}
	defer func() {
		if err = os.RemoveAll(appDir); err != nil {
			t.Fatalf("failed to remove test dir %q: %v", appDir, err)
		}
	}()

	// Break the HTML.
	templFile := filepath.Join(appDir, "templates.templ")
	err = replaceInFile(templFile,
		`<div data-testid="modification">Original</div>`,
		`<div data-testid="modification" -unclosed div-</div>`)
	if err != nil {
		t.Errorf("failed to replace text in file: %v", err)
	}

	log := slog.New(slog.NewJSONHandler(io.Discard, nil))

	// Run.
	err = generatecmd.Run(context.Background(), log, generatecmd.Arguments{
		Path:              appDir,
		Watch:             false,
		IncludeVersion:    false,
		IncludeTimestamp:  false,
		KeepOrphanedFiles: false,
	})
	if err == nil {
		t.Errorf("expected generation error, got %v", err)
	}
}
