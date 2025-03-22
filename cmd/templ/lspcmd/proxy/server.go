package proxy

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/a-h/parse"
	lsp "github.com/a-h/templ/lsp/protocol"
	"github.com/a-h/templ/lsp/uri"

	"github.com/a-h/templ"
	"github.com/a-h/templ/cmd/templ/imports"
	"github.com/a-h/templ/generator"
	"github.com/a-h/templ/parser/v2"
)

// Server is responsible for rewriting messages that are
// originated from the text editor, and need to be sent to gopls.
//
// Since the editor is working on `templ` files, and `gopls` works
// on Go files, the job of this code is to rewrite incoming requests
// to adjust the file names from `*.templ` to `*_templ.go` and to
// remap the line/character positions in the `templ` files to their
// corresponding locations in the Go file.
//
// This allows gopls to operate as usual.
//
// This code also rewrites the responses back from gopls to do the
// inverse operation - to put the file names back, and readjust any
// character positions.
type Server struct {
	Log             *slog.Logger
	Target          lsp.Server
	SourceMapCache  *SourceMapCache
	DiagnosticCache *DiagnosticCache
	TemplSource     *DocumentContents
	GoSource        map[string]string
	preLoadURIs     []*lsp.DidOpenTextDocumentParams
}

func NewServer(log *slog.Logger, target lsp.Server, cache *SourceMapCache, diagnosticCache *DiagnosticCache) (s *Server) {
	return &Server{
		Log:             log,
		Target:          target,
		SourceMapCache:  cache,
		DiagnosticCache: diagnosticCache,
		TemplSource:     newDocumentContents(log),
		GoSource:        make(map[string]string),
	}
}

// updatePosition maps positions and filenames from source templ files into the target *.go files.
func (p *Server) updatePosition(templURI lsp.DocumentURI, current lsp.Position) (ok bool, goURI lsp.DocumentURI, updated lsp.Position) {
	log := p.Log.With(slog.String("uri", string(templURI)))
	var isTemplFile bool
	if isTemplFile, goURI = convertTemplToGoURI(templURI); !isTemplFile {
		return false, templURI, current
	}
	sourceMap, ok := p.SourceMapCache.Get(string(templURI))
	if !ok {
		log.Warn("completion: sourcemap not found in cache, it could be that didOpen was not called")
		return
	}
	// Map from the source position to target Go position.
	to, ok := sourceMap.TargetPositionFromSource(current.Line, current.Character)
	if !ok {
		log.Info("updatePosition: not found", slog.String("from", fmt.Sprintf("%d:%d", current.Line, current.Character)))
		return false, templURI, current
	}
	log.Info("updatePosition: found", slog.String("fromTempl", fmt.Sprintf("%d:%d", current.Line, current.Character)),
		slog.String("toGo", fmt.Sprintf("%d:%d", to.Line, to.Col)))
	updated.Line = to.Line
	updated.Character = to.Col

	return true, goURI, updated
}

func (p *Server) convertTemplRangeToGoRange(templURI lsp.DocumentURI, input lsp.Range) (output lsp.Range, ok bool) {
	output = input
	var sourceMap *parser.SourceMap
	sourceMap, ok = p.SourceMapCache.Get(string(templURI))
	if !ok {
		p.Log.Warn("templ->go: sourcemap not found in cache")
		return
	}
	// Map from the source position to target Go position.
	start, ok := sourceMap.TargetPositionFromSource(input.Start.Line, input.Start.Character)
	if ok {
		output.Start.Line = start.Line
		output.Start.Character = start.Col
	}
	end, ok := sourceMap.TargetPositionFromSource(input.End.Line, input.End.Character)
	if ok {
		output.End.Line = end.Line
		output.End.Character = end.Col
	}
	return
}

func (p *Server) convertGoRangeToTemplRange(templURI lsp.DocumentURI, input lsp.Range) (output lsp.Range) {
	output = input
	sourceMap, ok := p.SourceMapCache.Get(string(templURI))
	if !ok {
		p.Log.Warn("go->templ: sourcemap not found in cache")
		return
	}
	// Map from the source position to target Go position.
	start, startPositionMapped := sourceMap.SourcePositionFromTarget(input.Start.Line, input.Start.Character)
	if startPositionMapped {
		output.Start.Line = start.Line
		output.Start.Character = start.Col
	}
	end, endPositionMapped := sourceMap.SourcePositionFromTarget(input.End.Line, input.End.Character)
	if endPositionMapped {
		output.End.Line = end.Line
		output.End.Character = end.Col
	}
	if !startPositionMapped || !endPositionMapped {
		p.Log.Warn("go->templ: range not found in sourcemap", slog.Any("range", input))
	}
	return
}

// parseTemplate parses the templ file content, and notifies the end user via the LSP about how it went.
func (p *Server) parseTemplate(ctx context.Context, uri uri.URI, templateText string) (template parser.TemplateFile, ok bool, err error) {
	template, err = parser.ParseString(templateText)
	if err != nil {
		msg := &lsp.PublishDiagnosticsParams{
			URI: uri,
			Diagnostics: []lsp.Diagnostic{
				{
					Severity: lsp.DiagnosticSeverityError,
					Code:     "",
					Source:   "templ",
					Message:  err.Error(),
				},
			},
		}
		if pe, isParserError := err.(parse.ParseError); isParserError {
			msg.Diagnostics[0].Range = lsp.Range{
				Start: lsp.Position{
					Line:      uint32(pe.Pos.Line),
					Character: uint32(pe.Pos.Col),
				},
				End: lsp.Position{
					Line:      uint32(pe.Pos.Line),
					Character: uint32(pe.Pos.Col),
				},
			}
		}
		msg.Diagnostics = p.DiagnosticCache.AddGoDiagnostics(string(uri), msg.Diagnostics)
		err = lsp.ClientFromContext(ctx).PublishDiagnostics(ctx, msg)
		if err != nil {
			p.Log.Error("failed to publish error diagnostics", slog.Any("error", err))
		}
		return
	}
	template.Filepath = string(uri)
	parsedDiagnostics, err := parser.Diagnose(template)
	if err != nil {
		return
	}
	ok = true
	if len(parsedDiagnostics) > 0 {
		msg := &lsp.PublishDiagnosticsParams{
			URI: uri,
		}
		for _, d := range parsedDiagnostics {
			msg.Diagnostics = append(msg.Diagnostics, lsp.Diagnostic{
				Severity: lsp.DiagnosticSeverityWarning,
				Code:     "",
				Source:   "templ",
				Message:  d.Message,
				Range: lsp.Range{
					Start: lsp.Position{
						Line:      uint32(d.Range.From.Line),
						Character: uint32(d.Range.From.Col),
					},
					End: lsp.Position{
						Line:      uint32(d.Range.To.Line),
						Character: uint32(d.Range.To.Col),
					},
				},
			})
		}
		msg.Diagnostics = p.DiagnosticCache.AddGoDiagnostics(string(uri), msg.Diagnostics)
		err = lsp.ClientFromContext(ctx).PublishDiagnostics(ctx, msg)
		if err != nil {
			p.Log.Error("failed to publish error diagnostics", slog.Any("error", err))
		}
		return
	}
	// Clear templ diagnostics.
	p.DiagnosticCache.ClearTemplDiagnostics(string(uri))
	err = lsp.ClientFromContext(ctx).PublishDiagnostics(ctx, &lsp.PublishDiagnosticsParams{
		URI: uri,
		// Cannot be nil as per https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#publishDiagnosticsParams
		Diagnostics: []lsp.Diagnostic{},
	})
	if err != nil {
		p.Log.Error("failed to publish diagnostics", slog.Any("error", err))
		return
	}
	return
}

