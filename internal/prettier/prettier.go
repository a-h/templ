package prettier

import (
	"fmt"
	"html"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/a-h/templ/internal/htmlfind"
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

func Element(name string, typeAttrValue string, content string, depth int, prettierCommand string) (after string, err error) {
	var indentationWrapper strings.Builder

	// Add divs to the start and end of the script to ensure that prettier formats the content with
	// correct indentation.
	for i := range depth {
		indentationWrapper.WriteString(fmt.Sprintf("<div data-templ-depth=\"%d\">", i))
	}

	// Write start tag with type attribute if present.
	indentationWrapper.WriteString("<")
	indentationWrapper.WriteString(name)
	if typeAttrValue != "" {
		indentationWrapper.WriteString(" type=\"")
		indentationWrapper.WriteString(html.EscapeString(typeAttrValue))
		indentationWrapper.WriteString("\"")
	}
	indentationWrapper.WriteString(">")

	// Write contents.
	indentationWrapper.WriteString(content)

	// Write end tag.
	indentationWrapper.WriteString("</")
	indentationWrapper.WriteString(name)
	indentationWrapper.WriteString(">")

	for range depth {
		indentationWrapper.WriteString("</div>")
	}

	before := indentationWrapper.String()
	after, err = Run(before, "templ_content.html", prettierCommand)
	if err != nil {
		return "", fmt.Errorf("prettier error: %w", err)
	}
	if before == after {
		return before, nil
	}

	// Chop off the start and end divs we added to get prettier to format the content with correct
	// indentation.
	matcher := htmlfind.Element(name)
	nodes, err := htmlfind.AllReader(strings.NewReader(after), matcher)
	if err != nil {
		return before, fmt.Errorf("htmlfind error: %w", err)
	}
	if len(nodes) != 1 {
		return before, fmt.Errorf("expected 1 %q node, got %d", name, len(nodes))
	}
	scriptNode := nodes[0]
	if scriptNode.FirstChild == nil {
		return before, fmt.Errorf("%q node has no children", name)
	}
	var sb strings.Builder
	for node := range scriptNode.ChildNodes() {
		sb.WriteString(node.Data)
	}
	after = strings.TrimRight(sb.String(), " \t\r\n") + "\n" + strings.Repeat("\t", depth)

	return after, nil
}
