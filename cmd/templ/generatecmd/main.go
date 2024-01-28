package generatecmd

import (
	"context"
	_ "embed"
	"io"
	"log/slog"
	"runtime"

	_ "net/http/pprof"
)

type Arguments struct {
	FileName                        string
	Path                            string
	Watch                           bool
	OpenBrowser                     bool
	Command                         string
	ProxyPort                       int
	Proxy                           string
	WorkerCount                     int
	GenerateSourceMapVisualisations bool
	IncludeVersion                  bool
	IncludeTimestamp                bool
	Level                           string
	// PPROFPort is the port to run the pprof server on.
	PPROFPort         int
	KeepOrphanedFiles bool
}

var defaultWorkerCount = runtime.NumCPU()

func Run(ctx context.Context, stderr io.Writer, args Arguments) (err error) {
	level := slog.LevelWarn.Level()
	if args.Level == "debug" || args.Level == "verbose" {
		level = slog.LevelDebug.Level()
	}
	if args.Level == "info" {
		level = slog.LevelInfo.Level()
	}
	log := slog.New(slog.NewTextHandler(stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	}))
	return NewGenerate(log, args).Run(ctx)
}
