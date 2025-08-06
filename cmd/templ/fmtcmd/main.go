package fmtcmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/a-h/templ/cmd/templ/imports"
	"github.com/a-h/templ/cmd/templ/processor"
	parser "github.com/a-h/templ/parser/v2"
	"github.com/natefinch/atomic"
)

type Arguments struct {
	FailIfChanged bool
	ToStdout      bool
	StdinFilepath string
	Files         []string
	WorkerCount   int
}

func Run(log *slog.Logger, stdin io.Reader, stdout io.Writer, args Arguments) (err error) {
	// If no files are provided, read from stdin and write to stdout.
	if len(args.Files) == 0 {
		out, _ := format(writeToWriter(stdout), readFromReader(stdin, args.StdinFilepath), true)
		return out
	}
	process := func(fileName string) (error, bool) {
		read := readFromFile(fileName)
		write := writeToFile
		if args.ToStdout {
			write = writeToWriter(stdout)
		}
		writeIfUnchanged := args.ToStdout
		return format(write, read, writeIfUnchanged)
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

type reader func() (fileName, src string, err error)

func readFromReader(r io.Reader, stdinFilepath string) func() (fileName, src string, err error) {
	return func() (fileName, src string, err error) {
		b, err := io.ReadAll(r)
		if err != nil {
			return "", "", fmt.Errorf("failed to read stdin: %w", err)
		}
		return stdinFilepath, string(b), nil
	}
}

func readFromFile(name string) reader {
	return func() (fileName, src string, err error) {
		b, err := os.ReadFile(name)
		if err != nil {
			return "", "", fmt.Errorf("failed to read file %q: %w", fileName, err)
		}
		return name, string(b), nil
	}
}

type writer func(fileName, tgt string) error

var mu sync.Mutex

func writeToWriter(w io.Writer) func(fileName, tgt string) error {
	return func(fileName, tgt string) error {
		mu.Lock()
		defer mu.Unlock()
		_, err := w.Write([]byte(tgt))
		return err
	}
}

func writeToFile(fileName, tgt string) error {
	return atomic.WriteFile(fileName, bytes.NewBufferString(tgt))
}

func format(write writer, read reader, writeIfUnchanged bool) (err error, fileChanged bool) {
	fileName, src, err := read()
	if err != nil {
		return err, false
	}
	t, err := parser.ParseString(src)
	if err != nil {
		return err, false
	}
	t.Filepath = fileName
	t, err = imports.Process(t)
	if err != nil {
		return err, false
	}
	w := new(bytes.Buffer)
	if err = t.Write(w); err != nil {
		return fmt.Errorf("formatting error: %w", err), false
	}

	fileChanged = (src != w.String())

	if !writeIfUnchanged && !fileChanged {
		return nil, fileChanged
	}
	return write(fileName, w.String()), fileChanged
}
