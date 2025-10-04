package generatecmd

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/a-h/templ/internal/skipdir"
	templruntime "github.com/a-h/templ/runtime"
	"golang.org/x/sync/errgroup"

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

func NewGenerate(log *slog.Logger, args Arguments) (g *Generate, err error) {
	g = &Generate{
		Log:  log,
		Args: args,
	}
	return g, nil
}

type Generate struct {
	Log  *slog.Logger
	Args Arguments
}

type GenerationEvent struct {
	Event                fsnotify.Event
	WatchedFileUpdated   bool
	TemplFileTextUpdated bool
	TemplFileGoUpdated   bool
}

func (cmd Generate) Run(ctx context.Context) (err error) {
	if cmd.Args.NotifyProxy {
		return proxy.NotifyProxy(cmd.Args.ProxyBind, cmd.Args.ProxyPort)
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

	// For the initial filesystem walk and subsequent (optional) fsnotify events.
	events := make(chan fsnotify.Event)
	// For errs from the watcher.
	errs := make(chan error)

	// Start process to push events into the events channel.
	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		defer close(events)
		cmd.walkAndWatch(ctx, events, errs)
		return nil
	})

	// For triggering actions after generation has completed.
	postGeneration := make(chan *GenerationEvent, 256)

	// Start process to handle events.
	grp.Go(func() error {
		defer close(postGeneration)
		cmd.handleEvents(ctx, events, errs, fseh, postGeneration)
		return nil
	})

	// Start process to handle post-generation events.
	var updates int
	grp.Go(func() error {
		defer close(errs)
		updates, err = cmd.handlePostGenerationEvents(ctx, postGeneration)
		return err
	})

	// Read errors.
	var errorCount int
	for err := range errs {
		if err == nil {
			continue
		}
		if errors.Is(err, FatalError{}) {
			cmd.Log.Debug("Fatal error, exiting")
			return err
		}
		cmd.Log.Error("Error", slog.Any("error", err))
		errorCount++
	}

	// Wait for everything to complete.
	cmd.Log.Debug("Waiting for processes to complete")
	if err = grp.Wait(); err != nil {
		return err
	}
	if cmd.Args.Command != "" {
		cmd.Log.Debug("Killing command", slog.String("command", cmd.Args.Command))
		if err := run.KillAll(); err != nil {
			cmd.Log.Error("Error killing command", slog.Any("error", err))
		}
	}

	// Clean up temporary watch mode text files.
	if err := cmd.deleteWatchModeTextFiles(); err != nil {
		cmd.Log.Warn("Failed to delete watch mode text files", slog.Any("error", err))
	}

	// Check for errors after everything has completed.
	if errorCount > 0 {
		return fmt.Errorf("generation completed with %d errors", errorCount)
	}

	cmd.Log.Info("Complete", slog.Int("updates", updates), slog.Duration("duration", time.Since(start)))
	return nil
}

func (cmd Generate) groupUntilNoMessagesReceivedFor100ms(postGeneration chan *GenerationEvent) (grouped *GenerationEvent, updates int, ok bool, err error) {
	timeout := time.NewTimer(time.Hour * 24 * 365)
loop:
	for {
		select {
		case ge := <-postGeneration:
			if ge == nil {
				cmd.Log.Debug("Post-generation event channel closed, exiting")
				return nil, 0, false, nil
			}
			if grouped == nil {
				grouped = ge
			}
			grouped.WatchedFileUpdated = grouped.WatchedFileUpdated || ge.WatchedFileUpdated
			grouped.TemplFileTextUpdated = grouped.TemplFileTextUpdated || ge.TemplFileTextUpdated
			grouped.TemplFileGoUpdated = grouped.TemplFileGoUpdated || ge.TemplFileGoUpdated
			if grouped.WatchedFileUpdated || grouped.TemplFileTextUpdated || grouped.TemplFileGoUpdated {
				updates++
			}
			// Now we have received an event, wait for 100ms.
			// If no further messages are received in that time, the timeout will trigger.
			timeout = time.NewTimer(time.Millisecond * 100)
		case <-timeout.C:
			// If grouped is nil, or if no updates were made, reset the timer and continue waiting.
			if grouped == nil || (!grouped.WatchedFileUpdated && !grouped.TemplFileTextUpdated && !grouped.TemplFileGoUpdated) {
				timeout = time.NewTimer(time.Hour * 24 * 365)
				continue loop
			}
			// We have a grouped event, and no events have been sent in the last 100ms, so we need to return.
			return grouped, updates, true, nil
		}
	}
}

