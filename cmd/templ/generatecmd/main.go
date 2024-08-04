package generatecmd

import (
	"context"
	_ "embed"
	"log/slog"

	_ "net/http/pprof"
)

type Arguments struct {
	FileName                        string
	FileWriter                      FileWriterFunc
	Path                            string
	Watch                           bool
	OpenBrowser                     bool
	Command                         string
	ProxyBind                       string
	ProxyPort                       int
	Proxy                           string
	NotifyProxy                     bool
	WorkerCount                     int
	GenerateSourceMapVisualisations bool
	IncludeVersion                  bool
	IncludeTimestamp                bool
	// PPROFPort is the port to run the pprof server on.
	PPROFPort         int
	KeepOrphanedFiles bool
	Lazy              bool
}

func Run(ctx context.Context, log *slog.Logger, args Arguments) (err error) {
	return NewGenerate(log, args).Run(ctx)
}
