package generatecmd

import (
	"context"
	_ "embed"
	"io"
	"log/slog"
	"runtime"

	_ "net/http/pprof"

	"github.com/a-h/templ/cmd/templ/sloghandler"
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
	if args.Level == "debug" {
		level = slog.LevelDebug.Level()
	}
	if args.Level == "info" {
		level = slog.LevelInfo.Level()
	}
	// The built-in attributes with keys "time", "level", "source", and "msg"
	// are passed to this function, except that time is omitted
	// if zero, and source is omitted if AddSource is false.
	log := slog.New(sloghandler.NewHandler(stderr, &slog.HandlerOptions{
		AddSource: args.Level == "debug",
		Level:     level,
	}))
	return NewGenerate(log, args).Run(ctx)
}
