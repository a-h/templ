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
	"time"

	"github.com/a-h/templ/cmd/templ/visualize"
	"github.com/a-h/templ/generator"
	"github.com/a-h/templ/internal/syncmap"
	"github.com/a-h/templ/internal/syncset"
	"github.com/a-h/templ/parser/v2"
	"github.com/a-h/templ/runtime"
	"github.com/fsnotify/fsnotify"
	"golang.org/x/sync/errgroup"
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
		Log:                   log,
		dir:                   dir,
		fileNameToLastModTime: syncmap.New[string, time.Time](),
		fileNameToError:       syncset.New[string](),
		fileNameToOutput:      syncmap.New[string, generator.GeneratorOutput](),
		devMode:               devMode,
		hashes:                syncmap.New[string, [sha256.Size]byte](),
		genOpts:               genOpts,
		genSourceMapVis:       genSourceMapVis,
		keepOrphanedFiles:     keepOrphanedFiles,
		writer:                fileWriter,
		lazy:                  lazy,
	}
	return fseh
}

type FSEventHandler struct {
	Log *slog.Logger
	// dir is the root directory being processed.
	dir                   string
	fileNameToLastModTime *syncmap.Map[string, time.Time]
	fileNameToError       *syncset.Set[string]
	fileNameToOutput      *syncmap.Map[string, generator.GeneratorOutput]
	devMode               bool
	hashes                *syncmap.Map[string, [sha256.Size]byte]
	genOpts               []generator.GenerateOpt
	genSourceMapVis       bool
	Errors                []error
	keepOrphanedFiles     bool
	writer                FileWriterFunc
	lazy                  bool
}

type GenerateResult struct {
	// WatchedFileUpdated indicates that a file matching the watch pattern was updated.
	WatchedFileUpdated bool
	// TemplFileTextUpdated indicates that text literals were updated.
	TemplFileTextUpdated bool
	// TemplFileGoUpdated indicates that Go expressions were updated.
	TemplFileGoUpdated bool
}

func (h *FSEventHandler) HandleEvent(ctx context.Context, event fsnotify.Event) (result GenerateResult, err error) {
	// Handle _templ.go files.
	if !event.Has(fsnotify.Remove) && strings.HasSuffix(event.Name, "_templ.go") {
		_, err = os.Stat(strings.TrimSuffix(event.Name, "_templ.go") + ".templ")
		if !os.IsNotExist(err) {
			return GenerateResult{}, err
		}
		// File is orphaned.
		if h.keepOrphanedFiles {
			return GenerateResult{}, nil
		}
		h.Log.Debug("Deleting orphaned Go file", slog.String("file", event.Name))
		if err = os.Remove(event.Name); err != nil {
			h.Log.Warn("Failed to remove orphaned file", slog.Any("error", err))
		}
		return GenerateResult{WatchedFileUpdated: false, TemplFileGoUpdated: true, TemplFileTextUpdated: false}, nil
	}

	// If the file hasn't been updated since the last time we processed it, ignore it.
	fileInfo, err := os.Stat(event.Name)
	if err != nil {
		return GenerateResult{}, fmt.Errorf("failed to stat %q: %w", event.Name, err)
	}
	mustBeInTheFuture := func(previous, updated time.Time) bool {
		return updated.After(previous)
	}
	updatedModTime := h.fileNameToLastModTime.CompareAndSwap(event.Name, mustBeInTheFuture, fileInfo.ModTime())
	if !updatedModTime {
		h.Log.Debug("Skipping file because it wasn't updated", slog.String("file", event.Name))
		return GenerateResult{}, nil
	}

	// Process anything that isn't a templ file.
	if !strings.HasSuffix(event.Name, ".templ") {
		if h.devMode {
			h.Log.Info("Watched file updated", slog.String("file", event.Name))
		}
		result.WatchedFileUpdated = true
		return result, nil
	}

	// Handle templ files.

	// If the go file is newer than the templ file, skip generation, because it's up-to-date.
	if h.lazy && goFileIsUpToDate(event.Name, fileInfo.ModTime()) {
		h.Log.Debug("Skipping file because the Go file is up-to-date", slog.String("file", event.Name))
		return GenerateResult{}, nil
	}

	// Start a processor.
	start := time.Now()
	var diag []parser.Diagnostic
	result, diag, err = h.generate(ctx, event.Name)
	if err != nil {
		h.fileNameToError.Set(event.Name)
		return result, fmt.Errorf("failed to generate code for %q: %w", event.Name, err)
	}
	if len(diag) > 0 {
		for _, d := range diag {
			h.Log.Warn(d.Message,
				slog.String("from", fmt.Sprintf("%d:%d", d.Range.From.Line, d.Range.From.Col)),
				slog.String("to", fmt.Sprintf("%d:%d", d.Range.To.Line, d.Range.To.Col)),
			)
		}
		return result, nil
	}
	if errorCleared := h.fileNameToError.Delete(event.Name); errorCleared {
		h.Log.Info("Error cleared", slog.String("file", event.Name), slog.Int("errors", h.fileNameToError.Count()))
	}
	h.Log.Debug("Generated code", slog.String("file", event.Name), slog.Duration("in", time.Since(start)))

	return result, nil
}

