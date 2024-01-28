package generatecmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"sync"
	"time"

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

func NewGenerate(log *slog.Logger, args Arguments) *Generate {
	return &Generate{
		Log:  log,
		Args: &args,
	}
}

type Generate struct {
	Log  *slog.Logger
	Args *Arguments
}

type GenerationEvent struct {
	Event       fsnotify.Event
	GoUpdated   bool
	TextUpdated bool
}

func (cmd Generate) Run(ctx context.Context) (err error) {
	if cmd.Args.Watch && cmd.Args.FileName != "" {
		return fmt.Errorf("cannot watch a single file, remove the -f or -watch flag")
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

	fseh := NewFSEventHandler(cmd.Log, cmd.Args.Path, cmd.Args.Watch, opts, cmd.Args.GenerateSourceMapVisualisations, cmd.Args.KeepOrphanedFiles)

	// If we're processing a single file, don't bother setting up the channels/multithreaing.
	if cmd.Args.FileName != "" {
		_, _, err = fseh.HandleEvent(ctx, fsnotify.Event{
			Name: cmd.Args.FileName,
			Op:   fsnotify.Create,
		})
		return err
	}

	// Create channels:
	// For the initial filesystem walk and subsequent (optional) fsnotify events.
	events := make(chan fsnotify.Event)
	// count of events currently being processed by the event handler.
	var eventsWG sync.WaitGroup
	// Used to check that the event handler has completed.
	var eventHandlerWG sync.WaitGroup
	// For errs from the watcher.
	errs := make(chan error)
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
		defer close(errs)
		cmd.Log.Debug("Walking directory", slog.String("path", cmd.Args.Path), slog.Bool("devMode", cmd.Args.Watch))
		err = watcher.WalkFiles(ctx, cmd.Args.Path, events)
		if err != nil {
			cmd.Log.Error("WalkFiles failed, exiting", slog.Any("error", err))
			errs <- FatalError{Err: fmt.Errorf("failed to walk files: %w", err)}
			return
		}
		if !cmd.Args.Watch {
			cmd.Log.Debug("Dev mode not enabled, process can finish early")
			return
		}
		cmd.Log.Info("Watching files")
		rw, err := watcher.Recursive(ctx, cmd.Args.Path, events, errs)
		if err != nil {
			cmd.Log.Error("Recursive watcher setup failed, exiting", slog.Any("error", err))
			errs <- FatalError{Err: fmt.Errorf("failed to setup recursive watcher: %w", err)}
			return
		}
		cmd.Log.Debug("Waiting for context to be cancelled to stop watching files")
		<-ctx.Done()
		cmd.Log.Debug("Context cancelled, closing watcher")
		if err = rw.Close(); err != nil {
			cmd.Log.Error("Failed to close watcher", slog.Any("error", err))
			err = nil
		}
		cmd.Log.Debug("Waiting for events to be processed")
		eventsWG.Wait()
		cmd.Log.Debug("All pending events processed, waiting for pending post-generation events to complete")
		postGenerationEventsWG.Wait()
		cmd.Log.Debug("All post-generation events processed, running walk again, but in production mode")
		fseh.DevMode = false
		err = watcher.WalkFiles(ctx, cmd.Args.Path, events)
		if err != nil {
			cmd.Log.Error("Post dev mode WalkFiles failed", slog.Any("error", err))
			errs <- FatalError{Err: fmt.Errorf("failed to walk files: %w", err)}
			return
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
				goUpdated, textUpdated, err := fseh.HandleEvent(ctx, event)
				if err != nil {
					cmd.Log.Error("Event handler failed", slog.Any("error", err))
					errs <- err
				}
				if goUpdated || textUpdated {
					postGeneration <- &GenerationEvent{
						Event:       event,
						GoUpdated:   goUpdated,
						TextUpdated: textUpdated,
					}
				}
			}(event)
		}
		// Wait for all events to be processed before closing.
		eventsWG.Wait()
	}()

	// Start process to handle post-generation events.
	postGenerationWG.Add(1)
	var firstPostGenerationExecuted bool
	go func() {
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
				if p != nil && textUpdated || goUpdated {
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
		if errors.Is(err, context.Canceled) {
			cmd.Log.Debug("Context cancelled, exiting")
			return nil
		}
		if errors.Is(err, FatalError{}) {
			cmd.Log.Debug("Fatal error, exiting")
			return err
		}
		cmd.Log.Error("Error received", slog.Any("error", err))
	}

	// Wait for everything to complete.
	cmd.Log.Debug("Waiting for push handler to complete")
	pushHandlerWG.Wait()
	cmd.Log.Debug("Waiting for event handler to complete")
	eventHandlerWG.Wait()
	cmd.Log.Debug("Waiting for post-generation handler to complete")
	postGenerationWG.Wait()

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
	p = proxy.New(cmd.Args.ProxyPort, target)
	go func() {
		cmd.Log.Info("Proxying", slog.String("from", p.URL), slog.String("to", p.Target.String()))
		if err := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", cmd.Args.ProxyPort), p); err != nil {
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
			cmd.Log.Debug("Proxy not ready, retrying", slog.String("url", p.URL), slog.Any("backoff", d))
			time.Sleep(d)
		}
		if err := browser.OpenURL(p.URL); err != nil {
			cmd.Log.Error("Failed to open browser", slog.Any("error", err))
		}
	}()
	return p, nil
}
