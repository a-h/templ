package prettier

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const DefaultCommand = "prettier --use-tabs --stdin-filepath $TEMPL_PRETTIER_FILENAME"

func IsAvailable(command string) bool {
	executable := strings.Fields(command)[0]
	_, err := exec.LookPath(executable)
	return err == nil
}

// Run the prettier command with the given input and file name.
// $TEMPL_PRETTIER_FILENAME is set to the file name being formatted.
// To format blocks inside templ files a fake name is provided, e.g. format.html, format.js, format.css etc.
// The command is run in a shell, so it can be a complex command with pipes and redirections.
//
// Examples:
//
//	prettier --use-tabs --stdin-filepath $TEMPL_PRETTIER_FILENAME
//	prettierd --use-tabs --stdin-filepath $TEMPL_PRETTIER_FILENAME
//	npx prettier --use-tabs --stdin-filepath $TEMPL_PRETTIER_FILENAME
//	prettier --config ./frontend/.prettierrc --use-tabs --stdin-filepath $TEMPL_PRETTIER_FILENAME
func Run(input, fileName, command string) (formatted string, err error) {
	cmd := getCommand(runtime.GOOS, command)
	cmd.Env = append(os.Environ(), fmt.Sprintf("TEMPL_PRETTIER_FILENAME=%s", fileName))
	cmd.Stdin = strings.NewReader(input)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("unable to format with command %q, output: %q, error: %v", cmd.Args, string(output), err)
	}
	return string(output), nil
}

func getCommand(goos, command string) *exec.Cmd {
	if goos == "windows" {
		return exec.Command("cmd.exe", "/C", command)
	}
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	return exec.Command(shell, "-c", command)
}