func goFileIsUpToDate(templFileName string, templFileLastMod time.Time) (upToDate bool) {
	goFileName := strings.TrimSuffix(templFileName, ".templ") + "_templ.go"
	goFileInfo, err := os.Stat(goFileName)
	if err != nil {
		return false
	}
	return goFileInfo.ModTime().After(templFileLastMod)
}

// generate Go code for a single template.
// If a basePath is provided, the filename included in error messages is relative to it.
func (h *FSEventHandler) generate(ctx context.Context, fileName string) (result GenerateResult, diagnostics []parser.Diagnostic, err error) {
	t, err := parser.Parse(fileName)
	if err != nil {
		return GenerateResult{}, nil, fmt.Errorf("%s parsing error: %w", fileName, err)
	}
	targetFileName := strings.TrimSuffix(fileName, ".templ") + "_templ.go"

	// Only use relative filenames to the basepath for filenames in runtime error messages.
	absFilePath, err := filepath.Abs(fileName)
	if err != nil {
		return GenerateResult{}, nil, fmt.Errorf("failed to get absolute path for %q: %w", fileName, err)
	}
	relFilePath, err := filepath.Rel(h.dir, absFilePath)
	if err != nil {
		return GenerateResult{}, nil, fmt.Errorf("failed to get relative path for %q: %w", fileName, err)
	}
	// Convert Windows file paths to Unix-style for consistency.
	relFilePath = filepath.ToSlash(relFilePath)

	var b bytes.Buffer
	generatorOutput, err := generator.Generate(t, &b, append(h.genOpts, generator.WithFileName(relFilePath))...)
	if err != nil {
		return GenerateResult{}, nil, fmt.Errorf("%s generation error: %w", fileName, err)
	}

	formattedGoCode, err := format.Source(b.Bytes())
	if err != nil {
		err = remapErrorList(err, generatorOutput.SourceMap, fileName)
		return GenerateResult{}, nil, fmt.Errorf("%s source formatting error %w", fileName, err)
	}

	// Hash output, and write out the file if the goCodeHash has changed.
	goCodeHash := sha256.Sum256(formattedGoCode)
	if h.hashes.CompareAndSwap(targetFileName, syncmap.UpdateIfChanged, goCodeHash) {
		if err = h.writer(targetFileName, formattedGoCode); err != nil {
			return result, nil, fmt.Errorf("failed to write target file %q: %w", targetFileName, err)
		}
	}

	// Add the txt file if it has changed.
	if h.devMode {
		txtFileName := runtime.GetDevModeTextFileName(fileName)
		h.Log.Debug("Writing development mode text file", slog.String("file", fileName), slog.String("output", txtFileName))
		joined := strings.Join(generatorOutput.Literals, "\n")
		txtHash := sha256.Sum256([]byte(joined))
		if h.hashes.CompareAndSwap(txtFileName, syncmap.UpdateIfChanged, txtHash) {
			if err = os.WriteFile(txtFileName, []byte(joined), 0o644); err != nil {
				return result, nil, fmt.Errorf("failed to write string literal file %q: %w", txtFileName, err)
			}
		}
		// Check whether the change would require a recompilation or text update to take effect.
		previous, hasPrevious := h.fileNameToOutput.Get(fileName)
		if hasPrevious {
			result.TemplFileTextUpdated = generator.HasTextChanged(previous, generatorOutput)
			result.TemplFileGoUpdated = generator.HasGoChanged(previous, generatorOutput)
		}
		h.fileNameToOutput.Set(fileName, generatorOutput)
	}

	parsedDiagnostics, err := parser.Diagnose(t)
	if err != nil {
		return result, nil, fmt.Errorf("%s diagnostics error: %w", fileName, err)
	}

	if h.genSourceMapVis {
		err = generateSourceMapVisualisation(ctx, fileName, targetFileName, generatorOutput.SourceMap)
	}

	return result, parsedDiagnostics, err
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
	var grp errgroup.Group
	grp.Go(func() (err error) {
		templContents, err = os.ReadFile(templFileName)
		return err
	})
	grp.Go(func() (err error) {
		goContents, err = os.ReadFile(goFileName)
		return err
	})
	if err := grp.Wait(); err != nil {
		return err
	}
	component := visualize.HTML(templFileName, string(templContents), string(goContents), sourceMap)

	targetFileName := strings.TrimSuffix(templFileName, ".templ") + "_templ_sourcemap.html"
	w, err := os.Create(targetFileName)
	if err != nil {
		return fmt.Errorf("%s sourcemap visualisation error: %w", templFileName, err)
	}
	b := bufio.NewWriter(w)
	if err = component.Render(ctx, b); err != nil {
		_ = w.Close()
		return fmt.Errorf("%s sourcemap visualisation render error: %w", templFileName, err)
	}
	if err = b.Flush(); err != nil {
		_ = w.Close()
		return fmt.Errorf("%s sourcemap visualisation flush error: %w", templFileName, err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("%s sourcemap visualisation close error: %w", templFileName, err)
	}
	return nil
}
