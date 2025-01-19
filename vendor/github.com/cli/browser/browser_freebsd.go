package browser

import (
	"errors"
	"fmt"
	"os/exec"
)

func openBrowser(url string) error {
	err := runCmd("xdg-open", url)
	if errors.Is(err, exec.ErrNotFound) {
		return fmt.Errorf("%w - install xdg-utils from ports(8)", err)
	}
	return err
}
