package generatecmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	templruntime "github.com/a-h/templ/runtime"

	"github.com/a-h/templ"
	"github.com/a-h/templ/cmd/templ/generatecmd/modcheck"
	"github.com/a-h/templ/cmd/templ/generatecmd/proxy"
	"github.com/a-h/templ/cmd/templ/generatecmd/run"
	"github.com/a-h/templ/cmd/templ/generatecmd/watcher"
	"github.com/a-h/templ/generator"
	"github.com/cenkalti/backoff/v4"
	"github.com/cli/browser"
	"github.com/fsnotify/fsnotify"
)

const defaultWatchPattern = `(.+\.go$)|(.+\.templ$)`

func NewGenerate(log *slog.Logger, args Arguments) (g *Generate, err error) {
	g = &Generate{
		Log:  log,
		Args: &args,
	}
	if g.Args.WorkerCount == 0 {
		g.Args.WorkerCount = runtime.NumCPU()
	}
	if g.Args.WatchPattern == "" {
		g.Args.WatchPattern = defaultWatchPattern
	}
	g.WatchPattern, err = regexp.Compile(g.Args.WatchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to compile watch pattern %q: %w", g.Args.WatchPattern, err)
	}
	return g, nil
}

type Generate struct {
	Log          *slog.Logger
	Args         *Arguments
	WatchPattern *regexp.Regexp
}

type GenerationEvent struct {
	Event       fsnotify.Event
	Updated     bool
	GoUpdated   bool
	TextUpdated bool
}

