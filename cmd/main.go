package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/a-h/templ/cmd/compile"
	"github.com/a-h/templ/cmd/lsp"
)

var Version = ""

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "compile":
		compileCmd(os.Args[2:])
		return
	case "lsp":
		lspCmd(os.Args[2:])
		return
	case "version":
		fmt.Println(Version)
		return
	case "--version":
		fmt.Println(Version)
		return
	}
	usage()
}

func usage() {
	fmt.Println(`usage: templ <command> [parameters]
To see help text, you can run:
  templ compile --help
  templ lsp --help
  templ version
examples:
  templ compile`)
	os.Exit(1)
}

func compileCmd(args []string) {
	cmd := flag.NewFlagSet("compile", flag.ExitOnError)
	helpFlag := cmd.Bool("help", false, "Print help and exit.")
	err := cmd.Parse(args)
	if err != nil || *helpFlag {
		cmd.PrintDefaults()
		return
	}
	err = compile.Run(args)
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
	err = lsp.Run(lsp.Arguments{
		Log:           *log,
		GoplsLog:      *goplsLog,
		GoplsRPCTrace: *goplsRPCTrace,
	})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
