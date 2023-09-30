package run

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

var m = &sync.Mutex{}
var running = map[string]*exec.Cmd{}

func KillAll() {
	m.Lock()
	defer m.Unlock()
	for _, cmd := range running {
		if runtime.GOOS == "windows" {
			kill := exec.Command("TASKKILL", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
			kill.Stderr = os.Stderr
			kill.Stdout = os.Stdout
			kill.Run()
			continue
		}
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	running = map[string]*exec.Cmd{}
}

func Run(ctx context.Context, workingDir, input string) (cmd *exec.Cmd, err error) {
	m.Lock()
	defer m.Unlock()
	cmd, ok := running[input]
	if ok {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			KillAll()
		}()
		wg.Wait()
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