func (p *Server) Initialize(ctx context.Context, params *lsp.InitializeParams) (result *lsp.InitializeResult, err error) {
	p.Log.Info("client -> server: Initialize")
	defer p.Log.Info("client -> server: Initialize end")
	result, err = p.Target.Initialize(ctx, params)
	if err != nil {
		p.Log.Error("Initialize failed", slog.Any("error", err))
	}
	// Add the '<' and '{' trigger so that we can do snippets for tags.
	if result.Capabilities.CompletionProvider == nil {
		result.Capabilities.CompletionProvider = &lsp.CompletionOptions{}
	}
	result.Capabilities.CompletionProvider.TriggerCharacters = append(result.Capabilities.CompletionProvider.TriggerCharacters, "{", "<")
	// Remove all the gopls commands.
	if result.Capabilities.ExecuteCommandProvider == nil {
		result.Capabilities.ExecuteCommandProvider = &lsp.ExecuteCommandOptions{}
	}
	result.Capabilities.ExecuteCommandProvider.Commands = []string{}
	result.Capabilities.DocumentFormattingProvider = true
	result.Capabilities.SemanticTokensProvider = nil
	result.Capabilities.DocumentRangeFormattingProvider = false
	result.Capabilities.TextDocumentSync = lsp.TextDocumentSyncOptions{
		OpenClose:         true,
		Change:            lsp.TextDocumentSyncKindFull,
		WillSave:          false,
		WillSaveWaitUntil: false,
		Save:              &lsp.SaveOptions{IncludeText: true},
	}

	for _, c := range params.WorkspaceFolders {
		path := strings.TrimPrefix(c.URI, "file://")
		werr := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			p.Log.Info("found file", slog.String("path", path))
			uri := uri.URI("file://" + path)
			isTemplFile, goURI := convertTemplToGoURI(uri)

			if !isTemplFile {
				return nil
			}

			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			p.TemplSource.Set(string(uri), NewDocument(p.Log, string(b)))
			// Parse the template.
			template, ok, err := p.parseTemplate(ctx, uri, string(b))
			if err != nil {
				p.Log.Error("parseTemplate failure", slog.Any("error", err))
			}
			if !ok {
				p.Log.Info("parsing template did not succeed", slog.String("uri", string(uri)))
				return nil
			}
			w := new(strings.Builder)
			generatorOutput, err := generator.Generate(template, w)
			if err != nil {
				return fmt.Errorf("generate failure: %w", err)
			}
			p.Log.Info("setting source map cache contents", slog.String("uri", string(uri)))
			p.SourceMapCache.Set(string(uri), generatorOutput.SourceMap)
			// Set the Go contents.
			p.GoSource[string(uri)] = w.String()

			didOpenParams := &lsp.DidOpenTextDocumentParams{
				TextDocument: lsp.TextDocumentItem{
					URI:        goURI,
					Text:       w.String(),
					Version:    1,
					LanguageID: "go",
				},
			}

			p.preLoadURIs = append(p.preLoadURIs, didOpenParams)
			return nil
		})
		if werr != nil {
			p.Log.Error("walk error", slog.Any("error", werr))
		}
	}

	result.ServerInfo.Name = "templ-lsp"
	result.ServerInfo.Version = templ.Version()

	return result, err
}

func (p *Server) Initialized(ctx context.Context, params *lsp.InitializedParams) (err error) {
	p.Log.Info("client -> server: Initialized")
	defer p.Log.Info("client -> server: Initialized end")
	goInitErr := p.Target.Initialized(ctx, params)

	for i, doParams := range p.preLoadURIs {
		doErr := p.Target.DidOpen(ctx, doParams)
		if doErr != nil {
			return doErr
		}
		p.preLoadURIs[i] = nil
	}

	return goInitErr
}

func (p *Server) Shutdown(ctx context.Context) (err error) {
	p.Log.Info("client -> server: Shutdown")
	defer p.Log.Info("client -> server: Shutdown end")
	return p.Target.Shutdown(ctx)
}

func (p *Server) Exit(ctx context.Context) (err error) {
	p.Log.Info("client -> server: Exit")
	defer p.Log.Info("client -> server: Exit end")
	return p.Target.Exit(ctx)
}

func (p *Server) WorkDoneProgressCancel(ctx context.Context, params *lsp.WorkDoneProgressCancelParams) (err error) {
	p.Log.Info("client -> server: WorkDoneProgressCancel")
	defer p.Log.Info("client -> server: WorkDoneProgressCancel end")
	return p.Target.WorkDoneProgressCancel(ctx, params)
}

func (p *Server) LogTrace(ctx context.Context, params *lsp.LogTraceParams) (err error) {
	p.Log.Info("client -> server: LogTrace", slog.String("message", params.Message))
	defer p.Log.Info("client -> server: LogTrace end")
	return p.Target.LogTrace(ctx, params)
}

func (p *Server) SetTrace(ctx context.Context, params *lsp.SetTraceParams) (err error) {
	p.Log.Info("client -> server: SetTrace")
	defer p.Log.Info("client -> server: SetTrace end")
	return p.Target.SetTrace(ctx, params)
}