func (cmd Generate) Run(ctx context.Context) (err error) {
	if cmd.Args.NotifyProxy {
		return proxy.NotifyProxy(cmd.Args.ProxyBind, cmd.Args.ProxyPort)
	}
	if cmd.Args.Watch && cmd.Args.FileName != "" {
		return fmt.Errorf("cannot watch a single file, remove the -f or -watch flag")
	}
	writingToWriter := cmd.Args.FileWriter != nil
	if cmd.Args.FileName == "" && writingToWriter {
		return fmt.Errorf("only a single file can be output to stdout, add the -f flag to specify the file to generate code for")
	}
	// Default to writing to files.
	if cmd.Args.FileWriter == nil {
		cmd.Args.FileWriter = FileWriter
	}
	if cmd.Args.PPROFPort > 0 {
		go func() {
			_ = http.ListenAndServe(fmt.Sprintf("localhost:%d", cmd.Args.PPROFPort), nil)
		}()
	}

	// Use absolute path.
	if !path.IsAbs(cmd.Args.Path) {
		cmd.Args.Path, err = filepath.Abs(cmd.Args.Path)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}
	}

	// Configure generator.
	var opts []generator.GenerateOpt
	if cmd.Args.IncludeVersion {
		opts = append(opts, generator.WithVersion(templ.Version()))
	}
	if cmd.Args.IncludeTimestamp {
		opts = append(opts, generator.WithTimestamp(time.Now()))
	}

	// Check the version of the templ module.
	if err := modcheck.Check(cmd.Args.Path); err != nil {
		cmd.Log.Warn("templ version check: " + err.Error())
	}

	cmd.Log.Debug("Creating filesystem event handler")
	fseh := NewFSEventHandler(
		cmd.Log,
		cmd.Args.Path,
		cmd.Args.Watch,
		opts,
		cmd.Args.GenerateSourceMapVisualisations,
		cmd.Args.KeepOrphanedFiles,
		cmd.Args.FileWriter,
		cmd.Args.Lazy,
	)

	// If we're processing a single file, don't bother setting up the channels/multithreaing.
	if cmd.Args.FileName != "" {
		_, err = fseh.HandleEvent(ctx, fsnotify.Event{
			Name: cmd.Args.FileName,
			Op:   fsnotify.Create,
		})
		return err
	}

	// Start timer.
	start := time.Now()

	// Create channels:
	// For the initial filesystem walk and subsequent (optional) fsnotify events.
	events := make(chan fsnotify.Event)
	// Count of events currently being processed by the event handler.
	var eventsWG sync.WaitGroup
	// Used to check that the event handler has completed.
	var eventHandlerWG sync.WaitGroup
	// For errs from the watcher.
	errs := make(chan error)
	// Tracks whether errors occurred during the generation process.
	var errorCount atomic.Int64
	// For triggering actions after generation has completed.
	postGeneration := make(chan *GenerationEvent, 256)
	// Used to check that the post-generation handler has completed.
	var postGenerationWG sync.WaitGroup
	var postGenerationEventsWG sync.WaitGroup

	// Waitgroup for the push process.
	var pushHandlerWG sync.WaitGroup

	// Start process to push events into the channel.
	pushHandlerWG.Add(1)
	go func() {
		defer pushHandlerWG.Done()
		defer close(events)
		cmd.Log.Debug(
			"Walking directory",
			slog.String("path", cmd.Args.Path),
			slog.Bool("devMode", cmd.Args.Watch),
		)
		if err := watcher.WalkFiles(ctx, cmd.Args.Path, cmd.WatchPattern, events); err != nil {
			cmd.Log.Error("WalkFiles failed, exiting", slog.Any("error", err))
			errs <- FatalError{Err: fmt.Errorf("failed to walk files: %w", err)}
			return
		}
		if !cmd.Args.Watch {
			cmd.Log.Debug("Dev mode not enabled, process can finish early")
			return
		}
		cmd.Log.Info("Watching files")
		rw, err := watcher.Recursive(ctx, cmd.Args.Path, cmd.WatchPattern, events, errs)
		if err != nil {
			cmd.Log.Error("Recursive watcher setup failed, exiting", slog.Any("error", err))
			errs <- FatalError{Err: fmt.Errorf("failed to setup recursive watcher: %w", err)}
			return
		}
		cmd.Log.Debug("Waiting for context to be cancelled to stop watching files")
		<-ctx.Done()
		cmd.Log.Debug("Context cancelled, closing watcher")
		if err := rw.Close(); err != nil {
			cmd.Log.Error("Failed to close watcher", slog.Any("error", err))
		}
		cmd.Log.Debug("Waiting for events to be processed")
		eventsWG.Wait()
		cmd.Log.Debug(
			"All pending events processed, waiting for pending post-generation events to complete",
		)
		postGenerationEventsWG.Wait()
		cmd.Log.Debug(
			"All post-generation events processed, deleting watch mode text files",
			slog.Int64("errorCount", errorCount.Load()),
		)
		fileEvents := make(chan fsnotify.Event)
		go func() {
			if err := watcher.WalkFiles(ctx, cmd.Args.Path, cmd.WatchPattern, fileEvents); err != nil {
				cmd.Log.Error("Post dev mode WalkFiles failed", slog.Any("error", err))
				errs <- FatalError{Err: fmt.Errorf("failed to walk files: %w", err)}
				return
			}
			close(fileEvents)
		}()
		for event := range fileEvents {
			if strings.HasSuffix(event.Name, "_templ.go") || strings.HasSuffix(event.Name, ".templ") {
				watchModeFileName := templruntime.GetDevModeTextFileName(event.Name)
				if err := os.Remove(watchModeFileName); err != nil && !errors.Is(err, os.ErrNotExist) {
					cmd.Log.Warn("Failed to remove watch mode text file", slog.Any("error", err))
				}
			}
		}
	}()

	// Start process to handle events.
	eventHandlerWG.Add(1)
	sem := make(chan struct{}, cmd.Args.WorkerCount)
	go func() {
		defer eventHandlerWG.Done()
		defer close(postGeneration)
		cmd.Log.Debug("Starting event handler")
		for event := range events {
			eventsWG.Add(1)
			sem <- struct{}{}
			go func(event fsnotify.Event) {
				cmd.Log.Debug("Processing file", slog.String("file", event.Name))
				defer eventsWG.Done()
				defer func() { <-sem }()
				r, err := fseh.HandleEvent(ctx, event)
				if err != nil {
					errs <- err
				}
				if !(r.GoUpdated || r.TextUpdated) {
					cmd.Log.Debug("File not updated", slog.String("file", event.Name))
					return
				}
				e := &GenerationEvent{
					Event:       event,
					Updated:     r.Updated,
					GoUpdated:   r.GoUpdated,
					TextUpdated: r.TextUpdated,
				}
				cmd.Log.Debug("File updated", slog.String("file", event.Name))
				postGeneration <- e
			}(event)
		}
		// Wait for all events to be processed before closing.
		eventsWG.Wait()
	}()

	// Start process to handle post-generation events.
	var updates int
	postGenerationWG.Add(1)
	var firstPostGenerationExecuted bool
	go func() {
		defer close(errs)
		defer postGenerationWG.Done()
		cmd.Log.Debug("Starting post-generation handler")
		timeout := time.NewTimer(time.Hour * 24 * 365)
		var goUpdated, textUpdated bool
		var p *proxy.Handler
		for {
			select {
			case ge := <-postGeneration:
				if ge == nil {
					cmd.Log.Debug("Post-generation event channel closed, exiting")
					return
				}
				goUpdated = goUpdated || ge.GoUpdated
				textUpdated = textUpdated || ge.TextUpdated
				if goUpdated || textUpdated {
					updates++
				}
				// Reset timer.
				if !timeout.Stop() {
					<-timeout.C
				}
				timeout.Reset(time.Millisecond * 100)
			case <-timeout.C:
				if !goUpdated && !textUpdated {
					// Nothing to process, reset timer and wait again.
					timeout.Reset(time.Hour * 24 * 365)
					break
				}
				postGenerationEventsWG.Add(1)
				if cmd.Args.Command != "" && goUpdated {
					cmd.Log.Debug("Executing command", slog.String("command", cmd.Args.Command))
					if cmd.Args.Watch {
						os.Setenv("TEMPL_DEV_MODE", "true")
					}
					if _, err := run.Run(ctx, cmd.Args.Path, cmd.Args.Command); err != nil {
						cmd.Log.Error("Error executing command", slog.Any("error", err))
					}
				}
				if !firstPostGenerationExecuted {
					cmd.Log.Debug("First post-generation event received, starting proxy")
					firstPostGenerationExecuted = true
					p, err = cmd.StartProxy(ctx)
					if err != nil {
						cmd.Log.Error("Failed to start proxy", slog.Any("error", err))
					}
				}
				// Send server-sent event.
				if p != nil && (textUpdated || goUpdated) {
					cmd.Log.Debug("Sending reload event")
					p.SendSSE("message", "reload")
				}
				postGenerationEventsWG.Done()
				// Reset timer.
				timeout.Reset(time.Millisecond * 100)
				textUpdated = false
				goUpdated = false
			}
		}
	}()

	// Read errors.
	for err := range errs {
		if err == nil {
			continue
		}
		if errors.Is(err, FatalError{}) {
			cmd.Log.Debug("Fatal error, exiting")
			return err
		}
		cmd.Log.Error("Error", slog.Any("error", err))
		errorCount.Add(1)
	}

	// Wait for everything to complete.
	cmd.Log.Debug("Waiting for push handler to complete")
	pushHandlerWG.Wait()
	cmd.Log.Debug("Waiting for event handler to complete")
	eventHandlerWG.Wait()
	cmd.Log.Debug("Waiting for post-generation handler to complete")
	postGenerationWG.Wait()
	if cmd.Args.Command != "" {
		cmd.Log.Debug("Killing command", slog.String("command", cmd.Args.Command))
		if err := run.KillAll(); err != nil {
			cmd.Log.Error("Error killing command", slog.Any("error", err))
		}
	}

	// Check for errors after everything has completed.
	if errorCount.Load() > 0 {
		return fmt.Errorf("generation completed with %d errors", errorCount.Load())
	}

	cmd.Log.Info(
		"Complete",
		slog.Int("updates", updates),
		slog.Duration("duration", time.Since(start)),
	)
	return nil
}

