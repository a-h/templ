package diagnosecmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/a-h/templ"
	"github.com/a-h/templ/cmd/templ/lspcmd/pls"
)

type Arguments struct {
}

type diagnostic struct {
	Location string
	Version  string
	OK       bool
	Message  string
}

func diagnoseGo() (d diagnostic) {
	// Find Go.
	var err error
	d.Location, err = exec.LookPath("go")
	if err != nil {
		d.Message = fmt.Sprintf("failed to find go: %v", err)
		return
	}
	// Run go to find the version.
	cmd := exec.Command(d.Location, "version")
	v, err := cmd.Output()
	if err != nil {
		d.Message = fmt.Sprintf("failed to get go version, check that Go is installed: %v", err)
		return
	}
	d.Version = strings.TrimSpace(string(v))
	d.OK = true
	return
}

func diagnoseGopls() (d diagnostic) {
	var err error
	d.Location, err = pls.FindGopls()
	if err != nil {
		d.Message = fmt.Sprintf("failed to find gopls: %v", err)
		return
	}
	cmd := exec.Command(d.Location, "version")
	v, err := cmd.Output()
	if err != nil {
		d.Message = fmt.Sprintf("failed to get gopls version: %v", err)
		return
	}
	d.Version = strings.TrimSpace(string(v))
	d.OK = true
	return
}

func diagnoseTempl() (d diagnostic) {
	// Find templ.
	var err error
	d.Location, err = findTempl()
	if err != nil {
		d.Message = err.Error()
		return
	}
	// Run templ to find the version.
	cmd := exec.Command(d.Location, "version")
	v, err := cmd.Output()
	if err != nil {
		d.Message = fmt.Sprintf("failed to get templ version: %v", err)
		return
	}
	d.Version = strings.TrimSpace(string(v))
	if d.Version != templ.Version() {
		d.Message = fmt.Sprintf("version mismatch - you're running %q at the command line, but the version in the path is %q", templ.Version(), d.Version)
		return
	}
	d.OK = true
	return
}

func findTempl() (location string, err error) {
	executableName := "templ"
	if runtime.GOOS == "windows" {
		executableName = "templ.exe"
	}
	executableName, err = exec.LookPath(executableName)
	if err == nil {
		// Found on the path.
		return executableName, nil
	}

	// Unexpected error.
	if !errors.Is(err, exec.ErrNotFound) {
		return "", fmt.Errorf("unexpected error looking for templ: %w", err)
	}

	return "", fmt.Errorf("templ is not in the path (%q). You can install templ with `go install github.com/a-h/templ/cmd/templ@latest`", os.Getenv("PATH"))
}

func Run(ctx context.Context, log *slog.Logger, args Arguments) (err error) {
	log.Info("os", slog.String("name", runtime.GOOS), slog.String("arch", runtime.GOARCH))
	logDiagnostic(ctx, log, "go", diagnoseGo())
	logDiagnostic(ctx, log, "gopls", diagnoseGopls())
	logDiagnostic(ctx, log, "templ", diagnoseTempl())
	return nil
}

func logDiagnostic(ctx context.Context, log *slog.Logger, name string, d diagnostic) {
	level := slog.LevelInfo
	if !d.OK {
		level = slog.LevelError
	}
	args := []any{
		slog.String("location", d.Location),
		slog.String("version", d.Version),
	}
	if d.Message != "" {
		args = append(args, slog.String("message", d.Message))
	}
	log.Log(ctx, level, name, args...)
}
