package generatecmd

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"go/format"
	"go/scanner"
	"go/token"
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

type FileWriterFunc func(name string, contents []byte) error

func FileWriter(fileName string, contents []byte) error {
	return os.WriteFile(fileName, contents, 0o644)
}

func WriterFileWriter(w io.Writer) FileWriterFunc {
	return func(_ string, contents []byte) error {
		_, err := w.Write(contents)
		return err
	}
}

func NewFSEventHandler(
	log *slog.Logger,
	dir string,
	devMode bool,
	genOpts []generator.GenerateOpt,
	genSourceMapVis bool,
	keepOrphanedFiles bool,
	fileWriter FileWriterFunc,
	lazy bool,
) *FSEventHandler {
	if !path.IsAbs(dir) {
		dir, _ = filepath.Abs(dir)
	}
	fseh := &FSEventHandler{
		Log:                        log,
		dir:                        dir,
		fileNameToLastModTime:      make(map[string]time.Time),
		fileNameToLastModTimeMutex: &sync.Mutex{},
		fileNameToError:            make(map[string]struct{}),
		fileNameToErrorMutex:       &sync.Mutex{},
		fileNameToOutput:           make(map[string]generator.GeneratorOutput),
		fileNameToOutputMutex:      &sync.Mutex{},
		devMode:                    devMode,
		hashes:                     make(map[string][sha256.Size]byte),
		hashesMutex:                &sync.Mutex{},
		genOpts:                    genOpts,
		genSourceMapVis:            genSourceMapVis,
		keepOrphanedFiles:          keepOrphanedFiles,
		writer:                     fileWriter,
		lazy:                       lazy,
	}
	return fseh
}

type FSEventHandler struct {
	Log *slog.Logger
	// dir is the root directory being processed.
	dir                        string
	fileNameToLastModTime      map[string]time.Time
	fileNameToLastModTimeMutex *sync.Mutex
	fileNameToError            map[string]struct{}
	fileNameToErrorMutex       *sync.Mutex
	fileNameToOutput           map[string]generator.GeneratorOutput
	fileNameToOutputMutex      *sync.Mutex
	devMode                    bool
	hashes                     map[string][sha256.Size]byte
	hashesMutex                *sync.Mutex
	genOpts                    []generator.GenerateOpt
	genSourceMapVis            bool
	Errors                     []error
	keepOrphanedFiles          bool
	writer                     func(string, []byte) error
	lazy                       bool
}

func (h *FSEventHandler) HandleEvent(ctx context.Context, event fsnotify.Event) (updated, goUpdated, textUpdated bool, err error) {
	// Handle _templ.go files.
	if !event.Has(fsnotify.Remove) && strings.HasSuffix(event.Name, "_templ.go") {
		_, err = os.Stat(strings.TrimSuffix(event.Name, "_templ.go") + ".templ")
		if !os.IsNotExist(err) {
			return false, false, false, err
		}
		// File is orphaned.
		if h.keepOrphanedFiles {
			return false, false, false, nil
		}
		h.Log.Debug("Deleting orphaned Go file", slog.String("file", event.Name))
		if err = os.Remove(event.Name); err != nil {
			h.Log.Warn("Failed to remove orphaned file", slog.Any("error", err))
		}
		return true, true, false, nil
	}
	// Handle _templ.txt files.
	if !event.Has(fsnotify.Remove) && strings.HasSuffix(event.Name, "_templ.txt") {
		if h.devMode {
			// Don't delete the file if we're in dev mode, but mark that text was updated.
			return false, false, true, nil
		}
		h.Log.Debug("Deleting watch mode file", slog.String("file", event.Name))
		if err = os.Remove(event.Name); err != nil {
			h.Log.Warn("Failed to remove watch mode text file", slog.Any("error", err))
			return false, false, false, nil
		}
		return false, false, false, nil
	}

	// Handle .templ files.
	if !strings.HasSuffix(event.Name, ".templ") {
		return false, false, false, nil
	}

	// If the file hasn't been updated since the last time we processed it, ignore it.
	lastModTime, updatedModTime := h.UpsertLastModTime(event.Name)
	if !updatedModTime {
		h.Log.Debug("Skipping file because it wasn't updated", slog.String("file", event.Name))
		return false, false, false, nil
	}
	// If the go file is newer than the templ file, skip generation, because it's up-to-date.
	if h.lazy && goFileIsUpToDate(event.Name, lastModTime) {
		h.Log.Debug("Skipping file because the Go file is up-to-date", slog.String("file", event.Name))
		return false, false, false, nil
	}

	// Start a processor.
	start := time.Now()
	updated, goUpdated, textUpdated, diag, err := h.generate(ctx, event.Name)
	if err != nil {
		h.SetError(event.Name, true)
		return updated, goUpdated, textUpdated, fmt.Errorf("failed to generate code for %q: %w", event.Name, err)
	}
	if len(diag) > 0 {
		for _, d := range diag {
			h.Log.Warn(d.Message,
				slog.String("from", fmt.Sprintf("%d:%d", d.Range.From.Line, d.Range.From.Col)),
				slog.String("to", fmt.Sprintf("%d:%d", d.Range.To.Line, d.Range.To.Col)),
			)
		}
		return
	}
	if errorCleared, errorCount := h.SetError(event.Name, false); errorCleared {
		h.Log.Info("Error cleared", slog.String("file", event.Name), slog.Int("errors", errorCount))
	}
	h.Log.Debug("Generated code", slog.String("file", event.Name), slog.Duration("in", time.Since(start)))

	return updated, goUpdated, textUpdated, nil
}

