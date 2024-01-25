package proxy

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/a-h/parse"
	lsp "github.com/a-h/protocol"
	"github.com/a-h/templ"
	"github.com/a-h/templ/generator"
	"github.com/a-h/templ/parser/v2"
	"go.lsp.dev/uri"
	"go.uber.org/zap"
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
	Log             *zap.Logger
	Client          lsp.Client
	Target          lsp.Server
	SourceMapCache  *SourceMapCache
	DiagnosticCache *DiagnosticCache
	TemplSource     *DocumentContents
	GoSource        map[string]string
}

func NewServer(log *zap.Logger, target lsp.Server, cache *SourceMapCache, diagnosticCache *DiagnosticCache) (s *Server, init func(lsp.Client)) {
	s = &Server{
		Log:             log,
		Target:          target,
		SourceMapCache:  cache,
		DiagnosticCache: diagnosticCache,
		TemplSource:     newDocumentContents(log),
		GoSource:        make(map[string]string),
	}
	return s, func(client lsp.Client) {
		s.Client = client
	}
}

// updatePosition maps positions and filenames from source templ files into the target *.go files.
func (p *Server) updatePosition(templURI lsp.DocumentURI, current lsp.Position) (ok bool, goURI lsp.DocumentURI, updated lsp.Position) {
	log := p.Log.With(zap.String("uri", string(templURI)))
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
		log.Info("updatePosition: not found", zap.String("from", fmt.Sprintf("%d:%d", current.Line, current.Character)))
		return false, templURI, current
	}
	log.Info("updatePosition: found", zap.String("fromTempl", fmt.Sprintf("%d:%d", current.Line, current.Character)),
		zap.String("toGo", fmt.Sprintf("%d:%d", to.Line, to.Col)))
	updated.Line = to.Line
	updated.Character = to.Col
	return true, goURI, updated
}

