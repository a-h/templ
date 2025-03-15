package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"runtime"

	"github.com/a-h/templ"
	"github.com/a-h/templ/cmd/templ/fmtcmd"
	"github.com/a-h/templ/cmd/templ/generatecmd"
	"github.com/a-h/templ/cmd/templ/infocmd"
	"github.com/a-h/templ/cmd/templ/lspcmd"
	"github.com/a-h/templ/cmd/templ/sloghandler"
	"github.com/fatih/color"
)

func main() {
	code := run(os.Stdin, os.Stdout, os.Stderr, os.Args)
	if code != 0 {
		os.Exit(code)
	}
}

const usageText = `usage: templ <command> [<args>...]

templ - build HTML UIs with Go

See docs at https://templ.guide

commands:
  generate   Generates Go code from templ files
  fmt        Formats templ files
  lsp        Starts a language server for templ files
  info       Displays information about the templ environment
  version    Prints the version
`

func run(stdin io.Reader, stdout, stderr io.Writer, args []string) (code int) {
	if len(args) < 2 {
		fmt.Fprint(stderr, usageText)
		return 64 // EX_USAGE
	}
	switch args[1] {
	case "info":
		return infoCmd(stdout, stderr, args[2:])
	case "generate":
		return generateCmd(stdout, stderr, args[2:])
	case "fmt":
		return fmtCmd(stdin, stdout, stderr, args[2:])
	case "lsp":
		return lspCmd(stdin, stdout, stderr, args[2:])
	case "version", "--version":
		fmt.Fprintln(stdout, templ.Version())
		return 0
	case "help", "-help", "--help", "-h":
		fmt.Fprint(stdout, usageText)
		return 0
	}
	fmt.Fprint(stderr, usageText)
	return 64 // EX_USAGE
}

func newLogger(logLevel string, verbose bool, stderr io.Writer) *slog.Logger {
	if verbose {
		logLevel = "debug"
	}
	level := slog.LevelInfo.Level()
	switch logLevel {
	case "debug":
		level = slog.LevelDebug.Level()
	case "warn":
		level = slog.LevelWarn.Level()
	case "error":
		level = slog.LevelError.Level()
	}
	return slog.New(sloghandler.NewHandler(stderr, &slog.HandlerOptions{
		AddSource: logLevel == "debug",
		Level:     level,
	}))
}

const infoUsageText = `usage: templ info [<args>...]

Displays information about the templ environment.

Args:
  -json
    Output information in JSON format to stdout. (default false)
  -v
    Set log verbosity level to "debug". (default "info")
  -log-level
    Set log verbosity level. (default "info", options: "debug", "info", "warn", "error")
  -help
    Print help and exit.
`

func infoCmd(stdout, stderr io.Writer, args []string) (code int) {
	cmd := flag.NewFlagSet("diagnose", flag.ExitOnError)
	jsonFlag := cmd.Bool("json", false, "")
	verboseFlag := cmd.Bool("v", false, "")
	logLevelFlag := cmd.String("log-level", "info", "")
	helpFlag := cmd.Bool("help", false, "")
	err := cmd.Parse(args)
	if err != nil {
		fmt.Fprint(stderr, infoUsageText)
		return 64 // EX_USAGE
	}
	if *helpFlag {
		fmt.Fprint(stdout, infoUsageText)
		return
	}

	log := newLogger(*logLevelFlag, *verboseFlag, stderr)

	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		fmt.Fprintln(stderr, "Stopping...")
		cancel()
	}()

	err = infocmd.Run(ctx, log, stdout, infocmd.Arguments{
		JSON: *jsonFlag,
	})
	if err != nil {
		color.New(color.FgRed).Fprint(stderr, "(✗) ")
		fmt.Fprintln(stderr, "Command failed: "+err.Error())
		return 1
	}
	return 0
}

