package prettier

import (
	"bytes"
	"errors"
	"os/exec"
)

func Run(input string, fileName string) (formatted string, err error) {
	path, err := exec.LookPath("prettier")
	if err != nil {
		return "", errors.New("unable to format script or CSS: prettier not found in PATH, please install it")
	}
	cmd := exec.Command(path, "--use-tabs", "true", "--stdin-filepath", fileName)
	cmd.Stdin = bytes.NewBufferString(input)
	var out bytes.Buffer
	cmd.Stdout = &out
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return "", errors.New(stderr.String())
		}
		return "", err
	}
	return out.String(), nil
}
