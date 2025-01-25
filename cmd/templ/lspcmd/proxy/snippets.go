package proxy

import lsp "github.com/a-h/templ/lsp/protocol"

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
		Label:            "button",
		InsertText:       `button type="button" ${1:}>${2:}</button>`,
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
	{
		Label: "p",
		InsertText: `p>
    ${0}
</p>`,
		Kind:             lsp.CompletionItemKind(lsp.CompletionItemKindSnippet),
		InsertTextFormat: lsp.InsertTextFormatSnippet,
	},
	{
		Label: "head",
		InsertText: `head>
    ${0}
</head>`,
		Kind:             lsp.CompletionItemKind(lsp.CompletionItemKindSnippet),
		InsertTextFormat: lsp.InsertTextFormatSnippet,
	},
	{
		Label: "body",
		InsertText: `body>
    ${0}
</body>`,
		Kind:             lsp.CompletionItemKind(lsp.CompletionItemKindSnippet),
		InsertTextFormat: lsp.InsertTextFormatSnippet,
	},
	{
		Label:            "title",
		InsertText:       `title>${0}</title>`,
		Kind:             lsp.CompletionItemKind(lsp.CompletionItemKindSnippet),
		InsertTextFormat: lsp.InsertTextFormatSnippet,
	},
	{
		Label:            "h1",
		InsertText:       `h1>${0}</h1>`,
		Kind:             lsp.CompletionItemKind(lsp.CompletionItemKindSnippet),
		InsertTextFormat: lsp.InsertTextFormatSnippet,
	},
	{
		Label:            "h2",
		InsertText:       `h2>${0}</h2>`,
		Kind:             lsp.CompletionItemKind(lsp.CompletionItemKindSnippet),
		InsertTextFormat: lsp.InsertTextFormatSnippet,
	},
	{
		Label:            "h3",
		InsertText:       `h3>${0}</h3>`,
		Kind:             lsp.CompletionItemKind(lsp.CompletionItemKindSnippet),
		InsertTextFormat: lsp.InsertTextFormatSnippet,
	},
	{
		Label:            "h4",
		InsertText:       `h4>${0}</h4>`,
		Kind:             lsp.CompletionItemKind(lsp.CompletionItemKindSnippet),
		InsertTextFormat: lsp.InsertTextFormatSnippet,
	},
	{
		Label:            "h5",
		InsertText:       `h5>${0}</h5>`,
		Kind:             lsp.CompletionItemKind(lsp.CompletionItemKindSnippet),
		InsertTextFormat: lsp.InsertTextFormatSnippet,
	},
	{
		Label:            "h6",
		InsertText:       `h6>${0}</h6>`,
		Kind:             lsp.CompletionItemKind(lsp.CompletionItemKindSnippet),
		InsertTextFormat: lsp.InsertTextFormatSnippet,
	},
}

var snippet = []lsp.CompletionItem{
	{
		Label: "templ",
		InsertText: `templ ${2:TemplateName}() {
	${0}
}`,
		Kind:             lsp.CompletionItemKind(lsp.CompletionItemKindSnippet),
		InsertTextFormat: lsp.InsertTextFormatSnippet,
	},
}
