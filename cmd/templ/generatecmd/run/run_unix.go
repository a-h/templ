//go:build unix

package run

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

var (
	m       = &sync.Mutex{}
	running = map[string]*exec.Cmd{}
)

func KillAll() (err error) {
	m.Lock()
	defer m.Unlock()
	var errs []error
	for _, cmd := range running {
		if err := kill(cmd); err != nil {
			errs = append(errs, fmt.Errorf("failed to kill process %d: %w", cmd.Process.Pid, err))
		}
	}
	running = map[string]*exec.Cmd{}
	return errors.Join(errs...)
}

func kill(cmd *exec.Cmd) (err error) {
	pgid := -cmd.Process.Pid
	errs := make([]error, 4)
	errs[0] = ignoreExited(syscall.Kill(pgid, syscall.SIGINT))
	errs[1] = ignoreExited(syscall.Kill(pgid, syscall.SIGTERM))
	errs[2] = ignoreExited(cmd.Wait())
	errs[3] = ignoreExited(syscall.Kill(pgid, syscall.SIGKILL))
	return errors.Join(errs...)
}

func ignoreExited(err error) error {
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	if errors.Is(err, syscall.EPERM) {
		return nil
	}
	if _, ok := err.(*exec.ExitError); ok {
		return nil
	}
	return err
}

func Run(ctx context.Context, workingDir string, input string) (cmd *exec.Cmd, err error) {
	m.Lock()
	defer m.Unlock()
	cmd, ok := running[input]
	if ok {
		if err := kill(cmd); err != nil {
			return cmd, fmt.Errorf("failed to kill process %d: %w", cmd.Process.Pid, err)
		}

		delete(running, input)
	}
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	cmd = exec.CommandContext(ctx, shell, "-c", input)
	// Wait for the process to finish gracefully before termination.
	cmd.WaitDelay = time.Second * 3
	cmd.Env = os.Environ()
	cmd.Dir = workingDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		return cmd, err
	}
	running[input] = cmd
	return cmd, nil
}
