package generatecmd

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"go/format"
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
		Log:                        log,
		dir:                        dir,
		fileNameToLastModTime:      make(map[string]time.Time),
		fileNameToLastModTimeMutex: &sync.Mutex{},
		hashes:                     make(map[string][sha256.Size]byte),
		hashesMutex:                &sync.Mutex{},
		genOpts:                    genOpts,
		genSourceMapVis:            genSourceMapVis,
		DevMode:                    devMode,
		keepOrphanedFiles:          keepOrphanedFiles,
	}
	if devMode {
		fseh.genOpts = append(fseh.genOpts, generator.WithExtractStrings())
	}
	return fseh
}

type FSEventHandler struct {
	Log *slog.Logger
	// dir is the root directory being processed.
	dir                        string
	fileNameToLastModTime      map[string]time.Time
	fileNameToLastModTimeMutex *sync.Mutex
	hashes                     map[string][sha256.Size]byte
	hashesMutex                *sync.Mutex
	genOpts                    []generator.GenerateOpt
	genSourceMapVis            bool
	DevMode                    bool
	keepOrphanedFiles          bool
}

func (h *FSEventHandler) HandleEvent(ctx context.Context, event fsnotify.Event) (goUpdated, textUpdated bool, err error) {
	// Handle _templ.go files.
	if !event.Has(fsnotify.Remove) && strings.HasSuffix(event.Name, "_templ.go") {
		_, err = os.Stat(strings.TrimSuffix(event.Name, "_templ.go") + ".templ")
		if !os.IsNotExist(err) {
			return false, false, err
		}
		// File is orphaned.
		if h.keepOrphanedFiles {
			return false, false, nil
		}
		h.Log.Debug("Deleting orphaned Go file", slog.String("file", event.Name))
		if err = os.Remove(event.Name); err != nil {
			h.Log.Warn("Failed to remove orphaned file", slog.Any("error", err))
		}
		return true, false, nil
	}
	// Handle _templ.txt files.
	if !event.Has(fsnotify.Remove) && strings.HasSuffix(event.Name, "_templ.txt") {
		if h.DevMode {
			// Don't delete the file if we're in dev mode, but mark that text was updated.
			return false, true, nil
		}
		h.Log.Debug("Deleting watch mode file", slog.String("file", event.Name))
		if err = os.Remove(event.Name); err != nil {
			h.Log.Warn("Failed to remove watch mode text file", slog.Any("error", err))
			return false, false, nil
		}
		return false, false, nil
	}

	// Handle .templ files.
	if !strings.HasSuffix(event.Name, ".templ") {
		return false, false, nil
	}

	// If the file hasn't been updated since the last time we processed it, ignore it.
	if !h.UpsertLastModTime(event.Name) {
		return false, false, nil
	}

	// Start a processor.
	start := time.Now()
	goUpdated, textUpdated, diag, err := h.generate(ctx, event.Name)
	if err != nil {
		h.Log.Error("Error generating code", slog.String("file", event.Name), slog.Any("error", err))
		return goUpdated, textUpdated, fmt.Errorf("failed to generate code for %q: %w", event.Name, err)
	}
	if len(diag) > 0 {
		for _, d := range diag {
			h.Log.Warn(d.Message, slog.String("from", fmt.Sprintf("%d:%d", d.Range.From.Line, d.Range.From.Col)), slog.String("to", fmt.Sprintf("%d:%d", d.Range.To.Line, d.Range.To.Col)))
		}
		return
	}
	h.Log.Debug("Generated code", slog.String("file", event.Name), slog.Duration("in", time.Since(start)))

	return goUpdated, textUpdated, nil
}

func (h *FSEventHandler) UpsertLastModTime(fileName string) (updated bool) {
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return false
	}
	h.fileNameToLastModTimeMutex.Lock()
	defer h.fileNameToLastModTimeMutex.Unlock()
	lastModTime := h.fileNameToLastModTime[fileName]
	if !fileInfo.ModTime().After(lastModTime) {
		return false
	}
	h.fileNameToLastModTime[fileName] = fileInfo.ModTime()
	return true
}

func (h *FSEventHandler) UpsertHash(fileName string, hash [sha256.Size]byte) (updated bool) {
	h.hashesMutex.Lock()
	defer h.hashesMutex.Unlock()
	lastHash := h.hashes[fileName]
	if lastHash == hash {
		return false
	}
	h.hashes[fileName] = hash
	return true
}

// generate Go code for a single template.
// If a basePath is provided, the filename included in error messages is relative to it.
func (h *FSEventHandler) generate(ctx context.Context, fileName string) (goUpdated, textUpdated bool, diagnostics []parser.Diagnostic, err error) {
	t, err := parser.Parse(fileName)
	if err != nil {
		return false, false, nil, fmt.Errorf("%s parsing error: %w", fileName, err)
	}
	targetFileName := strings.TrimSuffix(fileName, ".templ") + "_templ.go"

	// Only use relative filenames to the basepath for filenames in runtime error messages.
	relFilePath, err := filepath.Rel(h.dir, fileName)
	if err != nil {
		return false, false, nil, fmt.Errorf("failed to get relative path for %q: %w", fileName, err)
	}

	var b bytes.Buffer
	sourceMap, literals, err := generator.Generate(t, &b, append(h.genOpts, generator.WithFileName(relFilePath))...)
	if err != nil {
		return false, false, nil, fmt.Errorf("%s generation error: %w", fileName, err)
	}

	formattedGoCode, err := format.Source(b.Bytes())
	if err != nil {
		return false, false, nil, fmt.Errorf("%s source formatting error: %w", fileName, err)
	}

	// Hash output, and write out the file if the goCodeHash has changed.
	goCodeHash := sha256.Sum256(formattedGoCode)
	if h.UpsertHash(targetFileName, goCodeHash) {
		goUpdated = true
		if err = os.WriteFile(targetFileName, formattedGoCode, 0o644); err != nil {
			return false, false, nil, fmt.Errorf("failed to write target file %q: %w", targetFileName, err)
		}
	}

	// Add the txt file if it has changed.
	if len(literals) > 0 {
		txtFileName := strings.TrimSuffix(fileName, ".templ") + "_templ.txt"
		txtHash := sha256.Sum256([]byte(literals))
		if h.UpsertHash(txtFileName, txtHash) {
			textUpdated = true
			if err = os.WriteFile(txtFileName, []byte(literals), 0o644); err != nil {
				return false, false, nil, fmt.Errorf("failed to write string literal file %q: %w", txtFileName, err)
			}
		}
	}

	if h.genSourceMapVis {
		err = generateSourceMapVisualisation(ctx, fileName, targetFileName, sourceMap)
	}

	return goUpdated, textUpdated, t.Diagnostics, err
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