const generateUsageText = `usage: templ generate [<args>...]

Generates Go code from templ files.

Args:
  -path <path>
    Generates code for all files in path. (default .)
  -f <file>
    Optionally generates code for a single file, e.g. -f header.templ
  -stdout
    Prints to stdout instead of writing generated files to the filesystem.
    Only applicable when -f is used.
  -source-map-visualisations
    Set to true to generate HTML files to visualise the templ code and its corresponding Go code.
  -include-version
    Set to false to skip inclusion of the templ version in the generated code. (default true)
  -include-timestamp
    Set to true to include the current time in the generated code.
  -watch
    Set to true to watch the path for changes and regenerate code.
  -watch-pattern <regexp>
    Set the regexp pattern of files that will be watched for changes. (default: '(.+\.go$)|(.+\.templ$)|(.+_templ\.txt$)')
  -cmd <cmd>
    Set the command to run after generating code.
  -proxy
    Set the URL to proxy after generating code and executing the command.
  -proxyport
    The port the proxy will listen on. (default 7331)
  -proxybind
    The address the proxy will listen on. (default 127.0.0.1)
  -notify-proxy
    If present, the command will issue a reload event to the proxy 127.0.0.1:7331, or use proxyport and proxybind to specify a different address.
  -w
    Number of workers to use when generating code. (default runtime.NumCPUs)
  -lazy
    Only generate .go files if the source .templ file is newer.	
  -pprof
    Port to run the pprof server on.
  -keep-orphaned-files
    Keeps orphaned generated templ files. (default false)
  -v
    Set log verbosity level to "debug". (default "info")
  -log-level
    Set log verbosity level. (default "info", options: "debug", "info", "warn", "error")
  -help
    Print help and exit.

Examples:

  Generate code for all files in the current directory and subdirectories:

    templ generate

  Generate code for a single file:

    templ generate -f header.templ

  Watch the current directory and subdirectories for changes and regenerate code:

    templ generate -watch
`

func generateCmd(stdout, stderr io.Writer, args []string) (code int) {
	cmd := flag.NewFlagSet("generate", flag.ExitOnError)
	fileNameFlag := cmd.String("f", "", "")
	pathFlag := cmd.String("path", ".", "")
	toStdoutFlag := cmd.Bool("stdout", false, "")
	sourceMapVisualisationsFlag := cmd.Bool("source-map-visualisations", false, "")
	includeVersionFlag := cmd.Bool("include-version", true, "")
	includeTimestampFlag := cmd.Bool("include-timestamp", false, "")
	watchFlag := cmd.Bool("watch", false, "")
	watchPatternFlag := cmd.String("watch-pattern", "(.+\\.go$)|(.+\\.templ$)", "")
	openBrowserFlag := cmd.Bool("open-browser", true, "")
	cmdFlag := cmd.String("cmd", "", "")
	proxyFlag := cmd.String("proxy", "", "")
	proxyPortFlag := cmd.Int("proxyport", 7331, "")
	proxyBindFlag := cmd.String("proxybind", "127.0.0.1", "")
	notifyProxyFlag := cmd.Bool("notify-proxy", false, "")
	workerCountFlag := cmd.Int("w", runtime.NumCPU(), "")
	pprofPortFlag := cmd.Int("pprof", 0, "")
	keepOrphanedFilesFlag := cmd.Bool("keep-orphaned-files", false, "")
	verboseFlag := cmd.Bool("v", false, "")
	logLevelFlag := cmd.String("log-level", "info", "")
	lazyFlag := cmd.Bool("lazy", false, "")
	helpFlag := cmd.Bool("help", false, "")
	err := cmd.Parse(args)
	if err != nil {
		fmt.Fprint(stderr, generateUsageText)
		return 64 // EX_USAGE
	}
	if *helpFlag {
		fmt.Fprint(stdout, generateUsageText)
		return
	}

	log := newLogger(*logLevelFlag, *verboseFlag, stderr)

	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		fmt.Fprintln(stderr, "Stopping...")
		cancel()
	}()

	var fw generatecmd.FileWriterFunc
	if *toStdoutFlag {
		fw = generatecmd.WriterFileWriter(stdout)
	}

	err = generatecmd.Run(ctx, log, generatecmd.Arguments{
		FileName:                        *fileNameFlag,
		Path:                            *pathFlag,
		FileWriter:                      fw,
		Watch:                           *watchFlag,
		WatchPattern:                    *watchPatternFlag,
		OpenBrowser:                     *openBrowserFlag,
		Command:                         *cmdFlag,
		Proxy:                           *proxyFlag,
		ProxyPort:                       *proxyPortFlag,
		ProxyBind:                       *proxyBindFlag,
		NotifyProxy:                     *notifyProxyFlag,
		WorkerCount:                     *workerCountFlag,
		GenerateSourceMapVisualisations: *sourceMapVisualisationsFlag,
		IncludeVersion:                  *includeVersionFlag,
		IncludeTimestamp:                *includeTimestampFlag,
		PPROFPort:                       *pprofPortFlag,
		KeepOrphanedFiles:               *keepOrphanedFilesFlag,
		Lazy:                            *lazyFlag,
	})
	if err != nil {
		color.New(color.FgRed).Fprint(stderr, "(✗) ")
		fmt.Fprintln(stderr, "Command failed: "+err.Error())
		return 1
	}
	return 0
}

