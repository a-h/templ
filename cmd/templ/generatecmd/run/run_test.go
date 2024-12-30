package run_test

import (
	"context"
	"embed"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/a-h/templ/cmd/templ/generatecmd/run"
)

//go:embed testprogram/*
var testprogram embed.FS

func TestGoRun(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}

	// Copy testprogram to a temporary directory.
	dir, err := os.MkdirTemp("", "testprogram")
	if err != nil {
		t.Fatalf("failed to make test dir: %v", err)
	}
	files, err := testprogram.ReadDir("testprogram")
	if err != nil {
		t.Fatalf("failed to read embedded dir: %v", err)
	}
	for _, file := range files {
		srcFileName := "testprogram/" + file.Name()
		srcData, err := testprogram.ReadFile(srcFileName)
		if err != nil {
			t.Fatalf("failed to read src file %q: %v", srcFileName, err)
		}
		tgtFileName := filepath.Join(dir, file.Name())
		tgtFile, err := os.Create(tgtFileName)
		if err != nil {
			t.Fatalf("failed to create tgt file %q: %v", tgtFileName, err)
		}
		defer tgtFile.Close()
		if _, err := tgtFile.Write(srcData); err != nil {
			t.Fatalf("failed to write to tgt file %q: %v", tgtFileName, err)
		}
	}
	// Rename the go.mod.embed file to go.mod.
	if err := os.Rename(filepath.Join(dir, "go.mod.embed"), filepath.Join(dir, "go.mod")); err != nil {
		t.Fatalf("failed to rename go.mod.embed: %v", err)
	}

	tests := []struct {
		name string
		cmd  string
	}{
		{
			name: "Well behaved programs get shut down",
			cmd:  "go run .",
		},
		{
			name: "Badly behaved programs get shut down",
			cmd:  "go run . -badly-behaved",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cmd, err := run.Run(ctx, dir, tt.cmd)
			if err != nil {
				t.Fatalf("failed to run program: %v", err)
			}

			time.Sleep(1 * time.Second)

			pid := cmd.Process.Pid

			if err := run.KillAll(); err != nil {
				t.Fatalf("failed to kill all: %v", err)
			}

			// Check the parent process is no longer running.
			if err := cmd.Process.Signal(os.Signal(syscall.Signal(0))); err == nil {
				t.Fatalf("process %d is still running", pid)
			}
			// Check that the child was stopped.
			body, err := readResponse("http://localhost:7777")
			if err == nil {
				t.Fatalf("child process is still running: %s", body)
			}
		})
	}
}

func readResponse(url string) (body string, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return body, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return body, err
	}
	return string(b), nil
}