func (p *Server) convertTemplRangeToGoRange(templURI lsp.DocumentURI, input lsp.Range) (output lsp.Range) {
	output = input
	sourceMap, ok := p.SourceMapCache.Get(string(templURI))
	if !ok {
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
		return
	}
	// Map from the source position to target Go position.
	start, ok := sourceMap.SourcePositionFromTarget(input.Start.Line, input.Start.Character)
	if ok {
		output.Start.Line = start.Line
		output.Start.Character = start.Col
	}
	end, ok := sourceMap.SourcePositionFromTarget(input.End.Line, input.End.Character)
	if ok {
		output.End.Line = end.Line
		output.End.Character = end.Col
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
		err = p.Client.PublishDiagnostics(ctx, msg)
		if err != nil {
			p.Log.Error("failed to publish error diagnostics", zap.Error(err))
		}
		return
	}
	ok = true
	if len(template.Diagnostics) > 0 {
		msg := &lsp.PublishDiagnosticsParams{
			URI: uri,
		}
		for _, d := range template.Diagnostics {
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
		err = p.Client.PublishDiagnostics(ctx, msg)
		if err != nil {
			p.Log.Error("failed to publish error diagnostics", zap.Error(err))
		}
		return
	}
	// Clear templ diagnostics.
	p.DiagnosticCache.ClearTemplDiagnostics(string(uri))
	err = p.Client.PublishDiagnostics(ctx, &lsp.PublishDiagnosticsParams{
		URI: uri,
		// Cannot be nil as per https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#publishDiagnosticsParams
		Diagnostics: []lsp.Diagnostic{},
	})
	if err != nil {
		p.Log.Error("failed to publish diagnostics", zap.Error(err))
		return
	}
	return
}

func (p *Server) Initialize(ctx context.Context, params *lsp.InitializeParams) (result *lsp.InitializeResult, err error) {
	p.Log.Info("client -> server: Initialize")
	defer p.Log.Info("client -> server: Initialize end")
	result, err = p.Target.Initialize(ctx, params)
	if err != nil {
		p.Log.Error("Initialize failed", zap.Error(err))
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

	result.ServerInfo.Name = "templ-lsp"
	result.ServerInfo.Version = templ.Version()

	return result, err
}

func (p *Server) Initialized(ctx context.Context, params *lsp.InitializedParams) (err error) {
	p.Log.Info("client -> server: Initialized")
	defer p.Log.Info("client -> server: Initialized end")
	return p.Target.Initialized(ctx, params)
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
	p.Log.Info("client -> server: LogTrace", zap.String("message", params.Message))
	defer p.Log.Info("client -> server: LogTrace end")
	return p.Target.LogTrace(ctx, params)
}

func (p *Server) SetTrace(ctx context.Context, params *lsp.SetTraceParams) (err error) {
	p.Log.Info("client -> server: SetTrace")
	defer p.Log.Info("client -> server: SetTrace end")
	return p.Target.SetTrace(ctx, params)
}

func (p *Server) CodeAction(ctx context.Context, params *lsp.CodeActionParams) (result []lsp.CodeAction, err error) {
	p.Log.Info("client -> server: CodeAction")
	defer p.Log.Info("client -> server: CodeAction end")
	isTemplFile, goURI := convertTemplToGoURI(params.TextDocument.URI)
	if !isTemplFile {
		return p.Target.CodeAction(ctx, params)
	}
	templURI := params.TextDocument.URI
	params.TextDocument.URI = goURI
	result, err = p.Target.CodeAction(ctx, params)
	if err != nil {
		return
	}
	for i := 0; i < len(result); i++ {
		r := result[i]
		// Rewrite the Diagnostics range field.
		for di := 0; di < len(r.Diagnostics); di++ {
			r.Diagnostics[di].Range = p.convertGoRangeToTemplRange(templURI, r.Diagnostics[di].Range)
		}
		// Rewrite the DocumentChanges.
		for dci := 0; dci < len(r.Edit.DocumentChanges); dci++ {
			dc := r.Edit.DocumentChanges[0]
			for ei := 0; ei < len(dc.Edits); ei++ {
				dc.Edits[ei].Range = p.convertGoRangeToTemplRange(templURI, dc.Edits[ei].Range)
			}
			dc.TextDocument.URI = templURI
		}
		result[i] = r
	}
	return
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
	for i := 0; i < len(result); i++ {
		cl := result[i]
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
	for i := 0; i < len(result); i++ {
		r := result[i]
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
	// Call the target.
	result, err = p.Target.Completion(ctx, params)
	if err != nil {
		p.Log.Warn("completion: got gopls error", zap.Error(err))
		return
	}
	if result == nil {
		return
	}
	// Rewrite the result positions.
	p.Log.Info("completion: received items", zap.Int("count", len(result.Items)))
	for i := 0; i < len(result.Items); i++ {
		item := result.Items[i]
		if item.TextEdit != nil {
			item.TextEdit.Range = p.convertGoRangeToTemplRange(templURI, item.TextEdit.Range)
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
	LineIndex int
	Text      string
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
	for i := 0; i < len(result); i++ {
		if isTemplGoFile, templURI := convertTemplGoToTemplURI(result[i].URI); isTemplGoFile {
			result[i].URI = templURI
			result[i].Range = p.convertGoRangeToTemplRange(templURI, result[i].Range)
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
	for i := 0; i < len(result); i++ {
		if isTemplGoFile, templURI := convertTemplGoToTemplURI(result[i].URI); isTemplGoFile {
			result[i].URI = templURI
			result[i].Range = p.convertGoRangeToTemplRange(templURI, result[i].Range)
		}
	}
	return
}

func (p *Server) DidChange(ctx context.Context, params *lsp.DidChangeTextDocumentParams) (err error) {
	p.Log.Info("client -> server: DidChange", zap.Any("params", params))
	defer p.Log.Info("client -> server: DidChange end")
	isTemplFile, goURI := convertTemplToGoURI(params.TextDocument.URI)
	if !isTemplFile {
		p.Log.Error("not a templ file")
		return
	}
	// Apply content changes to the cached template.
	d, err := p.TemplSource.Apply(string(params.TextDocument.URI), params.ContentChanges)
	if err != nil {
		p.Log.Error("error applying changes", zap.Error(err))
		return
	}
	// Update the Go code.
	p.Log.Info("parsing template")
	template, ok, err := p.parseTemplate(ctx, params.TextDocument.URI, d.String())
	if err != nil {
		p.Log.Error("parseTemplate failure", zap.Error(err))
	}
	if !ok {
		return
	}
	w := new(strings.Builder)
	sm, _, err := generator.Generate(template, w)
	if err != nil {
		p.Log.Error("generate failure", zap.Error(err))
		return
	}
	// Cache the sourcemap.
	p.Log.Info("setting cache", zap.String("uri", string(params.TextDocument.URI)))
	p.SourceMapCache.Set(string(params.TextDocument.URI), sm)
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
	p.Log.Info("client -> server: DidOpen", zap.String("uri", string(params.TextDocument.URI)))
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
		p.Log.Error("parseTemplate failure", zap.Error(err))
	}
	if !ok {
		p.Log.Info("parsing template did not succeed", zap.String("uri", string(params.TextDocument.URI)))
		return nil
	}
	// Generate the output code and cache the source map and Go contents to use during completion
	// requests.
	w := new(strings.Builder)
	sm, _, err := generator.Generate(template, w)
	if err != nil {
		return
	}
	p.Log.Info("setting source map cache contents", zap.String("uri", string(params.TextDocument.URI)))
	p.SourceMapCache.Set(string(params.TextDocument.URI), sm)
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
	for i := 0; i < len(result); i++ {
		result[i].Range = p.convertGoRangeToTemplRange(templURI, result[i].Range)
	}
	return
}

func (p *Server) DocumentHighlight(ctx context.Context, params *lsp.DocumentHighlightParams) (result []lsp.DocumentHighlight, err error) {
	p.Log.Info("client -> server: DocumentHighlight")
	defer p.Log.Info("client -> server: DocumentHighlight end")
	return
}

func (p *Server) DocumentLink(ctx context.Context, params *lsp.DocumentLinkParams) (result []lsp.DocumentLink, err error) {
	p.Log.Info("client -> server: DocumentLink", zap.String("uri", string(params.TextDocument.URI)))
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
	params.Range = p.convertTemplRangeToGoRange(templURI, params.Range)
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

func (p *Server) DocumentSymbol(ctx context.Context, params *lsp.DocumentSymbolParams) (result []interface{} /* []SymbolInformation | []DocumentSymbol */, err error) {
	p.Log.Info("client -> server: DocumentSymbol")
	defer p.Log.Info("client -> server: DocumentSymbol end")
	// TODO: Rewrite the request and response, but for now, ignore it.
	// return p.Target.DocumentSymbol(ctx params)
	return
}

func (p *Server) ExecuteCommand(ctx context.Context, params *lsp.ExecuteCommandParams) (result interface{}, err error) {
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
		p.Log.Error("parseTemplate failure", zap.Error(err))
	}
	if !ok {
		return
	}
	w := new(strings.Builder)
	err = template.Write(w)
	if err != nil {
		p.Log.Error("handleFormatting: faled to write template", zap.Error(err))
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
	result, err = p.Target.Hover(ctx, params)
	if err != nil {
		return
	}
	// Rewrite the response.
	if result != nil && result.Range != nil {
		p.Log.Info("hover: result returned")
		r := p.convertGoRangeToTemplRange(templURI, *result.Range)
		p.Log.Info("hover: setting range", zap.Any("range", r))
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
	for i := 0; i < len(result); i++ {
		r := result[i]
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
	for i := 0; i < len(result); i++ {
		r := result[i]
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
	for i := 0; i < len(result); i++ {
		r := result[i]
		r.Range = p.convertGoRangeToTemplRange(templURI, r.Range)
		result[i] = r
	}
	return result, err
}

func (p *Server) References(ctx context.Context, params *lsp.ReferenceParams) (result []lsp.Location, err error) {
	p.Log.Info("client -> server: References")
	defer p.Log.Info("client -> server: References end")
	templURI := params.TextDocument.URI
	// Rewrite the request.
	var isTemplURI bool
	isTemplURI, params.TextDocument.URI = convertTemplToGoURI(params.TextDocument.URI)
	if !isTemplURI {
		err = fmt.Errorf("not a templ file")
		return
	}
	// Call gopls.
	result, err = p.Target.References(ctx, params)
	if err != nil {
		return
	}
	// Rewrite the response.
	for i := 0; i < len(result); i++ {
		r := result[i]
		r.URI = templURI
		r.Range = p.convertGoRangeToTemplRange(templURI, r.Range)
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

func (p *Server) SemanticTokensFullDelta(ctx context.Context, params *lsp.SemanticTokensDeltaParams) (result interface{} /* SemanticTokens | SemanticTokensDelta */, err error) {
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

func (p *Server) Request(ctx context.Context, method string, params interface{}) (result interface{}, err error) {
	p.Log.Info("client -> server: Request")
	defer p.Log.Info("client -> server: Request end")
	return p.Target.Request(ctx, method, params)
}
