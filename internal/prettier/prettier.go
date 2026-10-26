package prettier

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var posixCommands = map[string]string{
	"prettier":  "prettier --use-tabs --stdin-filepath $TEMPL_PRETTIER_FILENAME",
	"prettierd": "prettierd --use-tabs --stdin-filepath $TEMPL_PRETTIER_FILENAME",
}

var nuCommands = map[string]string{
	"prettier":  "prettier --use-tabs --stdin-filepath $env.TEMPL_PRETTIER_FILENAME",
	"prettierd": "prettierd --use-tabs --stdin-filepath $env.TEMPL_PRETTIER_FILENAME",
}

var shellNameToCommands = map[string]map[string]string{
	"nu": nuCommands,
}

// DefaultCommand returns the default prettier command appropriate for the current shell.
// It prefers prettier over prettierd, falling back to prettierd if prettier is not found.
func DefaultCommand() string {
	return defaultCommand(os.Getenv("SHELL"), exec.LookPath)
}

func defaultCommand(shell string, lookPath func(string) (string, error)) string {
	if shell == "" {
		shell = "/bin/sh"
	}
	shellName := filepath.Base(shell)
	commands := posixCommands
	if shellCommands, ok := shellNameToCommands[shellName]; ok {
		commands = shellCommands
	}
	for _, name := range []string{"prettier", "prettierd"} {
		if _, err := lookPath(name); err == nil {
			return commands[name]
		}
	}
	// Neither found, return the prettier command so error messages reference it.
	return commands["prettier"]
}

func IsAvailable(command string) bool {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return false
	}
	_, err := exec.LookPath(fields[0])
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
	cmd := getCommand(runtime.GOOS, os.Getenv("SHELL"), command)
	cmd.Env = append(os.Environ(), fmt.Sprintf("TEMPL_PRETTIER_FILENAME=%s", fileName))
	cmd.Stdin = strings.NewReader(input)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("unable to format with command %q, output: %q, error: %v", cmd.Args, string(output), err)
	}
	return string(output), nil
}

func getCommand(goos, shell, command string) *exec.Cmd {
	if shell == "" {
		shell = "/bin/sh"
	}
	if goos == "windows" {
		return exec.Command("cmd.exe", "/C", command)
	}
	return exec.Command(shell, "-c", command)
}