var supportedCodeActions = map[string]bool{}

func (p *Server) CodeAction(ctx context.Context, params *lsp.CodeActionParams) (result []lsp.CodeAction, err error) {
	p.Log.Info("client -> server: CodeAction", slog.Any("params", params))
	defer p.Log.Info("client -> server: CodeAction end")
	isTemplFile, goURI := convertTemplToGoURI(params.TextDocument.URI)
	if !isTemplFile {
		return p.Target.CodeAction(ctx, params)
	}
	templURI := params.TextDocument.URI
	var ok bool
	if params.Range, ok = p.convertTemplRangeToGoRange(templURI, params.Range); !ok {
		// Don't pass the request to gopls if the range is not within a Go code block.
		return
	}
	params.TextDocument.URI = goURI
	result, err = p.Target.CodeAction(ctx, params)
	if err != nil {
		return
	}
	var updatedResults []lsp.CodeAction
	// Filter out commands that are not yet supported.
	// For example, "Fill Struct" runs the `gopls.apply_fix` command.
	// This command has a set of arguments, including Fix, Range and URI.
	// However, these are just a map[string]any so for each command that we want to support,
	// we need to know what the arguments are so that we can rewrite them.
	for _, r := range result {
		if !supportedCodeActions[r.Title] {
			continue
		}
		// Rewrite the Diagnostics range field.
		for di, diag := range r.Diagnostics {
			r.Diagnostics[di].Range = p.convertGoRangeToTemplRange(templURI, diag.Range)
		}
		// Rewrite the DocumentChanges.
		if r.Edit != nil {
			for dci, dc := range r.Edit.DocumentChanges {
				for ei, edit := range dc.Edits {
					dc.Edits[ei].Range = p.convertGoRangeToTemplRange(templURI, edit.Range)
				}
				dc.TextDocument.URI = templURI
				r.Edit.DocumentChanges[dci] = dc
			}
		}
		updatedResults = append(updatedResults, r)
	}
	return updatedResults, nil
}

func (p *Server) CodeLens(ctx context.Context, params *lsp.CodeLensParams) (result []lsp.CodeLens, err error) {
	p.Log.Info("client -> server: CodeLens")
	defer p.Log.Info("client -> server: CodeLens end")
	isTemplFile, goURI := convertTemplToGoURI(params.TextDocument.URI)
	if !isTemplFile {
		return p.Target.CodeLens(ctx, params)
	}
	templURI := params.TextDocument.URI
	params.TextDocument.URI = goURI
	result, err = p.Target.CodeLens(ctx, params)
	if err != nil {
		return
	}
	if result == nil {
		return
	}
	for i, cl := range result {
		cl.Range = p.convertGoRangeToTemplRange(templURI, cl.Range)
		result[i] = cl
	}
	return
}

func (p *Server) CodeLensResolve(ctx context.Context, params *lsp.CodeLens) (result *lsp.CodeLens, err error) {
	p.Log.Info("client -> server: CodeLensResolve")
	defer p.Log.Info("client -> server: CodeLensResolve end")
	return p.Target.CodeLensResolve(ctx, params)
}

func (p *Server) ColorPresentation(ctx context.Context, params *lsp.ColorPresentationParams) (result []lsp.ColorPresentation, err error) {
	p.Log.Info("client -> server: ColorPresentation ColorPresentation")
	defer p.Log.Info("client -> server: ColorPresentation end")
	isTemplFile, goURI := convertTemplToGoURI(params.TextDocument.URI)
	if !isTemplFile {
		return p.Target.ColorPresentation(ctx, params)
	}
	templURI := params.TextDocument.URI
	params.TextDocument.URI = goURI
	result, err = p.Target.ColorPresentation(ctx, params)
	if err != nil {
		return
	}
	if result == nil {
		return
	}
	for i, r := range result {
		if r.TextEdit != nil {
			r.TextEdit.Range = p.convertGoRangeToTemplRange(templURI, r.TextEdit.Range)
		}
		result[i] = r
	}
	return
}

func (p *Server) Completion(ctx context.Context, params *lsp.CompletionParams) (result *lsp.CompletionList, err error) {
	p.Log.Info("client -> server: Completion")
	defer p.Log.Info("client -> server: Completion end")
	if params.Context != nil && params.Context.TriggerCharacter == "<" {
		result = &lsp.CompletionList{
			Items: htmlSnippets,
		}
		return
	}
	// Get the sourcemap from the cache.
	templURI := params.TextDocument.URI
	var ok bool
	ok, params.TextDocument.URI, params.TextDocumentPositionParams.Position = p.updatePosition(templURI, params.TextDocumentPositionParams.Position)
	if !ok {
		return nil, nil
	}

	// Ensure that Go source is available.
	gosrc := strings.Split(p.GoSource[string(templURI)], "\n")
	if len(gosrc) < int(params.TextDocumentPositionParams.Position.Line) {
		p.Log.Info("completion: line position out of range")
		return nil, nil
	}
	if len(gosrc[params.TextDocumentPositionParams.Position.Line]) < int(params.TextDocumentPositionParams.Position.Character) {
		p.Log.Info("completion: col position out of range")
		return nil, nil
	}

	// Call the target.
	result, err = p.Target.Completion(ctx, params)
	if err != nil {
		p.Log.Warn("completion: got gopls error", slog.Any("error", err))
		return
	}
	if result == nil {
		return
	}
	// Rewrite the result positions.
	p.Log.Info("completion: received items", slog.Int("count", len(result.Items)))

	for i, item := range result.Items {
		if item.TextEdit != nil {
			if item.TextEdit.TextEdit != nil {
				item.TextEdit.TextEdit.Range = p.convertGoRangeToTemplRange(templURI, item.TextEdit.TextEdit.Range)
			}
			if item.TextEdit.InsertReplaceEdit != nil {
				item.TextEdit.InsertReplaceEdit.Insert = p.convertGoRangeToTemplRange(templURI, item.TextEdit.InsertReplaceEdit.Insert)
				item.TextEdit.InsertReplaceEdit.Replace = p.convertGoRangeToTemplRange(templURI, item.TextEdit.InsertReplaceEdit.Replace)
			}
		}
		if len(item.AdditionalTextEdits) > 0 {
			doc, ok := p.TemplSource.Get(string(templURI))
			if !ok {
				continue
			}
			pkg := getPackageFromItemDetail(item.Detail)
			imp := addImport(doc.Lines, pkg)
			item.AdditionalTextEdits = []lsp.TextEdit{
				{
					Range: lsp.Range{
						Start: lsp.Position{Line: uint32(imp.LineIndex), Character: 0},
						End:   lsp.Position{Line: uint32(imp.LineIndex), Character: 0},
					},
					NewText: imp.Text,
				},
			}
		}
		result.Items[i] = item
	}

	// Add templ snippet.
	result.Items = append(result.Items, snippet...)

	return
}

