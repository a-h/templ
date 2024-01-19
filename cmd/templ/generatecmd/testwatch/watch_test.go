package testwatch

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/a-h/templ/cmd/templ/generatecmd/modcheck"
	"github.com/a-h/templ/cmd/templ/generatecmd/run"
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

func TestWatch(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not find working dir: %v", err)
	}
	moduleRoot, err := modcheck.WalkUp(wd)
	if err != nil {
		t.Fatalf("could not find local templ go.mod file: %v", err)
	}

	dir, err := createTestProject(moduleRoot)
	if err != nil {
		t.Fatalf("failed to create test project: %v", err)
	}

	port, err := getPort()
	if err != nil {
		t.Fatalf("failed to get available port: %v", err)
	}

	ctx := context.Background()
	t.Run("can start the test app", func(t *testing.T) {
		cmd, err := run.Run(ctx, dir, fmt.Sprintf("go run . -port %d", port))
		if err != nil {
			t.Fatalf("failed to start test app: %v", err)
		}

		url := fmt.Sprintf("http://localhost:%d", port)

		// Wait for server to start.
		time.Sleep(time.Second)

		t.Run("can read count", func(t *testing.T) {
			doc, err := getHTML(url)
			if err != nil {
				t.Fatalf("failed to read HTML: %v", err)
			}
			if actualCount := doc.Find(`div[data-testid="count"]`).Text(); actualCount != "1" {
				t.Errorf("expected count 2, got %s", actualCount)
			}
		})

		defer func(cmd *exec.Cmd) {
			if err = run.Stop(cmd); err != nil {
				t.Errorf("failed to stop test server: %v", err)
			}
		}(cmd)
	})

	if err = os.RemoveAll(dir); err != nil {
		t.Fatalf("failed to remove test dir %q: %v", dir, err)
	}
}
