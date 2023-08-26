package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/a-h/templ"
	"github.com/a-h/templ/cmd/templ/fmtcmd"
	"github.com/a-h/templ/cmd/templ/generatecmd"
	"github.com/a-h/templ/cmd/templ/lspcmd"
	"github.com/a-h/templ/cmd/templ/migratecmd"
)

// Source builds use this value. When installed using `go install github.com/a-h/templ/cmd/templ@latest` the `version` variable is empty, but
// the debug.ReadBuildInfo return value provides the package version number installed by `go install`
func goInstallVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	return info.Main.Version
}

func getVersion() string {
	if templ.Version != "" {
		return templ.Version
	}
	return goInstallVersion()
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "generate":
		generateCmd(os.Args[2:])
		return
	case "migrate":
		migrateCmd(os.Args[2:])
		return
	case "fmt":
		fmtCmd(os.Args[2:])
		return
	case "lsp":
		lspCmd(os.Args[2:])
		return
	case "version":
		fmt.Println(getVersion())
		return
	case "--version":
		fmt.Println(getVersion())
		return
	}
	usage()
}

func usage() {
	fmt.Println(`usage: templ <command> [parameters]
To see help text, you can run:
  templ generate --help
  templ fmt --help
  templ lsp --help
  templ migrate --help
  templ version
examples:
  templ generate`)
	os.Exit(1)
}

func generateCmd(args []string) {
	cmd := flag.NewFlagSet("generate", flag.ExitOnError)
	fileNameFlag := cmd.String("f", "", "Optionally generates code for a single file, e.g. -f header.templ")
	pathFlag := cmd.String("path", ".", "Generates code for all files in path.")
	sourceMapVisualisations := cmd.Bool("sourceMapVisualisations", false, "Set to true to generate HTML files to visualise the templ code and its corresponding Go code.")
	watchFlag := cmd.Bool("watch", false, "Set to true to watch the path for changes and regenerate code.")
	cmdFlag := cmd.String("cmd", "", "Set the command to run after generating code.")
	proxyFlag := cmd.String("proxy", "", "Set the URL to proxy after generating code and executing the command.")
	proxyPortFlag := cmd.Int("proxyport", 7331, "The port the proxy will listen on.")
	workerCountFlag := cmd.Int("w", runtime.NumCPU(), "Number of workers to run in parallel.")
	pprofPortFlag := cmd.Int("pprof", 0, "Port to start pprof web server on.")
	helpFlag := cmd.Bool("help", false, "Print help and exit.")
	err := cmd.Parse(args)
	if err != nil || *helpFlag {
		cmd.PrintDefaults()
		return
	}
	err = generatecmd.Run(generatecmd.Arguments{
		FileName:                        *fileNameFlag,
		Path:                            *pathFlag,
		Watch:                           *watchFlag,
		Command:                         *cmdFlag,
		Proxy:                           *proxyFlag,
		ProxyPort:                       *proxyPortFlag,
		WorkerCount:                     *workerCountFlag,
		GenerateSourceMapVisualisations: *sourceMapVisualisations,
		PPROFPort:                       *pprofPortFlag,
	})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func migrateCmd(args []string) {
	cmd := flag.NewFlagSet("migrate", flag.ExitOnError)
	fileName := cmd.String("f", "", "Optionally migrate a single file, e.g. -f header.templ")
	path := cmd.String("path", ".", "Migrates code for all files in path.")
	helpFlag := cmd.Bool("help", false, "Print help and exit.")
	err := cmd.Parse(args)
	if err != nil || *helpFlag {
		cmd.PrintDefaults()
		return
	}
	err = migratecmd.Run(migratecmd.Arguments{
		FileName: *fileName,
		Path:     *path,
	})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func fmtCmd(args []string) {
	cmd := flag.NewFlagSet("fmt", flag.ExitOnError)
	helpFlag := cmd.Bool("help", false, "Print help and exit.")
	err := cmd.Parse(args)
	if err != nil || *helpFlag {
		cmd.PrintDefaults()
		return
	}
	err = fmtcmd.Run(args)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func lspCmd(args []string) {
	cmd := flag.NewFlagSet("lsp", flag.ExitOnError)
	log := cmd.String("log", "", "The file to log templ LSP output to, or leave empty to disable logging.")
	goplsLog := cmd.String("goplsLog", "", "The file to log gopls output, or leave empty to disable logging.")
	goplsRPCTrace := cmd.Bool("goplsRPCTrace", false, "Set gopls to log input and output messages.")
	helpFlag := cmd.Bool("help", false, "Print help and exit.")
	pprofFlag := cmd.Bool("pprof", false, "Enable pprof web server (default address is localhost:9999)")
	httpDebugFlag := cmd.String("http", "", "Enable http debug server by setting a listen address (e.g. localhost:7474)")
	err := cmd.Parse(args)
	if err != nil || *helpFlag {
		cmd.PrintDefaults()
		return
	}
	err = lspcmd.Run(lspcmd.Arguments{
		Log:           *log,
		GoplsLog:      *goplsLog,
		GoplsRPCTrace: *goplsRPCTrace,
		PPROF:         *pprofFlag,
		HTTPDebug:     *httpDebugFlag,
	})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
