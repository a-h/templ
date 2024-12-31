//go:build windows

package run

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

var m = &sync.Mutex{}
var running = map[string]*exec.Cmd{}

func KillAll() (err error) {
	m.Lock()
	defer m.Unlock()
	for _, cmd := range running {
		kill := exec.Command("TASKKILL", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
		kill.Stderr = os.Stderr
		kill.Stdout = os.Stdout
		err := kill.Run()
		if err != nil {
			return err
		}
	}
	running = map[string]*exec.Cmd{}
	return
}

func Stop(cmd *exec.Cmd) (err error) {
	kill := exec.Command("TASKKILL", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
	kill.Stderr = os.Stderr
	kill.Stdout = os.Stdout
	return kill.Run()
}

func Run(ctx context.Context, workingDir string, input string) (cmd *exec.Cmd, err error) {
	m.Lock()
	defer m.Unlock()
	cmd, ok := running[input]
	if ok {
		kill := exec.Command("TASKKILL", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
		kill.Stderr = os.Stderr
		kill.Stdout = os.Stdout
		err := kill.Run()
		if err != nil {
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
	running[input] = cmd
	err = cmd.Start()
	return
}
