package lspcmd

import "github.com/sourcegraph/go-lsp"

var htmlSnippets = []lsp.CompletionItem{
	{
		Label:            "a",
		InsertText:       `a href="${1:}">{%= ${2:""} %}</a>`,
		Kind:             lsp.CompletionItemKind(lsp.CIKSnippet),
		InsertTextFormat: lsp.ITFSnippet,
	},
	{
		Label: "div",
		InsertText: `div>
	${0}
</div>`,
		Kind:             lsp.CompletionItemKind(lsp.CIKSnippet),
		InsertTextFormat: lsp.ITFSnippet,
	},
}

var templateSnippets = []lsp.CompletionItem{
	{
		Label:            "%= string",
		InsertText:       `= ${1:string} %}`,
		Kind:             lsp.CompletionItemKind(lsp.CIKSnippet),
		InsertTextFormat: lsp.ITFSnippet,
	},
	{
		Label:            "%! template",
		InsertText:       `! ${1:template} %}`,
		Kind:             lsp.CompletionItemKind(lsp.CIKSnippet),
		InsertTextFormat: lsp.ITFSnippet,
	},
	{
		Label: "% templ",
		InsertText: `% templ ${1:name}(${2}) %}
	$0
{% endtempl %}`,
		Kind:             lsp.CompletionItemKind(lsp.CIKSnippet),
		InsertTextFormat: lsp.ITFSnippet,
	},
	{
		Label: "% css",
		InsertText: `% css ${1:name}(${2}) %}
	${3}: ${4};$0
{% endcss %}`,
		Kind:             lsp.CompletionItemKind(lsp.CIKSnippet),
		InsertTextFormat: lsp.ITFSnippet,
	},
	{
		Label: "% if",
		InsertText: `% if ${1:true} %}
	$0
{% endif %}`,
		Kind:             lsp.CompletionItemKind(lsp.CIKSnippet),
		InsertTextFormat: lsp.ITFSnippet,
	},
	{
		Label: "% for",
		InsertText: ` for ${1} %}
	$0
{% endfor %}`,
		Kind:             lsp.CompletionItemKind(lsp.CIKSnippet),
		InsertTextFormat: lsp.ITFSnippet,
	},
	{
		Label: "% switch",
		InsertText: ` switch ${1} %}
	case ${2}:
		$0
	{% endcase %}
{% endswitch %}`,
		Kind:             lsp.CompletionItemKind(lsp.CIKSnippet),
		InsertTextFormat: lsp.ITFSnippet,
	},
}