func goFileIsUpToDate(templFileName string, templFileLastMod time.Time) (upToDate bool) {
	goFileName := strings.TrimSuffix(templFileName, ".templ") + "_templ.go"
	goFileInfo, err := os.Stat(goFileName)
	if err != nil {
		return false
	}
	return goFileInfo.ModTime().After(templFileLastMod)
}

func (h *FSEventHandler) SetError(fileName string, hasError bool) (previouslyHadError bool, errorCount int) {
	h.fileNameToErrorMutex.Lock()
	defer h.fileNameToErrorMutex.Unlock()
	_, previouslyHadError = h.fileNameToError[fileName]
	delete(h.fileNameToError, fileName)
	if hasError {
		h.fileNameToError[fileName] = struct{}{}
	}
	return previouslyHadError, len(h.fileNameToError)
}

func (h *FSEventHandler) UpsertLastModTime(fileName string) (modTime time.Time, updated bool) {
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return modTime, false
	}
	h.fileNameToLastModTimeMutex.Lock()
	defer h.fileNameToLastModTimeMutex.Unlock()
	previousModTime := h.fileNameToLastModTime[fileName]
	currentModTime := fileInfo.ModTime()
	if !currentModTime.After(previousModTime) {
		return currentModTime, false
	}
	h.fileNameToLastModTime[fileName] = currentModTime
	return currentModTime, true
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
func (h *FSEventHandler) generate(ctx context.Context, fileName string) (updated, goUpdated, textUpdated bool, diagnostics []parser.Diagnostic, err error) {
	t, err := parser.Parse(fileName)
	if err != nil {
		return false, false, false, nil, fmt.Errorf("%s parsing error: %w", fileName, err)
	}
	targetFileName := strings.TrimSuffix(fileName, ".templ") + "_templ.go"

	// Only use relative filenames to the basepath for filenames in runtime error messages.
	absFilePath, err := filepath.Abs(fileName)
	if err != nil {
		return false, false, false, nil, fmt.Errorf("failed to get absolute path for %q: %w", fileName, err)
	}
	relFilePath, err := filepath.Rel(h.dir, absFilePath)
	if err != nil {
		return false, false, false, nil, fmt.Errorf("failed to get relative path for %q: %w", fileName, err)
	}
	// Convert Windows file paths to Unix-style for consistency.
	relFilePath = filepath.ToSlash(relFilePath)

	var b bytes.Buffer
	generatorOutput, err := generator.Generate(t, &b, append(h.genOpts, generator.WithFileName(relFilePath))...)
	if err != nil {
		return false, false, false, nil, fmt.Errorf("%s generation error: %w", fileName, err)
	}

	formattedGoCode, err := format.Source(b.Bytes())
	if err != nil {
		err = remapErrorList(err, generatorOutput.SourceMap, fileName)
		if fileName == "/Users/adrian/github.com/a-h/templ/generator/test-whitespace-around-go-keywords/template.templ" {
			fmt.Println(string(b.Bytes()))
		}
		return false, false, false, nil, fmt.Errorf("%s source formatting error %w", fileName, err)
	}

	// Hash output, and write out the file if the goCodeHash has changed.
	goCodeHash := sha256.Sum256(formattedGoCode)
	if h.UpsertHash(targetFileName, goCodeHash) {
		updated = true
		if err = h.writer(targetFileName, formattedGoCode); err != nil {
			return false, false, false, nil, fmt.Errorf("failed to write target file %q: %w", targetFileName, err)
		}
	}

	// Add the txt file if it has changed.
	if h.devMode {
		txtFileName := strings.TrimSuffix(fileName, ".templ") + "_templ.txt"
		joined := strings.Join(generatorOutput.Literals, "\n")
		txtHash := sha256.Sum256([]byte(joined))
		if h.UpsertHash(txtFileName, txtHash) {
			textUpdated = true
			if err = os.WriteFile(txtFileName, []byte(joined), 0o644); err != nil {
				return false, false, false, nil, fmt.Errorf("failed to write string literal file %q: %w", txtFileName, err)
			}

			// Check whether the change would require a recompilation to take effect.
			h.fileNameToOutputMutex.Lock()
			defer h.fileNameToOutputMutex.Unlock()
			previous := h.fileNameToOutput[fileName]
			if generator.HasChanged(previous, generatorOutput) {
				goUpdated = true
			}
			h.fileNameToOutput[fileName] = generatorOutput
		}
	}

	parsedDiagnostics, err := parser.Diagnose(t)
	if err != nil {
		return updated, goUpdated, textUpdated, nil, fmt.Errorf("%s diagnostics error: %w", fileName, err)
	}

	if h.genSourceMapVis {
		err = generateSourceMapVisualisation(ctx, fileName, targetFileName, generatorOutput.SourceMap)
	}

	return updated, goUpdated, textUpdated, parsedDiagnostics, err
}

// Takes an error from the formatter and attempts to convert the positions reported in the target file to their positions
// in the source file.
func remapErrorList(err error, sourceMap *parser.SourceMap, fileName string) error {
	list, ok := err.(scanner.ErrorList)
	if !ok || len(list) == 0 {
		return err
	}
	for i, e := range list {
		// The positions in the source map are off by one line because of the package definition.
		srcPos, ok := sourceMap.SourcePositionFromTarget(uint32(e.Pos.Line-1), uint32(e.Pos.Column))
		if !ok {
			continue
		}
		list[i].Pos = token.Position{
			Filename: fileName,
			Offset:   int(srcPos.Index),
			Line:     int(srcPos.Line) + 1,
			Column:   int(srcPos.Col),
		}
	}
	return list
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
