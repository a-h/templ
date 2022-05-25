package proxy

import lsp "go.lsp.dev/protocol"

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
		InsertText:       `a href="${1:}">{%= ${2:""} %}</a>`,
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

var templateSnippets = []lsp.CompletionItem{
	{
		Label:            "{ string }",
		InsertText:       `{ ${1:string} }`,
		Kind:             lsp.CompletionItemKind(lsp.CompletionItemKindSnippet),
		InsertTextFormat: lsp.InsertTextFormatSnippet,
	},
	{
		Label:            "{! template",
		InsertText:       `{! ${1:template} }`,
		Kind:             lsp.CompletionItemKind(lsp.CompletionItemKindSnippet),
		InsertTextFormat: lsp.InsertTextFormatSnippet,
	},
	{
		Label: "templ",
		InsertText: `templ ${1:name}(${2}) {
	$0
}`,
		Kind:             lsp.CompletionItemKind(lsp.CompletionItemKindSnippet),
		InsertTextFormat: lsp.InsertTextFormatSnippet,
	},
	{
		Label: "css",
		InsertText: `css ${1:name}(${2}) {
	$0
}`,
		Kind:             lsp.CompletionItemKind(lsp.CompletionItemKindSnippet),
		InsertTextFormat: lsp.InsertTextFormatSnippet,
	},
}
