package proxy

import (
	"log/slog"
	"path"
	"strings"

	lsp "github.com/a-h/templ/lsp/protocol"
)

func convertTemplToGoURI(templURI lsp.DocumentURI) (isTemplFile bool, goURI lsp.DocumentURI) {
	base, fileName := path.Split(string(templURI))
	if !strings.HasSuffix(fileName, ".templ") {
		return
	}
	return true, lsp.DocumentURI(base + (strings.TrimSuffix(fileName, ".templ") + "_templ.go"))
}

// isNonTemplGoURI returns true if the URI refers to a .go file that is not a _templ.go file.
func isNonTemplGoURI(docURI lsp.DocumentURI) bool {
	s := string(docURI)
	return strings.HasSuffix(s, ".go") && !strings.HasSuffix(s, "_templ.go")
}

func convertTemplGoToTemplURI(goURI lsp.DocumentURI) (isTemplGoFile bool, templURI lsp.DocumentURI) {
	base, fileName := path.Split(string(goURI))
	if !strings.HasSuffix(fileName, "_templ.go") {
		return
	}
	return true, lsp.DocumentURI(base + (strings.TrimSuffix(fileName, "_templ.go") + ".templ"))
}

// convertGoRangeToTemplRange converts a Go range to a templ range using the source map cache.
func convertGoRangeToTemplRange(cache *SourceMapCache, log *slog.Logger, templURI lsp.DocumentURI, input lsp.Range) (output lsp.Range) {
	output = input
	sourceMap, ok := cache.Get(string(templURI))
	if !ok {
		log.Warn("go->templ: sourcemap not found in cache",
			slog.String("uri", string(templURI)),
			slog.Any("cachedURIs", cache.URIs()))
		return
	}
	start, startMapped := sourceMap.SourcePositionFromTarget(input.Start.Line, input.Start.Character)
	if startMapped {
		output.Start.Line = start.Line
		output.Start.Character = start.Col
	}
	end, endMapped := sourceMap.SourcePositionFromTarget(input.End.Line, input.End.Character)
	if endMapped {
		output.End.Line = end.Line
		output.End.Character = end.Col
	}
	if !startMapped || !endMapped {
		log.Warn("go->templ: range not found in sourcemap",
			slog.String("uri", string(templURI)),
			slog.Any("range", input),
			slog.Bool("startMapped", startMapped),
			slog.Bool("endMapped", endMapped))
	}
	return
}

// convertLocationResults converts _templ.go URIs and ranges in location results back to .templ URIs.
func convertLocationResults(cache *SourceMapCache, log *slog.Logger, result []lsp.Location) {
	for i, r := range result {
		isTemplGoFile, templURI := convertTemplGoToTemplURI(r.URI)
		if !isTemplGoFile {
			continue
		}
		log.Info("convertLocationResults: converting",
			slog.String("from", string(r.URI)),
			slog.String("to", string(templURI)),
			slog.Any("goRange", r.Range))
		result[i].URI = templURI
		result[i].Range = convertGoRangeToTemplRange(cache, log, templURI, r.Range)
		log.Info("convertLocationResults: converted",
			slog.Any("templRange", result[i].Range))
	}
}

// convertCallHierarchyItem converts a _templ.go call hierarchy item back to .templ.
func convertCallHierarchyItem(cache *SourceMapCache, log *slog.Logger, item *lsp.CallHierarchyItem) {
	isTemplGoFile, templURI := convertTemplGoToTemplURI(item.URI)
	if !isTemplGoFile {
		return
	}
	item.URI = templURI
	item.Range = convertGoRangeToTemplRange(cache, log, templURI, item.Range)
	item.SelectionRange = convertGoRangeToTemplRange(cache, log, templURI, item.SelectionRange)
}

// convertWorkspaceEdit converts _templ.go URIs and ranges in a workspace edit back to .templ.
func convertWorkspaceEdit(cache *SourceMapCache, log *slog.Logger, edit *lsp.WorkspaceEdit) {
	if edit == nil {
		return
	}
	for i, dc := range edit.DocumentChanges {
		isTemplGoFile, templURI := convertTemplGoToTemplURI(dc.TextDocument.URI)
		if !isTemplGoFile {
			continue
		}
		for j, e := range dc.Edits {
			dc.Edits[j].Range = convertGoRangeToTemplRange(cache, log, templURI, e.Range)
		}
		dc.TextDocument.URI = templURI
		edit.DocumentChanges[i] = dc
	}
	if edit.Changes == nil {
		return
	}
	converted := make(map[lsp.DocumentURI][]lsp.TextEdit)
	for docURI, edits := range edit.Changes {
		isTemplGoFile, templURI := convertTemplGoToTemplURI(docURI)
		if !isTemplGoFile {
			converted[docURI] = edits
			continue
		}
		for i, e := range edits {
			edits[i].Range = convertGoRangeToTemplRange(cache, log, templURI, e.Range)
		}
		converted[templURI] = edits
	}
	edit.Changes = converted
}