var completionWithImport = regexp.MustCompile(`^.*\(from\s(".+")\)$`)

func getPackageFromItemDetail(pkg string) string {
	if m := completionWithImport.FindStringSubmatch(pkg); len(m) == 2 {
		return m[1]
	}
	return pkg
}

type importInsert struct {
	Text      string
	LineIndex int
}

var nonImportKeywordRegexp = regexp.MustCompile(`^(?:templ|func|css|script|var|const|type)\s`)

func addImport(lines []string, pkg string) (result importInsert) {
	var isInMultiLineImport bool
	lastSingleLineImportIndex := -1
	for lineIndex, line := range lines {
		if strings.HasPrefix(line, "import (") {
			isInMultiLineImport = true
			continue
		}
		if strings.HasPrefix(line, "import \"") {
			lastSingleLineImportIndex = lineIndex
			continue
		}
		if isInMultiLineImport && strings.HasPrefix(line, ")") {
			return importInsert{
				LineIndex: lineIndex,
				Text:      fmt.Sprintf("\t%s\n", pkg),
			}
		}
		// Only add import statements before templates, functions, css, and script templates.
		if nonImportKeywordRegexp.MatchString(line) {
			break
		}
	}
	var suffix string
	if lastSingleLineImportIndex == -1 {
		lastSingleLineImportIndex = 1
		suffix = "\n"
	}
	return importInsert{
		LineIndex: lastSingleLineImportIndex + 1,
		Text:      fmt.Sprintf("import %s\n%s", pkg, suffix),
	}
}

func (p *Server) CompletionResolve(ctx context.Context, params *lsp.CompletionItem) (result *lsp.CompletionItem, err error) {
	p.Log.Info("client -> server: CompletionResolve")
	defer p.Log.Info("client -> server: CompletionResolve end")
	return p.Target.CompletionResolve(ctx, params)
}

func (p *Server) Declaration(ctx context.Context, params *lsp.DeclarationParams) (result []lsp.Location /* Declaration | DeclarationLink[] | null */, err error) {
	p.Log.Info("client -> server: Declaration")
	defer p.Log.Info("client -> server: Declaration end")
	// Rewrite the request.
	templURI := params.TextDocument.URI
	var ok bool
	ok, params.TextDocument.URI, params.Position = p.updatePosition(templURI, params.Position)
	if !ok {
		return nil, nil
	}
	// Call gopls and get the result.
	result, err = p.Target.Declaration(ctx, params)
	if err != nil {
		return
	}
	if result == nil {
		return
	}
	for i, r := range result {
		if isTemplGoFile, templURI := convertTemplGoToTemplURI(r.URI); isTemplGoFile {
			result[i].URI = templURI
			result[i].Range = p.convertGoRangeToTemplRange(templURI, r.Range)
		}
	}
	return
}

func (p *Server) Definition(ctx context.Context, params *lsp.DefinitionParams) (result []lsp.Location /* Definition | DefinitionLink[] | null */, err error) {
	p.Log.Info("client -> server: Definition")
	defer p.Log.Info("client -> server: Definition end")
	// Rewrite the request.
	templURI := params.TextDocument.URI
	var ok bool
	ok, params.TextDocument.URI, params.Position = p.updatePosition(templURI, params.Position)
	if !ok {
		return result, nil
	}
	// Call gopls and get the result.
	result, err = p.Target.Definition(ctx, params)
	if err != nil {
		return
	}
	if result == nil {
		return
	}
	for i, r := range result {
		if isTemplGoFile, templURI := convertTemplGoToTemplURI(r.URI); isTemplGoFile {
			result[i].URI = templURI
			result[i].Range = p.convertGoRangeToTemplRange(templURI, r.Range)
		}
	}
	return
}

func (p *Server) DidChange(ctx context.Context, params *lsp.DidChangeTextDocumentParams) (err error) {
	p.Log.Info("client -> server: DidChange", slog.Any("params", params))
	defer p.Log.Info("client -> server: DidChange end")
	isTemplFile, goURI := convertTemplToGoURI(params.TextDocument.URI)
	if !isTemplFile {
		p.Log.Error("not a templ file")
		return
	}
	// Apply content changes to the cached template.
	d, err := p.TemplSource.Apply(string(params.TextDocument.URI), params.ContentChanges)
	if err != nil {
		p.Log.Error("error applying changes", slog.Any("error", err))
		return
	}
	// Update the Go code.
	p.Log.Info("parsing template")
	template, ok, err := p.parseTemplate(ctx, params.TextDocument.URI, d.String())
	if err != nil {
		p.Log.Error("parseTemplate failure", slog.Any("error", err))
	}
	if !ok {
		return
	}
	w := new(strings.Builder)
	// In future updates, we may pass `WithSkipCodeGeneratedComment` to the generator.
	// This will enable a number of actions within gopls that it doesn't currently apply because
	// it recognises templ code as being auto-generated.
	//
	// This change would increase the surface area of gopls that we use, so may surface a number of issues
	// if enabled.
	generatorOutput, err := generator.Generate(template, w)
	if err != nil {
		p.Log.Error("generate failure", slog.Any("error", err))
		return
	}
	// Cache the sourcemap.
	p.Log.Info("setting cache", slog.String("uri", string(params.TextDocument.URI)))
	p.SourceMapCache.Set(string(params.TextDocument.URI), generatorOutput.SourceMap)
	p.GoSource[string(params.TextDocument.URI)] = w.String()
	// Change the path.
	params.TextDocument.URI = goURI
	params.TextDocument.TextDocumentIdentifier.URI = goURI
	// Overwrite all the Go contents.
	params.ContentChanges = []lsp.TextDocumentContentChangeEvent{{
		Text: w.String(),
	}}
	return p.Target.DidChange(ctx, params)
}

