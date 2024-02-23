package generatecmd

import (
	"context"
	_ "embed"
	"io"
	"log/slog"

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
	LogLevel                        string
	// PPROFPort is the port to run the pprof server on.
	PPROFPort         int
	KeepOrphanedFiles bool
}

func Run(ctx context.Context, w io.Writer, args Arguments) (err error) {
	level := slog.LevelInfo.Level()
	switch args.LogLevel {
	case "debug":
		level = slog.LevelDebug.Level()
	case "warn":
		level = slog.LevelWarn.Level()
	case "error":
		level = slog.LevelError.Level()
	}
	log := slog.New(sloghandler.NewHandler(w, &slog.HandlerOptions{
		AddSource: args.LogLevel == "debug",
		Level:     level,
	}))
	return NewGenerate(log, args).Run(ctx)
}
