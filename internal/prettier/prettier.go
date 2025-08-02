package prettier

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

var resolveFormatter = sync.OnceFunc(func() {
	candidates := [][]string{
		{"prettierd"},
		{"prettier"},
		{"npx", "prettier"},
	}
	for _, cmd := range candidates {
		if path, err := exec.LookPath(cmd[0]); err == nil {
			formatCmd = append([]string{path}, cmd[1:]...)
			return
		}
	}
	lookupErr = errors.New("unable to find a formatter: tried prettierd, prettier, and npx prettier")
})

var (
	formatCmd []string
	lookupErr error
)

func Run(input string, fileName string) (formatted string, err error) {
	resolveFormatter()
	if lookupErr != nil {
		return "", lookupErr
	}

	cmd := exec.Command(formatCmd[0], append(formatCmd[1:], "--use-tabs", "--stdin-filepath", fileName)...)
	cmd.Stdin = strings.NewReader(input)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("unable to format with command %q, output: %q, error: %v", cmd.Args, string(output), err)
	}
	return string(output), nil
}