func (p *Server) DidChangeConfiguration(ctx context.Context, params *lsp.DidChangeConfigurationParams) (err error) {
	p.Log.Info("client -> server: DidChangeConfiguration")
	defer p.Log.Info("client -> server: DidChangeConfiguration end")
	return p.Target.DidChangeConfiguration(ctx, params)
}

func (p *Server) DidChangeWatchedFiles(ctx context.Context, params *lsp.DidChangeWatchedFilesParams) (err error) {
	p.Log.Info("client -> server: DidChangeWatchedFiles")
	defer p.Log.Info("client -> server: DidChangeWatchedFiles end")
	return p.Target.DidChangeWatchedFiles(ctx, params)
}

func (p *Server) DidChangeWorkspaceFolders(ctx context.Context, params *lsp.DidChangeWorkspaceFoldersParams) (err error) {
	p.Log.Info("client -> server: DidChangeWorkspaceFolders")
	defer p.Log.Info("client -> server: DidChangeWorkspaceFolders end")
	return p.Target.DidChangeWorkspaceFolders(ctx, params)
}

func (p *Server) DidClose(ctx context.Context, params *lsp.DidCloseTextDocumentParams) (err error) {
	p.Log.Info("client -> server: DidClose")
	defer p.Log.Info("client -> server: DidClose end")
	isTemplFile, goURI := convertTemplToGoURI(params.TextDocument.URI)
	if !isTemplFile {
		return p.Target.DidClose(ctx, params)
	}
	// Delete the template and sourcemaps from caches.
	p.TemplSource.Delete(string(params.TextDocument.URI))
	p.SourceMapCache.Delete(string(params.TextDocument.URI))
	// Get gopls to delete the Go file from its cache.
	params.TextDocument.URI = goURI
	return p.Target.DidClose(ctx, params)
}

func (p *Server) DidOpen(ctx context.Context, params *lsp.DidOpenTextDocumentParams) (err error) {
	p.Log.Info("client -> server: DidOpen", slog.String("uri", string(params.TextDocument.URI)))
	defer p.Log.Info("client -> server: DidOpen end")
	isTemplFile, goURI := convertTemplToGoURI(params.TextDocument.URI)
	if !isTemplFile {
		return p.Target.DidOpen(ctx, params)
	}
	// Cache the template doc.
	p.TemplSource.Set(string(params.TextDocument.URI), NewDocument(p.Log, params.TextDocument.Text))
	// Parse the template.
	template, ok, err := p.parseTemplate(ctx, params.TextDocument.URI, params.TextDocument.Text)
	if err != nil {
		p.Log.Error("parseTemplate failure", slog.Any("error", err))
	}
	if !ok {
		p.Log.Info("parsing template did not succeed", slog.String("uri", string(params.TextDocument.URI)))
		return nil
	}
	// Generate the output code and cache the source map and Go contents to use during completion
	// requests.
	w := new(strings.Builder)
	generatorOutput, err := generator.Generate(template, w)
	if err != nil {
		return
	}
	p.Log.Info("setting source map cache contents", slog.String("uri", string(params.TextDocument.URI)))
	p.SourceMapCache.Set(string(params.TextDocument.URI), generatorOutput.SourceMap)
	// Set the Go contents.
	params.TextDocument.Text = w.String()
	p.GoSource[string(params.TextDocument.URI)] = params.TextDocument.Text
	// Change the path.
	params.TextDocument.URI = goURI
	return p.Target.DidOpen(ctx, params)
}

func (p *Server) DidSave(ctx context.Context, params *lsp.DidSaveTextDocumentParams) (err error) {
	p.Log.Info("client -> server: DidSave")
	defer p.Log.Info("client -> server: DidSave end")
	if isTemplFile, goURI := convertTemplToGoURI(params.TextDocument.URI); isTemplFile {
		params.TextDocument.URI = goURI
	}
	return p.Target.DidSave(ctx, params)
}

func (p *Server) DocumentColor(ctx context.Context, params *lsp.DocumentColorParams) (result []lsp.ColorInformation, err error) {
	p.Log.Info("client -> server: DocumentColor")
	defer p.Log.Info("client -> server: DocumentColor end")
	isTemplFile, goURI := convertTemplToGoURI(params.TextDocument.URI)
	if !isTemplFile {
		return p.Target.DocumentColor(ctx, params)
	}
	templURI := params.TextDocument.URI
	params.TextDocument.URI = goURI
	result, err = p.Target.DocumentColor(ctx, params)
	if err != nil {
		return
	}
	if result == nil {
		return
	}
	for i, r := range result {
		result[i].Range = p.convertGoRangeToTemplRange(templURI, r.Range)
	}
	return
}

func (p *Server) DocumentHighlight(ctx context.Context, params *lsp.DocumentHighlightParams) (result []lsp.DocumentHighlight, err error) {
	p.Log.Info("client -> server: DocumentHighlight")
	defer p.Log.Info("client -> server: DocumentHighlight end")
	return
}

func (p *Server) DocumentLink(ctx context.Context, params *lsp.DocumentLinkParams) (result []lsp.DocumentLink, err error) {
	p.Log.Info("client -> server: DocumentLink", slog.String("uri", string(params.TextDocument.URI)))
	defer p.Log.Info("client -> server: DocumentLink end")
	return
}

func (p *Server) DocumentLinkResolve(ctx context.Context, params *lsp.DocumentLink) (result *lsp.DocumentLink, err error) {
	p.Log.Info("client -> server: DocumentLinkResolve")
	defer p.Log.Info("client -> server: DocumentLinkResolve end")
	isTemplFile, goURI := convertTemplToGoURI(params.Target)
	if !isTemplFile {
		return p.Target.DocumentLinkResolve(ctx, params)
	}
	templURI := params.Target
	params.Target = goURI
	var ok bool
	if params.Range, ok = p.convertTemplRangeToGoRange(templURI, params.Range); !ok {
		return
	}
	// Rewrite the result.
	result, err = p.Target.DocumentLinkResolve(ctx, params)
	if err != nil {
		return
	}
	if result == nil {
		return
	}
	result.Target = templURI
	result.Range = p.convertGoRangeToTemplRange(templURI, result.Range)
	return
}

