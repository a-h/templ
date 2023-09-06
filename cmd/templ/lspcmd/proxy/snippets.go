package proxy

import lsp "github.com/a-h/protocol"

var htmlSnippets = []lsp.CompletionItem{
	{
		Label: "<?>",
		InsertText: `${1}>
	${0}
</${1}>`,
		Kind:             lsp.CompletionItemKind(lsp.CompletionItemKindSnippet),
		InsertTextFormat: lsp.InsertTextFormatSnippet,
	},
	{
		Label:            "a",
		InsertText:       `a href="${1:}">${2:}</a>`,
		Kind:             lsp.CompletionItemKind(lsp.CompletionItemKindSnippet),
		InsertTextFormat: lsp.InsertTextFormatSnippet,
	},
	{
		Label: "div",
		InsertText: `div>
	${0}
</div>`,
		Kind:             lsp.CompletionItemKind(lsp.CompletionItemKindSnippet),
		InsertTextFormat: lsp.InsertTextFormatSnippet,
	},
}
