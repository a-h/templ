package generatecmd

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"runtime"

	_ "net/http/pprof"

	"github.com/a-h/templ/cmd/templ/sloghandler"
)

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

const DefaultWatchPattern = `(.+\.go$)|(.+\.templ$)`

func NewArguments(stdout, stderr io.Writer, args []string) (cmdArgs Arguments, log *slog.Logger, help bool, err error) {
	cmd := flag.NewFlagSet("generate", flag.ContinueOnError)
	fileNameFlag := cmd.String("f", "", "")
	pathFlag := cmd.String("path", ".", "")
	toStdoutFlag := cmd.Bool("stdout", false, "")
	sourceMapVisualisationsFlag := cmd.Bool("source-map-visualisations", false, "")
	includeVersionFlag := cmd.Bool("include-version", true, "")
	includeTimestampFlag := cmd.Bool("include-timestamp", false, "")
	watchFlag := cmd.Bool("watch", false, "")
	watchPatternFlag := cmd.String("watch-pattern", DefaultWatchPattern, "")
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
	if err = cmd.Parse(args); err != nil {
		return Arguments{}, nil, false, fmt.Errorf("failed to parse arguments: %w", err)
	}

	log = sloghandler.NewLogger(*logLevelFlag, *verboseFlag, stderr)

	if *watchFlag && *fileNameFlag != "" {
		return Arguments{}, log, *helpFlag, fmt.Errorf("cannot watch a single file, remove the -f or -watch flag")
	}

	// Default to writing to files unless the stdout flag is set.
	fw := FileWriter
	if *toStdoutFlag {
		if *fileNameFlag == "" {
			return Arguments{}, log, *helpFlag, fmt.Errorf("only a single file can be output to stdout, add the -f flag to specify the file to generate code for")
		}
		fw = WriterFileWriter(stdout)
	}

	cmdArgs = Arguments{
		FileName:                        *fileNameFlag,
		FileWriter:                      fw,
		Path:                            *pathFlag,
		Watch:                           *watchFlag,
		OpenBrowser:                     *openBrowserFlag,
		Command:                         *cmdFlag,
		ProxyBind:                       *proxyBindFlag,
		ProxyPort:                       *proxyPortFlag,
		Proxy:                           *proxyFlag,
		NotifyProxy:                     *notifyProxyFlag,
		WorkerCount:                     *workerCountFlag,
		GenerateSourceMapVisualisations: *sourceMapVisualisationsFlag,
		IncludeVersion:                  *includeVersionFlag,
		IncludeTimestamp:                *includeTimestampFlag,
		PPROFPort:                       *pprofPortFlag,
		KeepOrphanedFiles:               *keepOrphanedFilesFlag,
		Lazy:                            *lazyFlag,
	}
	cmdArgs.WatchPattern, err = regexp.Compile(*watchPatternFlag)
	if err != nil {
		return cmdArgs, log, *helpFlag, fmt.Errorf("invalid watch pattern %q: %w", *watchPatternFlag, err)
	}

	return cmdArgs, log, *helpFlag, nil
}

type Arguments struct {
	FileName                        string
	FileWriter                      FileWriterFunc
	Path                            string
	Watch                           bool
	WatchPattern                    *regexp.Regexp
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

type ArgumentError struct {
	Message string
}

func (e *ArgumentError) Error() string {
	return e.Message
}

func (a *ArgumentError) Code() int {
	return 64 // EX_USAGE
}

func Run(ctx context.Context, stdout, stderr io.Writer, args []string) (err error) {
	cmdArgs, log, help, err := NewArguments(stdout, stderr, args)
	if err != nil {
		_, _ = fmt.Fprint(stderr, generateUsageText)
		return &ArgumentError{
			Message: err.Error(),
		}
	}
	if help {
		_, _ = fmt.Fprint(stdout, generateUsageText)
		return nil
	}
	g, err := NewGenerate(log, cmdArgs)
	if err != nil {
		return err
	}
	return g.Run(ctx)
}