func (p *Server) DocumentSymbol(ctx context.Context, params *lsp.DocumentSymbolParams) (result []lsp.SymbolInformationOrDocumentSymbol, err error) {
	p.Log.Info("client -> server: DocumentSymbol")
	defer p.Log.Info("client -> server: DocumentSymbol end")
	isTemplFile, goURI := convertTemplToGoURI(params.TextDocument.URI)
	if !isTemplFile {
		return p.Target.DocumentSymbol(ctx, params)
	}
	templURI := params.TextDocument.URI
	params.TextDocument.URI = goURI
	symbols, err := p.Target.DocumentSymbol(ctx, params)
	if err != nil {
		return nil, err
	}

	for _, s := range symbols {
		if s.DocumentSymbol != nil {
			p.convertSymbolRange(templURI, s.DocumentSymbol)
			result = append(result, s)
		}
		if s.SymbolInformation != nil {
			s.SymbolInformation.Location.URI = templURI
			s.SymbolInformation.Location.Range = p.convertGoRangeToTemplRange(templURI, s.SymbolInformation.Location.Range)
			result = append(result, s)
		}
	}

	return result, err
}

func (p *Server) convertSymbolRange(templURI lsp.DocumentURI, s *lsp.DocumentSymbol) {
	sourceMap, ok := p.SourceMapCache.Get(string(templURI))
	if !ok {
		p.Log.Warn("go->templ: sourcemap not found in cache")
		return
	}
	src, ok := sourceMap.SymbolSourceRangeFromTarget(s.Range.Start.Line, s.Range.Start.Character)
	if !ok {
		p.Log.Warn("go->templ: symbol range not found", slog.Any("symbol", s), slog.Any("choices", sourceMap.TargetSymbolRangeToSource))
		return
	}
	s.Range = lsp.Range{
		Start: lsp.Position{
			Line:      uint32(src.From.Line),
			Character: uint32(src.From.Col),
		},
		End: lsp.Position{
			Line:      uint32(src.To.Line),
			Character: uint32(src.To.Col),
		},
	}
	// Within the symbol, we can select sub-sections.
	// These are Go expressions, in the standard source map.
	s.SelectionRange = p.convertGoRangeToTemplRange(templURI, s.SelectionRange)
	for i := range s.Children {
		p.convertSymbolRange(templURI, &s.Children[i])
		if !isRangeWithin(s.Range, s.Children[i].Range) {
			p.Log.Error("child symbol range not within parent range", slog.Any("symbol", s.Children[i]), slog.Int("index", i))
		}
	}
	if !isRangeWithin(s.Range, s.SelectionRange) {
		p.Log.Error("selection range not within range", slog.Any("symbol", s))
	}
}

func isRangeWithin(parent, child lsp.Range) bool {
	if child.Start.Line < parent.Start.Line || child.End.Line > parent.End.Line {
		return false
	}
	if child.Start.Line == parent.Start.Line && child.Start.Character < parent.Start.Character {
		return false
	}
	if child.End.Line == parent.End.Line && child.End.Character > parent.End.Character {
		return false
	}
	return true
}

func (p *Server) ExecuteCommand(ctx context.Context, params *lsp.ExecuteCommandParams) (result any, err error) {
	p.Log.Info("client -> server: ExecuteCommand")
	defer p.Log.Info("client -> server: ExecuteCommand end")
	return p.Target.ExecuteCommand(ctx, params)
}

func (p *Server) FoldingRanges(ctx context.Context, params *lsp.FoldingRangeParams) (result []lsp.FoldingRange, err error) {
	p.Log.Info("client -> server: FoldingRanges")
	defer p.Log.Info("client -> server: FoldingRanges end")
	// There are no folding ranges in templ files.
	// return p.Target.FoldingRanges(ctx, params)
	return []lsp.FoldingRange{}, nil
}

func (p *Server) Formatting(ctx context.Context, params *lsp.DocumentFormattingParams) (result []lsp.TextEdit, err error) {
	p.Log.Info("client -> server: Formatting")
	defer p.Log.Info("client -> server: Formatting end")
	// Format the current document.
	d, _ := p.TemplSource.Get(string(params.TextDocument.URI))
	template, ok, err := p.parseTemplate(ctx, params.TextDocument.URI, d.String())
	if err != nil {
		p.Log.Error("parseTemplate failure", slog.Any("error", err))
		return
	}
	if !ok {
		return
	}
	p.Log.Info("attempting to organise imports", slog.String("uri", template.Filepath))
	template, err = imports.Process(template)
	if err != nil {
		p.Log.Error("organise imports failure", slog.Any("error", err))
		return
	}
	w := new(strings.Builder)
	err = template.Write(w)
	if err != nil {
		p.Log.Error("handleFormatting: faled to write template", slog.Any("error", err))
		return
	}
	// Replace everything.
	result = append(result, lsp.TextEdit{
		Range: lsp.Range{
			Start: lsp.Position{},
			End:   lsp.Position{Line: uint32(len(d.Lines)), Character: 0},
		},
		NewText: w.String(),
	})
	d.Replace(w.String())
	return
}

func (p *Server) Hover(ctx context.Context, params *lsp.HoverParams) (result *lsp.Hover, err error) {
	p.Log.Info("client -> server: Hover")
	defer p.Log.Info("client -> server: Hover end")
	// Rewrite the request.
	templURI := params.TextDocument.URI
	var ok bool
	ok, params.TextDocument.URI, params.Position = p.updatePosition(params.TextDocument.URI, params.Position)
	if !ok {
		return nil, nil
	}
	// Call gopls.
	result, err = p.Target.Hover(ctx, params)
	if err != nil {
		return
	}
	// Rewrite the response.
	if result != nil && result.Range != nil {
		p.Log.Info("hover: result returned")
		r := p.convertGoRangeToTemplRange(templURI, *result.Range)
		p.Log.Info("hover: setting range")
		result.Range = &r
	}
	return
}

func (p *Server) Implementation(ctx context.Context, params *lsp.ImplementationParams) (result []lsp.Location, err error) {
	p.Log.Info("client -> server: Implementation")
	defer p.Log.Info("client -> server: Implementation end")
	templURI := params.TextDocument.URI
	// Rewrite the request.
	var ok bool
	ok, params.TextDocument.URI, params.Position = p.updatePosition(params.TextDocument.URI, params.Position)
	if !ok {
		return nil, nil
	}
	result, err = p.Target.Implementation(ctx, params)
	if err != nil {
		return
	}
	if result == nil {
		return
	}
	// Rewrite the response.
	for i, r := range result {
		r.URI = templURI
		r.Range = p.convertGoRangeToTemplRange(templURI, r.Range)
		result[i] = r
	}
	return
}

