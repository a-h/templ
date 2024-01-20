package testwatch

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/a-h/templ/cmd/templ/generatecmd"
	"github.com/a-h/templ/cmd/templ/generatecmd/modcheck"
)

//go:embed testproject/*
var testproject embed.FS

func createTestProject(moduleRoot string) (dir string, err error) {
	dir, err = os.MkdirTemp("", "templ_watch_test_*")
	if err != nil {
		return dir, fmt.Errorf("failed to make test dir: %w", err)
	}
	files, err := testproject.ReadDir("testproject")
	if err != nil {
		return dir, fmt.Errorf("failed to read embedded dir: %w", err)
	}
	for _, file := range files {
		src := filepath.Join("testproject", file.Name())
		data, err := testproject.ReadFile(src)
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

func getHTML(url string) (doc *goquery.Document, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get %q: %w", url, err)
	}
	return goquery.NewDocumentFromReader(resp.Body)
}

func TestCanAccessDirect(t *testing.T) {
	if testing.Short() {
		return
	}
	args, teardown, err := Setup()
	if err != nil {
		t.Fatalf("failed to setup test: %v", err)
	}
	defer teardown(t)

	// Assert.
	doc, err := getHTML(args.AppURL)
	if err != nil {
		t.Fatalf("failed to read HTML: %v", err)
	}
	countText := doc.Find(`div[data-testid="count"]`).Text()
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
	args, teardown, err := Setup()
	if err != nil {
		t.Fatalf("failed to setup test: %v", err)
	}
	defer teardown(t)

	// Assert.
	doc, err := getHTML(args.ProxyURL)
	if err != nil {
		t.Fatalf("failed to read HTML: %v", err)
	}
	countText := doc.Find(`div[data-testid="count"]`).Text()
	actualCount, err := strconv.Atoi(countText)
	if err != nil {
		t.Fatalf("got count %q instead of integer", countText)
	}
	if actualCount < 1 {
		t.Errorf("expected count >= 1, got %d", actualCount)
	}
}

func NewTestArgs(modRoot, appDir string, appPort int, proxyPort int) TestArgs {
	return TestArgs{
		ModRoot:   modRoot,
		AppDir:    appDir,
		AppPort:   appPort,
		AppURL:    fmt.Sprintf("http://localhost:%d", appPort),
		ProxyPort: proxyPort,
		ProxyURL:  fmt.Sprintf("http://localhost:%d", proxyPort),
	}
}

type TestArgs struct {
	ModRoot   string
	AppDir    string
	AppPort   int
	AppURL    string
	ProxyPort int
	ProxyURL  string
}

func Setup() (args TestArgs, teardown func(t *testing.T), err error) {
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

	args = NewTestArgs(moduleRoot, appDir, appPort, proxyPort)
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	var cmdErr error

	wg.Add(1)
	go func() {
		defer wg.Done()
		cmdErr = generatecmd.Run(ctx, os.Stdout, generatecmd.Arguments{
			Path:              appDir,
			Watch:             true,
			Command:           fmt.Sprintf("go run . -port %d", args.AppPort),
			ProxyPort:         proxyPort,
			Proxy:             args.AppURL,
			IncludeVersion:    false,
			IncludeTimestamp:  false,
			KeepOrphanedFiles: false,
		})
	}()

	// Wait for server to start.
	time.Sleep(time.Second)

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
