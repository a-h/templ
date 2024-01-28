package generatecmd

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"go/format"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/a-h/templ/cmd/templ/visualize"
	"github.com/a-h/templ/generator"
	"github.com/a-h/templ/parser/v2"
	"github.com/fsnotify/fsnotify"
)

func NewFSEventHandler(log *slog.Logger, dir string, devMode bool, genOpts []generator.GenerateOpt, genSourceMapVis bool, keepOrphanedFiles bool) *FSEventHandler {
	if !path.IsAbs(dir) {
		dir, _ = filepath.Abs(dir)
	}
	fseh := &FSEventHandler{
		Log:                   log,
		dir:                   dir,
		stdout:                os.Stdout,
		fileNameToLastModTime: make(map[string]time.Time),
		hashes:                make(map[string][sha256.Size]byte),
		genOpts:               genOpts,
		genSourceMapVis:       genSourceMapVis,
		DevMode:               devMode,
		keepOrphanedFiles:     keepOrphanedFiles,
	}
	if devMode {
		fseh.genOpts = append(fseh.genOpts, generator.WithExtractStrings())
	}
	return fseh
}

type FSEventHandler struct {
	Log *slog.Logger
	// dir is the root directory being processed.
	dir                   string
	stdout                io.Writer
	stderr                io.Writer
	fileNameToLastModTime map[string]time.Time
	hashes                map[string][sha256.Size]byte
	genOpts               []generator.GenerateOpt
	genSourceMapVis       bool
	DevMode               bool
	keepOrphanedFiles     bool
}

func (h *FSEventHandler) HandleEvent(ctx context.Context, event fsnotify.Event) (generated bool, err error) {
	// Handle _templ.go files.
	if !event.Has(fsnotify.Remove) && strings.HasSuffix(event.Name, "_templ.go") {
		_, err = os.Stat(strings.TrimSuffix(event.Name, "_templ.go") + ".templ")
		if err != nil {
			return false, nil
		}
		// File is orphaned.
		if h.keepOrphanedFiles {
			return false, nil
		}
		if err = os.Remove(event.Name); err != nil {
			h.Log.Warn("Failed to remove orphaned file", slog.Any("error", err))
		}
		return true, nil
	}
	// Handle _templ.txt files.
	if !event.Has(fsnotify.Remove) && strings.HasSuffix(event.Name, "_templ.txt") {
		if h.DevMode {
			// Don't do anything in watch mode.
			return false, nil
		}
		if err = os.Remove(event.Name); err != nil {
			h.Log.Warn("Failed to remove watch mode text file", slog.Any("error", err))
			return false, nil
		}
		h.Log.Debug("Deleted watch mode file", slog.String("file", event.Name))
		return false, nil
	}

	// Handle .templ files.
	if !strings.HasSuffix(event.Name, ".templ") {
		return false, nil
	}

	// If the file hasn't been updated since the last time we processed it, ignore it.
	lastModTime := h.fileNameToLastModTime[event.Name]
	fileInfo, err := os.Stat(event.Name)
	if err != nil {
		return false, fmt.Errorf("failed to get file info: %w", err)
	}
	if fileInfo.ModTime().Before(lastModTime) {
		return false, nil
	}

	// Start a processor.
	h.fileNameToLastModTime[event.Name] = fileInfo.ModTime()

	start := time.Now()
	diag, err := h.generate(ctx, event.Name)
	if err != nil {
		h.Log.Error("Error generating code", slog.String("file", event.Name), slog.Any("error", err))
		return false, fmt.Errorf("failed to generate code for %q: %w", event.Name, err)
	}
	if len(diag) > 0 {
		for _, d := range diag {
			h.Log.Warn(d.Message, slog.String("from", fmt.Sprintf("%d:%d", d.Range.From.Line, d.Range.From.Col)), slog.String("to", fmt.Sprintf("%d:%d", d.Range.To.Line, d.Range.To.Col)))
		}
		return
	}
	h.Log.Debug("Generated code for %q in %s\n", event.Name, time.Since(start))

	return true, nil
}

// generate Go code for a single template.
// If a basePath is provided, the filename included in error messages is relative to it.
func (h *FSEventHandler) generate(ctx context.Context, fileName string) (diagnostics []parser.Diagnostic, err error) {
	t, err := parser.Parse(fileName)
	if err != nil {
		return nil, fmt.Errorf("%s parsing error: %w", fileName, err)
	}
	targetFileName := strings.TrimSuffix(fileName, ".templ") + "_templ.go"

	// Only use relative filenames to the basepath for filenames in runtime error messages.
	relFilePath, err := filepath.Rel(h.dir, fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to get relative path for %q: %w", fileName, err)
	}

	var b bytes.Buffer
	sourceMap, literals, err := generator.Generate(t, &b, append(h.genOpts, generator.WithFileName(relFilePath))...)
	if err != nil {
		return nil, fmt.Errorf("%s generation error: %w", fileName, err)
	}

	formattedGoCode, err := format.Source(b.Bytes())
	if err != nil {
		return nil, fmt.Errorf("%s source formatting error: %w", fileName, err)
	}

	// Hash output, and write out the file if the goCodeHash has changed.
	goCodeHash := sha256.Sum256(formattedGoCode)
	if h.hashes[targetFileName] != goCodeHash {
		if err = os.WriteFile(targetFileName, formattedGoCode, 0o644); err != nil {
			return nil, fmt.Errorf("failed to write target file %q: %w", targetFileName, err)
		}
		h.hashes[targetFileName] = goCodeHash
	}

	// Add the txt file if it has changed.
	if len(literals) > 0 {
		txtFileName := strings.TrimSuffix(fileName, ".templ") + "_templ.txt"
		txtHash := sha256.Sum256([]byte(literals))
		if h.hashes[txtFileName] != txtHash {
			if err = os.WriteFile(txtFileName, []byte(literals), 0o644); err != nil {
				return nil, fmt.Errorf("failed to write string literal file %q: %w", txtFileName, err)
			}
			h.hashes[txtFileName] = txtHash
		}
	}

	if h.genSourceMapVis {
		err = generateSourceMapVisualisation(ctx, fileName, targetFileName, sourceMap)
	}
	return t.Diagnostics, err
}

func generateSourceMapVisualisation(ctx context.Context, templFileName, goFileName string, sourceMap *parser.SourceMap) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	var templContents, goContents []byte
	var templErr, goErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		templContents, templErr = os.ReadFile(templFileName)
	}()
	go func() {
		defer wg.Done()
		goContents, goErr = os.ReadFile(goFileName)
	}()
	wg.Wait()
	if templErr != nil {
		return templErr
	}
	if goErr != nil {
		return templErr
	}

	targetFileName := strings.TrimSuffix(templFileName, ".templ") + "_templ_sourcemap.html"
	w, err := os.Create(targetFileName)
	if err != nil {
		return fmt.Errorf("%s sourcemap visualisation error: %w", templFileName, err)
	}
	defer w.Close()
	b := bufio.NewWriter(w)
	defer b.Flush()

	return visualize.HTML(templFileName, string(templContents), string(goContents), sourceMap).Render(ctx, b)
}