func (p *Server) OnTypeFormatting(ctx context.Context, params *lsp.DocumentOnTypeFormattingParams) (result []lsp.TextEdit, err error) {
	p.Log.Info("client -> server: OnTypeFormatting")
	defer p.Log.Info("client -> server: OnTypeFormatting end")
	templURI := params.TextDocument.URI
	// Rewrite the request.
	var ok bool
	ok, params.TextDocument.URI, params.Position = p.updatePosition(params.TextDocument.URI, params.Position)
	if !ok {
		return nil, nil
	}
	// Get the response.
	result, err = p.Target.OnTypeFormatting(ctx, params)
	if err != nil {
		return
	}
	if result == nil {
		return
	}
	// Rewrite the response.
	for i, r := range result {
		r.Range = p.convertGoRangeToTemplRange(templURI, r.Range)
		result[i] = r
	}
	return
}

func (p *Server) PrepareRename(ctx context.Context, params *lsp.PrepareRenameParams) (result *lsp.Range, err error) {
	p.Log.Info("client -> server: PrepareRename")
	defer p.Log.Info("client -> server: PrepareRename end")
	templURI := params.TextDocument.URI
	// Rewrite the request.
	var ok bool
	ok, params.TextDocument.URI, params.Position = p.updatePosition(params.TextDocument.URI, params.Position)
	if !ok {
		return nil, nil
	}
	// Get the response.
	result, err = p.Target.PrepareRename(ctx, params)
	if err != nil {
		return
	}
	if result == nil {
		return
	}
	// Rewrite the response.
	output := p.convertGoRangeToTemplRange(templURI, *result)
	return &output, nil
}

func (p *Server) RangeFormatting(ctx context.Context, params *lsp.DocumentRangeFormattingParams) (result []lsp.TextEdit, err error) {
	p.Log.Info("client -> server: RangeFormatting")
	defer p.Log.Info("client -> server: RangeFormatting end")
	templURI := params.TextDocument.URI
	// Rewrite the request.
	var isTemplURI bool
	isTemplURI, params.TextDocument.URI = convertTemplToGoURI(params.TextDocument.URI)
	if !isTemplURI {
		err = fmt.Errorf("not a templ file")
		return
	}
	// Call gopls.
	result, err = p.Target.RangeFormatting(ctx, params)
	if err != nil {
		return
	}
	// Rewrite the response.
	for i, r := range result {
		r.Range = p.convertGoRangeToTemplRange(templURI, r.Range)
		result[i] = r
	}
	return result, err
}

func (p *Server) References(ctx context.Context, params *lsp.ReferenceParams) (result []lsp.Location, err error) {
	p.Log.Info("client -> server: References")
	defer p.Log.Info("client -> server: References end")
	// Rewrite the request.
	var ok bool
	ok, params.TextDocument.URI, params.Position = p.updatePosition(params.TextDocument.URI, params.Position)
	if !ok {
		return nil, nil
	}
	// Call gopls.
	result, err = p.Target.References(ctx, params)
	if err != nil {
		return
	}
	// Rewrite the response.
	for i, r := range result {
		isTemplURI, templURI := convertTemplGoToTemplURI(r.URI)
		if isTemplURI {
			p.Log.Info(fmt.Sprintf("references-%d - range conversion for %s", i, r.URI))
			r.URI, r.Range = templURI, p.convertGoRangeToTemplRange(templURI, r.Range)
		}
		p.Log.Info(fmt.Sprintf("references-%d: %+v", i, r))
		result[i] = r
	}
	return result, err
}

func (p *Server) Rename(ctx context.Context, params *lsp.RenameParams) (result *lsp.WorkspaceEdit, err error) {
	p.Log.Info("client -> server: Rename")
	defer p.Log.Info("client -> server: Rename end")
	return p.Target.Rename(ctx, params)
}

func (p *Server) SignatureHelp(ctx context.Context, params *lsp.SignatureHelpParams) (result *lsp.SignatureHelp, err error) {
	p.Log.Info("client -> server: SignatureHelp")
	defer p.Log.Info("client -> server: SignatureHelp end")
	var ok bool
	ok, params.TextDocument.URI, params.Position = p.updatePosition(params.TextDocument.URI, params.Position)
	if !ok {
		return nil, nil
	}
	return p.Target.SignatureHelp(ctx, params)
}

func (p *Server) Symbols(ctx context.Context, params *lsp.WorkspaceSymbolParams) (result []lsp.SymbolInformation, err error) {
	p.Log.Info("client -> server: Symbols")
	defer p.Log.Info("client -> server: Symbols end")
	return p.Target.Symbols(ctx, params)
}

func (p *Server) TypeDefinition(ctx context.Context, params *lsp.TypeDefinitionParams) (result []lsp.Location, err error) {
	p.Log.Info("client -> server: TypeDefinition")
	defer p.Log.Info("client -> server: TypeDefinition end")
	var ok bool
	ok, params.TextDocument.URI, params.Position = p.updatePosition(params.TextDocument.URI, params.Position)
	if !ok {
		return nil, nil
	}
	return p.Target.TypeDefinition(ctx, params)
}

func (p *Server) WillSave(ctx context.Context, params *lsp.WillSaveTextDocumentParams) (err error) {
	p.Log.Info("client -> server: WillSave")
	defer p.Log.Info("client -> server: WillSave end")
	var ok bool
	ok, params.TextDocument.URI = convertTemplToGoURI(params.TextDocument.URI)
	if !ok {
		p.Log.Error("not a templ file")
		return nil
	}
	return p.Target.WillSave(ctx, params)
}

func (p *Server) WillSaveWaitUntil(ctx context.Context, params *lsp.WillSaveTextDocumentParams) (result []lsp.TextEdit, err error) {
	p.Log.Info("client -> server: WillSaveWaitUntil")
	defer p.Log.Info("client -> server: WillSaveWaitUntil end")
	return p.Target.WillSaveWaitUntil(ctx, params)
}

func (p *Server) ShowDocument(ctx context.Context, params *lsp.ShowDocumentParams) (result *lsp.ShowDocumentResult, err error) {
	p.Log.Info("client -> server: ShowDocument")
	defer p.Log.Info("client -> server: ShowDocument end")
	return p.Target.ShowDocument(ctx, params)
}

