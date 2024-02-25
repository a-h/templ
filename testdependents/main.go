package main

import (
	"bufio"
	_ "embed"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"context"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/uuid"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

//go:embed dependents.csv
var dependents string

var flagAccessToken = flag.String("access-token", "", "The access token to use to access the source repository.")

func main() {
	flag.Parse()

	f, err := os.Create("log.txt")
	if err != nil {
		fmt.Printf("failed to create log file: %v\n", err)
		os.Exit(1)
	}
	logOutput := io.MultiWriter(os.Stdout, f)
	log := slog.New(slog.NewJSONHandler(logOutput, &slog.HandlerOptions{}))

	if *flagAccessToken == "" {
		log.Error("required access-token argument is missing")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Info("interrupted")
		cancel()
	}()

	r := csv.NewReader(strings.NewReader(dependents))
	records, err := r.ReadAll()
	if err != nil {
		log.Error("failed to read CSV", slog.Any("error", err))
		os.Exit(1)
	}

	// Get a list of names. The CSV fields are "Repo", "Stars", "Forks".
	names := make([]string, len(records)-1)
	for i, record := range records[1:] {
		names[i] = record[0]
	}

	repoToOK := make(map[string]bool)
	var pass, fail int

loop:
	for i, name := range names {
		url := fmt.Sprintf("https://github.com/%s.git", name)
		log.Info("cloning", slog.String("name", name), slog.Int("index", i), slog.Int("total", len(names)), slog.String("url", url))
		err := cloneRepo(*flagAccessToken, name, url)
		if err != nil {
			log.Error("failed to clone", slog.String("name", name), slog.Any("error", err))
		}
		result, err := runTests(ctx, filepath.Join("testdata", name))
		if err != nil {
			log.Error("failed to run tests", slog.String("name", name), slog.Any("error", err))
		}
		if !result.previousOK {
			log.Warn("previous container failed", slog.String("name", name), slog.String("logs", result.previousLogs))
			repoToOK[name] = false
			fail++
			continue
		}
		if !result.currentOK {
			log.Error("current container failed", slog.String("name", name), slog.String("logs", result.currentLogs))
			repoToOK[name] = false
			fail++
			continue
		}
		repoToOK[name] = true
		pass++
		log.Info("tests passed", slog.String("name", name), slog.Int("pass", pass), slog.Int("fail", fail))
		select {
		case <-ctx.Done():
			log.Info("context cancelled")
			break loop
		default:
		}
	}

	log.Info("done", slog.Int("pass", pass), slog.Int("fail", fail))
	log.Info("results", slog.Any("repoToOK", repoToOK))
}

func cloneRepo(accessToken, name, url string) error {
	dir := filepath.Join("testdata", name)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	_, err = git.PlainClone(dir, false, &git.CloneOptions{
		URL: url,
		Auth: &http.BasicAuth{
			Username: "git",
			Password: accessToken,
		},
		Progress: nil, // os.Stdout,
	})
	if err == git.ErrRepositoryAlreadyExists {
		return nil
	}
	return err
}

type testResult struct {
	previousOK   bool
	currentOK    bool
	previousLogs string
	currentLogs  string
}

func runTests(ctx context.Context, dir string) (result testResult, err error) {
	if !filepath.IsAbs(dir) {
		dir, err = filepath.Abs(dir)
		if err != nil {
			return result, fmt.Errorf("failed to get absolute path: %w", err)
		}
	}
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return result, fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer cli.Close()

	result.previousLogs, result.previousOK, err = runTestContainer(ctx, cli, containerVersionPrevious, dir)
	if err != nil {
		return result, fmt.Errorf("failed to run previous container: %w", err)
	}

	result.currentLogs, result.currentOK, err = runTestContainer(ctx, cli, containerVersionCurrent, dir)
	if err != nil {
		return result, fmt.Errorf("failed to run current container: %w", err)
	}

	return
}

type containerVersion string

const (
	containerVersionPrevious containerVersion = "previous"
	containerVersionCurrent  containerVersion = "current"
)

func runTestContainer(ctx context.Context, cli *client.Client, version containerVersion, dir string) (logs string, ok bool, err error) {
	prefix := uuid.New().String()

	previousContainer := &container.Config{
		Image:        fmt.Sprintf("templ-dependent:%s", version),
		Env:          []string{fmt.Sprintf("TEMPL_PREFIX=%s", prefix)},
		AttachStdout: true,
	}
	previousHost := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: dir,
				Target: "/app",
			},
		},
	}
	resp, err := cli.ContainerCreate(ctx, previousContainer, previousHost, nil, nil, "")
	if err != nil {
		return logs, false, fmt.Errorf("failed to create templ-dependent:%s container: %w", version, err)
	}
	defer cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})

	err = cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		return logs, false, fmt.Errorf("failed to start templ-dependent:%s container: %w", version, err)
	}

	sc, ec := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-ec:
		if err != nil {
			return logs, false, fmt.Errorf("failed to wait for templ-dependent:%s container: %w", version, err)
		}
	case <-sc:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return logs, false, fmt.Errorf("failed to get logs from templ-dependent:%s container: %w", version, err)
	}

	// Look for the OK string in the output.
	var sb strings.Builder
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), fmt.Sprintf("%s OK", prefix)) {
			ok = true
		}
		sb.WriteString(scanner.Text())
		sb.WriteString("\n")
	}

	return sb.String(), ok, nil
}