const fmtUsageText = `usage: templ fmt [<args> ...]

Format all files in directory:

  templ fmt .

Format stdin to stdout:

  templ fmt < header.templ

Format file or directory to stdout:

  templ fmt -stdout FILE

Args:
  -stdout
    Prints to stdout instead of in-place format
  -stdin-filepath
    Provides the formatter with filepath context when using -stdout.
    Required for organising imports.
  -v
    Set log verbosity level to "debug". (default "info")
  -log-level
    Set log verbosity level. (default "info", options: "debug", "info", "warn", "error")
  -w
    Number of workers to use when formatting code. (default runtime.NumCPUs).
  -fail
    Fails with exit code 1 if files are changed. (e.g. in CI)
  -help
    Print help and exit.
`

func fmtCmd(stdin io.Reader, stdout, stderr io.Writer, args []string) (code int) {
	cmd := flag.NewFlagSet("fmt", flag.ExitOnError)
	helpFlag := cmd.Bool("help", false, "")
	workerCountFlag := cmd.Int("w", runtime.NumCPU(), "")
	verboseFlag := cmd.Bool("v", false, "")
	logLevelFlag := cmd.String("log-level", "info", "")
	failIfChanged := cmd.Bool("fail", false, "")
	stdoutFlag := cmd.Bool("stdout", false, "")
	stdinFilepath := cmd.String("stdin-filepath", "", "")
	err := cmd.Parse(args)
	if err != nil {
		fmt.Fprint(stderr, fmtUsageText)
		return 64 // EX_USAGE
	}
	if *helpFlag {
		fmt.Fprint(stdout, fmtUsageText)
		return
	}

	log := newLogger(*logLevelFlag, *verboseFlag, stderr)

	err = fmtcmd.Run(log, stdin, stdout, fmtcmd.Arguments{
		ToStdout:      *stdoutFlag,
		Files:         cmd.Args(),
		WorkerCount:   *workerCountFlag,
		StdinFilepath: *stdinFilepath,
		FailIfChanged: *failIfChanged,
	})
	if err != nil {
		return 1
	}
	return 0
}

const lspUsageText = `usage: templ lsp [<args> ...]

Starts a language server for templ.

Args:
  -log string
    The file to log templ LSP output to, or leave empty to disable logging.
  -goplsLog string
    The file to log gopls output, or leave empty to disable logging.
  -goplsRPCTrace
    Set gopls to log input and output messages.
  -help
    Print help and exit.
  -pprof
    Enable pprof web server (default address is localhost:9999)
  -http string
    Enable http debug server by setting a listen address (e.g. localhost:7474)
`

func lspCmd(stdin io.Reader, stdout, stderr io.Writer, args []string) (code int) {
	cmd := flag.NewFlagSet("lsp", flag.ExitOnError)
	logFlag := cmd.String("log", "", "")
	goplsLog := cmd.String("goplsLog", "", "")
	goplsRPCTrace := cmd.Bool("goplsRPCTrace", false, "")
	helpFlag := cmd.Bool("help", false, "")
	pprofFlag := cmd.Bool("pprof", false, "")
	httpDebugFlag := cmd.String("http", "", "")
	err := cmd.Parse(args)
	if err != nil {
		fmt.Fprint(stderr, lspUsageText)
		return 64 // EX_USAGE
	}
	if *helpFlag {
		fmt.Fprint(stdout, lspUsageText)
		return
	}

	err = lspcmd.Run(stdin, stdout, stderr, lspcmd.Arguments{
		Log:           *logFlag,
		GoplsLog:      *goplsLog,
		GoplsRPCTrace: *goplsRPCTrace,
		PPROF:         *pprofFlag,
		HTTPDebug:     *httpDebugFlag,
	})
	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		return 1
	}
	return 0
}
