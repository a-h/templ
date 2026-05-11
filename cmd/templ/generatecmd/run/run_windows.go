//go:build windows

package run

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"sync"
)

var (
	m       = &sync.Mutex{}
	running = map[string]*exec.Cmd{}
)

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
	shell := os.Getenv("COMSPEC")
	if shell == "" {
		shell = "cmd.exe"
	}
	cmd = exec.Command(shell, "/C", input)
	cmd.Env = os.Environ()
	cmd.Dir = workingDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return cmd, err
	}
	running[input] = cmd
	return cmd, nil
}