func (p *Server) WillCreateFiles(ctx context.Context, params *lsp.CreateFilesParams) (result *lsp.WorkspaceEdit, err error) {
	p.Log.Info("client -> server: WillCreateFiles")
	defer p.Log.Info("client -> server: WillCreateFiles end")
	return p.Target.WillCreateFiles(ctx, params)
}

func (p *Server) DidCreateFiles(ctx context.Context, params *lsp.CreateFilesParams) (err error) {
	p.Log.Info("client -> server: DidCreateFiles")
	defer p.Log.Info("client -> server: DidCreateFiles end")
	return p.Target.DidCreateFiles(ctx, params)
}

func (p *Server) WillRenameFiles(ctx context.Context, params *lsp.RenameFilesParams) (result *lsp.WorkspaceEdit, err error) {
	p.Log.Info("client -> server: WillRenameFiles")
	defer p.Log.Info("client -> server: WillRenameFiles end")
	return p.Target.WillRenameFiles(ctx, params)
}

func (p *Server) DidRenameFiles(ctx context.Context, params *lsp.RenameFilesParams) (err error) {
	p.Log.Info("client -> server: DidRenameFiles")
	defer p.Log.Info("client -> server: DidRenameFiles end")
	return p.Target.DidRenameFiles(ctx, params)
}

func (p *Server) WillDeleteFiles(ctx context.Context, params *lsp.DeleteFilesParams) (result *lsp.WorkspaceEdit, err error) {
	p.Log.Info("client -> server: WillDeleteFiles")
	defer p.Log.Info("client -> server: WillDeleteFiles end")
	return p.Target.WillDeleteFiles(ctx, params)
}

func (p *Server) DidDeleteFiles(ctx context.Context, params *lsp.DeleteFilesParams) (err error) {
	p.Log.Info("client -> server: DidDeleteFiles")
	defer p.Log.Info("client -> server: DidDeleteFiles end")
	return p.Target.DidDeleteFiles(ctx, params)
}

func (p *Server) CodeLensRefresh(ctx context.Context) (err error) {
	p.Log.Info("client -> server: CodeLensRefresh")
	defer p.Log.Info("client -> server: CodeLensRefresh end")
	return p.Target.CodeLensRefresh(ctx)
}

func (p *Server) PrepareCallHierarchy(ctx context.Context, params *lsp.CallHierarchyPrepareParams) (result []lsp.CallHierarchyItem, err error) {
	p.Log.Info("client -> server: PrepareCallHierarchy")
	defer p.Log.Info("client -> server: PrepareCallHierarchy end")
	return p.Target.PrepareCallHierarchy(ctx, params)
}

func (p *Server) IncomingCalls(ctx context.Context, params *lsp.CallHierarchyIncomingCallsParams) (result []lsp.CallHierarchyIncomingCall, err error) {
	p.Log.Info("client -> server: IncomingCalls")
	defer p.Log.Info("client -> server: IncomingCalls end")
	return p.Target.IncomingCalls(ctx, params)
}

func (p *Server) OutgoingCalls(ctx context.Context, params *lsp.CallHierarchyOutgoingCallsParams) (result []lsp.CallHierarchyOutgoingCall, err error) {
	p.Log.Info("client -> server: OutgoingCalls")
	defer p.Log.Info("client -> server: OutgoingCalls end")
	return p.Target.OutgoingCalls(ctx, params)
}

func (p *Server) SemanticTokensFull(ctx context.Context, params *lsp.SemanticTokensParams) (result *lsp.SemanticTokens, err error) {
	p.Log.Info("client -> server: SemanticTokensFull")
	defer p.Log.Info("client -> server: SemanticTokensFull end")
	isTemplFile, goURI := convertTemplToGoURI(params.TextDocument.URI)
	if !isTemplFile {
		return nil, nil
	}
	params.TextDocument.URI = goURI
	return p.Target.SemanticTokensFull(ctx, params)
}

func (p *Server) SemanticTokensFullDelta(ctx context.Context, params *lsp.SemanticTokensDeltaParams) (result any /* SemanticTokens | SemanticTokensDelta */, err error) {
	p.Log.Info("client -> server: SemanticTokensFullDelta")
	defer p.Log.Info("client -> server: SemanticTokensFullDelta end")
	isTemplFile, goURI := convertTemplToGoURI(params.TextDocument.URI)
	if !isTemplFile {
		return nil, nil
	}
	params.TextDocument.URI = goURI
	return p.Target.SemanticTokensFullDelta(ctx, params)
}

func (p *Server) SemanticTokensRange(ctx context.Context, params *lsp.SemanticTokensRangeParams) (result *lsp.SemanticTokens, err error) {
	p.Log.Info("client -> server: SemanticTokensRange")
	defer p.Log.Info("client -> server: SemanticTokensRange end")
	isTemplFile, goURI := convertTemplToGoURI(params.TextDocument.URI)
	if !isTemplFile {
		return nil, nil
	}
	params.TextDocument.URI = goURI
	return p.Target.SemanticTokensRange(ctx, params)
}

func (p *Server) SemanticTokensRefresh(ctx context.Context) (err error) {
	p.Log.Info("client -> server: SemanticTokensRefresh")
	defer p.Log.Info("client -> server: SemanticTokensRefresh end")
	return p.Target.SemanticTokensRefresh(ctx)
}

func (p *Server) LinkedEditingRange(ctx context.Context, params *lsp.LinkedEditingRangeParams) (result *lsp.LinkedEditingRanges, err error) {
	p.Log.Info("client -> server: LinkedEditingRange")
	defer p.Log.Info("client -> server: LinkedEditingRange end")
	return p.Target.LinkedEditingRange(ctx, params)
}

func (p *Server) Moniker(ctx context.Context, params *lsp.MonikerParams) (result []lsp.Moniker, err error) {
	p.Log.Info("client -> server: Moniker")
	defer p.Log.Info("client -> server: Moniker end")
	templURI := params.TextDocument.URI
	var ok bool
	ok, params.TextDocument.URI, params.TextDocumentPositionParams.Position = p.updatePosition(templURI, params.TextDocumentPositionParams.Position)
	if !ok {
		return nil, nil
	}
	return p.Target.Moniker(ctx, params)
}

func (p *Server) Request(ctx context.Context, method string, params any) (result any, err error) {
	p.Log.Info("client -> server: Request")
	defer p.Log.Info("client -> server: Request end")
	return p.Target.Request(ctx, method, params)
}
