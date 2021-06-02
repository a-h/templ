package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/a-h/templ/cmd/templ/fmtcmd"
	"github.com/a-h/templ/cmd/templ/generatecmd"
	"github.com/a-h/templ/cmd/templ/lspcmd"
)

func version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	return info.Main.Version
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
	case "fmt":
		fmtCmd(os.Args[2:])
		return
	case "lsp":
		lspCmd(os.Args[2:])
		return
	case "version":
		fmt.Println(version())
		return
	case "--version":
		fmt.Println(version())
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
  templ version
examples:
  templ compile`)
	os.Exit(1)
}

func generateCmd(args []string) {
	cmd := flag.NewFlagSet("generate", flag.ExitOnError)
	helpFlag := cmd.Bool("help", false, "Print help and exit.")
	err := cmd.Parse(args)
	if err != nil || *helpFlag {
		cmd.PrintDefaults()
		return
	}
	err = generatecmd.Run(args)
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
	err := cmd.Parse(args)
	if err != nil || *helpFlag {
		cmd.PrintDefaults()
		return
	}
	err = lspcmd.Run(lspcmd.Arguments{
		Log:           *log,
		GoplsLog:      *goplsLog,
		GoplsRPCTrace: *goplsRPCTrace,
	})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