func (cmd *Generate) StartProxy(ctx context.Context) (p *proxy.Handler, err error) {
	if cmd.Args.Proxy == "" {
		cmd.Log.Debug("No proxy URL specified, not starting proxy")
		return nil, nil
	}
	var target *url.URL
	target, err = url.Parse(cmd.Args.Proxy)
	if err != nil {
		return nil, FatalError{Err: fmt.Errorf("failed to parse proxy URL: %w", err)}
	}
	if cmd.Args.ProxyPort == 0 {
		cmd.Args.ProxyPort = 7331
	}
	if cmd.Args.ProxyBind == "" {
		cmd.Args.ProxyBind = "127.0.0.1"
	}
	p = proxy.New(cmd.Log, cmd.Args.ProxyBind, cmd.Args.ProxyPort, target)
	go func() {
		cmd.Log.Info("Proxying", slog.String("from", p.URL), slog.String("to", p.Target.String()))
		if err := http.ListenAndServe(fmt.Sprintf("%s:%d", cmd.Args.ProxyBind, cmd.Args.ProxyPort), p); err != nil {
			cmd.Log.Error("Proxy failed", slog.Any("error", err))
		}
	}()
	if !cmd.Args.OpenBrowser {
		cmd.Log.Debug("Not opening browser")
		return p, nil
	}
	go func() {
		cmd.Log.Debug("Waiting for proxy to be ready", slog.String("url", p.URL))
		backoff := backoff.NewExponentialBackOff()
		backoff.InitialInterval = time.Second
		var client http.Client
		client.Timeout = 1 * time.Second
		for {
			if _, err := client.Get(p.URL); err == nil {
				break
			}
			d := backoff.NextBackOff()
			cmd.Log.Debug(
				"Proxy not ready, retrying",
				slog.String("url", p.URL),
				slog.Any("backoff", d),
			)
			time.Sleep(d)
		}
		if err := browser.OpenURL(p.URL); err != nil {
			cmd.Log.Error("Failed to open browser", slog.Any("error", err))
		}
	}()
	return p, nil
}