func (cmd Generate) handlePostGenerationEvents(ctx context.Context, postGeneration chan *GenerationEvent) (updates int, err error) {
	cmd.Log.Debug("Starting post-generation handler")
	var p *proxy.Handler
loop:
	for {
		grouped, updated, ok, err := cmd.groupUntilNoMessagesReceivedFor100ms(postGeneration)
		if err != nil {
			return 0, fmt.Errorf("error grouping post-generation events: %w", err)
		}
		if !ok {
			break loop
		}

		// The Go application needs to be restarted if any watched non-templ watched files (i.e. non-templ Go files)
		// were updated, or if any Go code within a templ file was updated.
		needsRestart := grouped.WatchedFileUpdated || grouped.TemplFileGoUpdated
		// If the text in a templ file, or any other changes have happened, reload the browser.
		needsBrowserReload := grouped.TemplFileTextUpdated || grouped.TemplFileGoUpdated || grouped.WatchedFileUpdated

		cmd.Log.Info("Post-generation event received, processing...", slog.Bool("needsRestart", needsRestart), slog.Bool("needsBrowserReload", needsBrowserReload))
		updates += updated

		if cmd.Args.Command != "" && needsRestart {
			cmd.Log.Info("Executing command", slog.String("command", cmd.Args.Command))
			if cmd.Args.Watch {
				if err := os.Setenv("TEMPL_DEV_MODE", "true"); err != nil {
					cmd.Log.Error("Error setting TEMPL_DEV_MODE environment variable", slog.Any("error", err))
				}
			}
			if _, err := run.Run(ctx, cmd.Args.Path, cmd.Args.Command); err != nil {
				cmd.Log.Error("Error executing command", slog.Any("error", err))
			}
		}
		if cmd.Args.Proxy != "" {
			if p == nil {
				cmd.Log.Debug("Starting proxy...")
				p, err = cmd.startProxy()
				if err != nil {
					cmd.Log.Error("Failed to start proxy", slog.Any("error", err))
				}
			}
			if needsBrowserReload {
				cmd.Log.Debug("Sending reload event")
				p.SendSSE("message", "reload")
			}
		}
	}
	return updates, nil
}

func (cmd Generate) handleEvents(ctx context.Context, events chan fsnotify.Event, errs chan error, fseh *FSEventHandler, postGeneration chan *GenerationEvent) {
	var eventsWG sync.WaitGroup
	sem := make(chan struct{}, cmd.Args.WorkerCount)
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
			if !r.WatchedFileUpdated && !r.TemplFileTextUpdated && !r.TemplFileGoUpdated {
				cmd.Log.Debug("File not updated", slog.String("file", event.Name))
				return
			}
			e := &GenerationEvent{
				Event:                event,
				WatchedFileUpdated:   r.WatchedFileUpdated,
				TemplFileTextUpdated: r.TemplFileTextUpdated,
				TemplFileGoUpdated:   r.TemplFileGoUpdated,
			}
			cmd.Log.Debug("File updated", slog.String("file", event.Name))
			postGeneration <- e
		}(event)
	}
	// Wait for all events to be processed before closing.
	eventsWG.Wait()
}

func (cmd *Generate) walkAndWatch(ctx context.Context, events chan fsnotify.Event, errs chan error) {
	cmd.Log.Debug("Walking directory", slog.String("path", cmd.Args.Path), slog.Bool("devMode", cmd.Args.Watch))
	if err := watcher.WalkFiles(ctx, cmd.Args.Path, cmd.Args.WatchPattern, cmd.Args.IgnorePattern, events); err != nil {
		cmd.Log.Error("WalkFiles failed, exiting", slog.Any("error", err))
		errs <- FatalError{Err: fmt.Errorf("failed to walk files: %w", err)}
		return
	}
	if !cmd.Args.Watch {
		cmd.Log.Debug("Dev mode not enabled, process can finish early")
		return
	}
	cmd.Log.Info("Watching files")
	rw, err := watcher.Recursive(ctx, cmd.Args.WatchPattern, cmd.Args.IgnorePattern, events, errs)
	if err != nil {
		cmd.Log.Error("Recursive watcher setup failed, exiting", slog.Any("error", err))
		errs <- FatalError{Err: fmt.Errorf("failed to setup recursive watcher: %w", err)}
		return
	}
	if err = rw.Add(cmd.Args.Path); err != nil {
		cmd.Log.Error("Failed to add path to watcher", slog.Any("error", err))
		errs <- FatalError{Err: fmt.Errorf("failed to add path to watcher: %w", err)}
		return
	}
	defer func() {
		if err := rw.Close(); err != nil {
			cmd.Log.Error("Failed to close watcher", slog.Any("error", err))
		}
	}()
	cmd.Log.Debug("Waiting for context to be cancelled to stop watching files")
	<-ctx.Done()
}

func (cmd *Generate) deleteWatchModeTextFiles() error {
	return fs.WalkDir(os.DirFS(cmd.Args.Path), ".", func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		absPath, err := filepath.Abs(filepath.Join(cmd.Args.Path, path))
		if err != nil {
			return nil
		}
		if info.IsDir() && skipdir.ShouldSkip(absPath) {
			return filepath.SkipDir
		}
		if !strings.HasSuffix(absPath, "_templ.go") && !strings.HasSuffix(absPath, ".templ") {
			return nil
		}
		watchModeFileName := templruntime.GetDevModeTextFileName(absPath)
		if err := os.Remove(watchModeFileName); err != nil && !errors.Is(err, os.ErrNotExist) {
			cmd.Log.Warn("Failed to remove watch mode text file", slog.Any("error", err))
		}
		return nil
	})
}

func (cmd *Generate) startProxy() (p *proxy.Handler, err error) {
	var target *url.URL
	target, err = url.Parse(cmd.Args.Proxy)
	if err != nil {
		return nil, FatalError{Err: fmt.Errorf("failed to parse proxy URL: %w", err)}
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
			cmd.Log.Debug("Proxy not ready, retrying", slog.String("url", p.URL), slog.Any("backoff", d))
			time.Sleep(d)
		}
		if err := browser.OpenURL(p.URL); err != nil {
			cmd.Log.Error("Failed to open browser", slog.Any("error", err))
		}
	}()
	return p, nil
}
