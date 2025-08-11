package fmtcmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/a-h/templ/cmd/templ/processor"
	"github.com/a-h/templ/internal/format"
	"github.com/natefinch/atomic"
)

type Arguments struct {
	FailIfChanged    bool
	ToStdout         bool
	StdinFilepath    string
	Files            []string
	WorkerCount      int
	PrettierCommand  string
	PrettierRequired bool
}

func Run(log *slog.Logger, stdin io.Reader, stdout io.Writer, args Arguments) (err error) {
	// If no files are provided, read from stdin and write to stdout.
	formatterConfig := format.Config{
		PrettierCommand:  args.PrettierCommand,
		PrettierRequired: args.PrettierRequired,
	}
	if len(args.Files) == 0 {
		src, err := io.ReadAll(stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		formatted, _, err := format.Templ(src, args.StdinFilepath, formatterConfig)
		if err != nil {
			return fmt.Errorf("failed to format stdin: %w", err)
		}
		if _, err = stdout.Write(formatted); err != nil {
			return fmt.Errorf("failed to write to stdout: %w", err)
		}
		return nil
	}
	// If files are provided, process each file.
	process := func(fileName string) (error, bool) {
		src, err := os.ReadFile(fileName)
		if err != nil {
			return fmt.Errorf("failed to read file %q: %w", fileName, err), false
		}
		formatted, changed, err := format.Templ(src, fileName, formatterConfig)
		if err != nil {
			return fmt.Errorf("failed to format file %q: %w", fileName, err), false
		}
		if !changed && !args.ToStdout {
			return nil, false
		}
		if args.ToStdout {
			if _, err := stdout.Write(formatted); err != nil {
				return fmt.Errorf("failed to write to stdout: %w", err), false
			}
			return nil, true
		}
		if err := atomic.WriteFile(fileName, bytes.NewBuffer(formatted)); err != nil {
			return fmt.Errorf("failed to write file %q: %w", fileName, err), false
		}
		return nil, true
	}
	dir := args.Files[0]
	return NewFormatter(log, dir, process, args.WorkerCount, args.FailIfChanged).Run()
}

type Formatter struct {
	Log          *slog.Logger
	Dir          string
	Process      func(fileName string) (error, bool)
	WorkerCount  int
	FailIfChange bool
}

func NewFormatter(log *slog.Logger, dir string, process func(fileName string) (error, bool), workerCount int, failIfChange bool) *Formatter {
	f := &Formatter{
		Log:          log,
		Dir:          dir,
		Process:      process,
		WorkerCount:  workerCount,
		FailIfChange: failIfChange,
	}
	if f.WorkerCount == 0 {
		f.WorkerCount = runtime.NumCPU()
	}
	return f
}

func (f *Formatter) Run() (err error) {
	var errs []error
	changesMade := 0
	start := time.Now()
	results := make(chan processor.Result)
	f.Log.Debug("Walking directory", slog.String("path", f.Dir))
	go processor.Process(f.Dir, f.Process, f.WorkerCount, results)
	var successCount, errorCount int
	for r := range results {
		if r.ChangesMade {
			changesMade += 1
		}
		if r.Error != nil {
			f.Log.Error(r.FileName, slog.Any("error", r.Error))
			errorCount++
			errs = append(errs, r.Error)
			continue
		}
		f.Log.Debug(r.FileName, slog.Duration("duration", r.Duration))
		successCount++
	}

	if f.FailIfChange && changesMade > 0 {
		f.Log.Error("Templates were valid but not properly formatted", slog.Int("count", successCount+errorCount), slog.Int("changed", changesMade), slog.Int("errors", errorCount), slog.Duration("duration", time.Since(start)))
		return fmt.Errorf("templates were not formatted properly")
	}

	f.Log.Info("Format Complete", slog.Int("count", successCount+errorCount), slog.Int("errors", errorCount), slog.Int("changed", changesMade), slog.Duration("duration", time.Since(start)))

	if err = errors.Join(errs...); err != nil {
		return fmt.Errorf("formatting failed: %w", err)
	}

	return nil
}
