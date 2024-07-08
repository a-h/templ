package infocmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/a-h/templ"
	"github.com/a-h/templ/cmd/templ/lspcmd/pls"
)

type Arguments struct {
	JSON bool `flag:"json" help:"Output info as JSON."`
}

type Info struct {
	OS struct {
		GOOS   string `json:"goos"`
		GOARCH string `json:"goarch"`
	} `json:"os"`
	Go    ToolInfo `json:"go"`
	Gopls ToolInfo `json:"gopls"`
	Templ ToolInfo `json:"templ"`
}

type ToolInfo struct {
	Location string `json:"location"`
	Version  string `json:"version"`
	OK       bool   `json:"ok"`
	Message  string `json:"message,omitempty"`
}

func getGoInfo() (d ToolInfo) {
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

func getGoplsInfo() (d ToolInfo) {
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

func getTemplInfo() (d ToolInfo) {
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

func getInfo() (d Info) {
	d.OS.GOOS = runtime.GOOS
	d.OS.GOARCH = runtime.GOARCH
	d.Go = getGoInfo()
	d.Gopls = getGoplsInfo()
	d.Templ = getTemplInfo()
	return
}

func Run(ctx context.Context, log *slog.Logger, stdout io.Writer, args Arguments) (err error) {
	info := getInfo()
	if args.JSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(info)
	}
	log.Info("os", slog.String("goos", info.OS.GOOS), slog.String("goarch", info.OS.GOARCH))
	logInfo(ctx, log, "go", info.Go)
	logInfo(ctx, log, "gopls", info.Gopls)
	logInfo(ctx, log, "templ", info.Templ)
	return nil
}

func logInfo(ctx context.Context, log *slog.Logger, name string, ti ToolInfo) {
	level := slog.LevelInfo
	if !ti.OK {
		level = slog.LevelError
	}
	args := []any{
		slog.String("location", ti.Location),
		slog.String("version", ti.Version),
	}
	if ti.Message != "" {
		args = append(args, slog.String("message", ti.Message))
	}
	log.Log(ctx, level, name, args...)
}
