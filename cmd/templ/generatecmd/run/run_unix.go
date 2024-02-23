//go:build unix

package run

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
)

var m = &sync.Mutex{}
var running = map[string]*exec.Cmd{}

func KillAll() (err error) {
	m.Lock()
	defer m.Unlock()
	for _, cmd := range running {
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if err != nil {
			return err
		}
	}
	running = map[string]*exec.Cmd{}
	return
}

func Stop(cmd *exec.Cmd) (err error) {
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}

func Run(ctx context.Context, workingDir, input string) (cmd *exec.Cmd, err error) {
	m.Lock()
	defer m.Unlock()
	cmd, ok := running[input]
	if ok {
		if err = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
			return cmd, err
		}
		delete(running, input)
	}
	parts := strings.Fields(input)
	executable := parts[0]
	args := []string{}
	if len(parts) > 1 {
		args = append(args, parts[1:]...)
	}

	cmd = exec.Command(executable, args...)
	cmd.Env = os.Environ()
	cmd.Dir = workingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	running[input] = cmd
	err = cmd.Start()
	return
}
