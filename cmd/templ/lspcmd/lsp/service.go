package lsp

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"strings"
)

type None struct{}

type InitializeParams struct {
	ProcessID int `json:"processId,omitempty"`

	// RootPath is DEPRECATED in favor of the RootURI field.
	RootPath string `json:"rootPath,omitempty"`

	RootURI               DocumentURI        `json:"rootUri,omitempty"`
	ClientInfo            ClientInfo         `json:"clientInfo,omitempty"`
	Trace                 Trace              `json:"trace,omitempty"`
	InitializationOptions interface{}        `json:"initializationOptions,omitempty"`
	Capabilities          ClientCapabilities `json:"capabilities"`

	WorkDoneToken string `json:"workDoneToken,omitempty"`
}

// Root returns the RootURI if set, or otherwise the RootPath with 'file://' prepended.
func (p *InitializeParams) Root() DocumentURI {
	if p.RootURI != "" {
		return p.RootURI
	}
	if strings.HasPrefix(p.RootPath, "file://") {
		return DocumentURI(p.RootPath)
	}
	return DocumentURI("file://" + p.RootPath)
}

type DocumentURI string

type ClientInfo struct {
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}

type Trace string

type ClientCapabilities struct {
	Workspace    WorkspaceClientCapabilities    `json:"workspace,omitempty"`
	TextDocument TextDocumentClientCapabilities `json:"textDocument,omitempty"`
	Window       WindowClientCapabilities       `json:"window,omitempty"`
	Experimental interface{}                    `json:"experimental,omitempty"`

	// Below are Sourcegraph extensions. They do not live in lspext since
	// they are extending the field InitializeParams.Capabilities

	// XFilesProvider indicates the client provides support for
	// workspace/xfiles. This is a Sourcegraph extension.
	XFilesProvider bool `json:"xfilesProvider,omitempty"`

	// XContentProvider indicates the client provides support for
	// textDocument/xcontent. This is a Sourcegraph extension.
	XContentProvider bool `json:"xcontentProvider,omitempty"`

	// XCacheProvider indicates the client provides support for cache/get
	// and cache/set.
	XCacheProvider bool `json:"xcacheProvider,omitempty"`
}

type WorkspaceClientCapabilities struct {
	WorkspaceEdit struct {
		DocumentChanges    bool     `json:"documentChanges,omitempty"`
		ResourceOperations []string `json:"resourceOperations,omitempty"`
	} `json:"workspaceEdit,omitempty"`

	ApplyEdit bool `json:"applyEdit,omitempty"`

	Symbol struct {
		SymbolKind struct {
			ValueSet []int `json:"valueSet,omitempty"`
		} `json:"symbolKind,omitEmpty"`
	} `json:"symbol,omitempty"`

	ExecuteCommand *struct {
		DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	} `json:"executeCommand,omitempty"`

	DidChangeWatchedFiles *struct {
		DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	} `json:"didChangeWatchedFiles,omitempty"`

	WorkspaceFolders bool `json:"workspaceFolders,omitempty"`

	Configuration bool `json:"configuration,omitempty"`
}

type TextDocumentClientCapabilities struct {
	Declaration *struct {
		LinkSupport bool `json:"linkSupport,omitempty"`
	} `json:"declaration,omitempty"`

	Definition *struct {
		LinkSupport bool `json:"linkSupport,omitempty"`
	} `json:"definition,omitempty"`

	Implementation *struct {
		LinkSupport bool `json:"linkSupport,omitempty"`

		DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	} `json:"implementation,omitempty"`

	TypeDefinition *struct {
		LinkSupport bool `json:"linkSupport,omitempty"`
	} `json:"typeDefinition,omitempty"`

	Synchronization *struct {
		WillSave          bool `json:"willSave,omitempty"`
		DidSave           bool `json:"didSave,omitempty"`
		WillSaveWaitUntil bool `json:"willSaveWaitUntil,omitempty"`
	} `json:"synchronization,omitempty"`

	DocumentSymbol struct {
		SymbolKind struct {
			ValueSet []int `json:"valueSet,omitempty"`
		} `json:"symbolKind,omitEmpty"`

		HierarchicalDocumentSymbolSupport bool `json:"hierarchicalDocumentSymbolSupport,omitempty"`
	} `json:"documentSymbol,omitempty"`

	Formatting *struct {
		DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	} `json:"formatting,omitempty"`

	RangeFormatting *struct {
		DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	} `json:"rangeFormatting,omitempty"`

	Rename *struct {
		DynamicRegistration bool `json:"dynamicRegistration,omitempty"`

		PrepareSupport bool `json:"prepareSupport,omitempty"`
	} `json:"rename,omitempty"`

	SemanticHighlightingCapabilities *struct {
		SemanticHighlighting bool `json:"semanticHighlighting,omitempty"`
	} `json:"semanticHighlightingCapabilities,omitempty"`

	CodeAction struct {
		DynamicRegistration bool `json:"dynamicRegistration,omitempty"`

		IsPreferredSupport bool `json:"isPreferredSupport,omitempty"`

		CodeActionLiteralSupport struct {
			CodeActionKind struct {
				ValueSet []CodeActionKind `json:"valueSet,omitempty"`
			} `json:"codeActionKind,omitempty"`
		} `json:"codeActionLiteralSupport,omitempty"`
	} `json:"codeAction,omitempty"`

	Completion struct {
		CompletionItem struct {
			DocumentationFormat []DocumentationFormat `json:"documentationFormat,omitempty"`
			SnippetSupport      bool                  `json:"snippetSupport,omitempty"`
		} `json:"completionItem,omitempty"`

		CompletionItemKind struct {
			ValueSet []CompletionItemKind `json:"valueSet,omitempty"`
		} `json:"completionItemKind,omitempty"`

		ContextSupport bool `json:"contextSupport,omitempty"`
	} `json:"completion,omitempty"`

	SignatureHelp *struct {
		SignatureInformation struct {
			ParameterInformation struct {
				LabelOffsetSupport bool `json:"labelOffsetSupport,omitempty"`
			} `json:"parameterInformation,omitempty"`
		} `json:"signatureInformation,omitempty"`
	} `json:"signatureHelp,omitempty"`

	DocumentLink *struct {
		DynamicRegistration bool `json:"dynamicRegistration,omitempty"`

		TooltipSupport bool `json:"tooltipSupport,omitempty"`
	} `json:"documentLink,omitempty"`

	Hover *struct {
		ContentFormat []string `json:"contentFormat,omitempty"`
	} `json:"hover,omitempty"`

	FoldingRange *struct {
		DynamicRegistration bool `json:"dynamicRegistration,omitempty"`

		RangeLimit interface{} `json:"rangeLimit,omitempty"`

		LineFoldingOnly bool `json:"lineFoldingOnly,omitempty"`
	} `json:"foldingRange,omitempty"`

	CallHierarchy *struct {
		DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	} `json:"callHierarchy,omitempty"`

	ColorProvider *struct {
		DynamicRegistration bool `json:"dynamicRegistration,omitempty"`
	} `json:"colorProvider,omitempty"`
}

type WindowClientCapabilities struct {
	WorkDoneProgress bool `json:"workDoneProgress,omitempty"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities,omitempty"`
}

type InitializeError struct {
	Retry bool `json:"retry"`
}

type ResourceOperation string

const (
	ROCreate ResourceOperation = "create"
	RODelete ResourceOperation = "delete"
	RORename ResourceOperation = "rename"
)

// TextDocumentSyncKind is a DEPRECATED way to describe how text
// document syncing works. Use TextDocumentSyncOptions instead (or the
// Options field of TextDocumentSyncOptionsOrKind if you need to
// support JSON-(un)marshaling both).
type TextDocumentSyncKind int

const (
	TDSKNone        TextDocumentSyncKind = 0
	TDSKFull        TextDocumentSyncKind = 1
	TDSKIncremental TextDocumentSyncKind = 2
)

type TextDocumentSyncOptions struct {
	OpenClose         bool                 `json:"openClose,omitempty"`
	Change            TextDocumentSyncKind `json:"change"`
	WillSave          bool                 `json:"willSave,omitempty"`
	WillSaveWaitUntil bool                 `json:"willSaveWaitUntil,omitempty"`
	Save              *SaveOptions         `json:"save,omitempty"`
}

// TextDocumentSyncOptions holds either a TextDocumentSyncKind or
// TextDocumentSyncOptions. The LSP API allows either to be specified
// in the (ServerCapabilities).TextDocumentSync field.
type TextDocumentSyncOptionsOrKind struct {
	Kind    *TextDocumentSyncKind
	Options *TextDocumentSyncOptions
}

// MarshalJSON implements json.Marshaler.
func (v *TextDocumentSyncOptionsOrKind) MarshalJSON() ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}
	if v.Kind != nil {
		return json.Marshal(v.Kind)
	}
	return json.Marshal(v.Options)
}

// UnmarshalJSON implements json.Unmarshaler.
func (v *TextDocumentSyncOptionsOrKind) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		*v = TextDocumentSyncOptionsOrKind{}
		return nil
	}
	var kind TextDocumentSyncKind
	if err := json.Unmarshal(data, &kind); err == nil {
		// Create equivalent TextDocumentSyncOptions using the same
		// logic as in vscode-languageclient. Also set the Kind field
		// so that JSON-marshaling and unmarshaling are inverse
		// operations (for backward compatibility, preserving the
		// original input but accepting both).
		*v = TextDocumentSyncOptionsOrKind{
			Options: &TextDocumentSyncOptions{OpenClose: true, Change: kind},
			Kind:    &kind,
		}
		return nil
	}
	var tmp TextDocumentSyncOptions
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	*v = TextDocumentSyncOptionsOrKind{Options: &tmp}
	return nil
}

type SaveOptions struct {
	IncludeText bool `json:"includeText"`
}

type ServerCapabilities struct {
	TextDocumentSync                 *TextDocumentSyncOptionsOrKind   `json:"textDocumentSync,omitempty"`
	HoverProvider                    bool                             `json:"hoverProvider,omitempty"`
	CompletionProvider               *CompletionOptions               `json:"completionProvider,omitempty"`
	SignatureHelpProvider            *SignatureHelpOptions            `json:"signatureHelpProvider,omitempty"`
	DefinitionProvider               bool                             `json:"definitionProvider,omitempty"`
	TypeDefinitionProvider           bool                             `json:"typeDefinitionProvider,omitempty"`
	ReferencesProvider               bool                             `json:"referencesProvider,omitempty"`
	DocumentHighlightProvider        bool                             `json:"documentHighlightProvider,omitempty"`
	DocumentSymbolProvider           bool                             `json:"documentSymbolProvider,omitempty"`
	WorkspaceSymbolProvider          bool                             `json:"workspaceSymbolProvider,omitempty"`
	ImplementationProvider           bool                             `json:"implementationProvider,omitempty"`
	CodeActionProvider               bool                             `json:"codeActionProvider,omitempty"`
	CodeLensProvider                 *CodeLensOptions                 `json:"codeLensProvider,omitempty"`
	DocumentFormattingProvider       bool                             `json:"documentFormattingProvider,omitempty"`
	DocumentRangeFormattingProvider  bool                             `json:"documentRangeFormattingProvider,omitempty"`
	DocumentOnTypeFormattingProvider *DocumentOnTypeFormattingOptions `json:"documentOnTypeFormattingProvider,omitempty"`
	RenameProvider                   bool                             `json:"renameProvider,omitempty"`
	ExecuteCommandProvider           *ExecuteCommandOptions           `json:"executeCommandProvider,omitempty"`
	SemanticHighlighting             *SemanticHighlightingOptions     `json:"semanticHighlighting,omitempty"`

	// XWorkspaceReferencesProvider indicates the server provides support for
	// xworkspace/references. This is a Sourcegraph extension.
	XWorkspaceReferencesProvider bool `json:"xworkspaceReferencesProvider,omitempty"`

	// XDefinitionProvider indicates the server provides support for
	// textDocument/xdefinition. This is a Sourcegraph extension.
	XDefinitionProvider bool `json:"xdefinitionProvider,omitempty"`

	// XWorkspaceSymbolByProperties indicates the server provides support for
	// querying symbols by properties with WorkspaceSymbolParams.symbol. This
	// is a Sourcegraph extension.
	XWorkspaceSymbolByProperties bool `json:"xworkspaceSymbolByProperties,omitempty"`

	Experimental interface{} `json:"experimental,omitempty"`
}

type CompletionOptions struct {
	ResolveProvider   bool     `json:"resolveProvider,omitempty"`
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

type DocumentOnTypeFormattingOptions struct {
	FirstTriggerCharacter string   `json:"firstTriggerCharacter"`
	MoreTriggerCharacter  []string `json:"moreTriggerCharacter,omitempty"`
}

type CodeLensOptions struct {
	ResolveProvider bool `json:"resolveProvider,omitempty"`
}

type SignatureHelpOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

type ExecuteCommandOptions struct {
	Commands []string `json:"commands"`
}

type ExecuteCommandParams struct {
	Command   string        `json:"command"`
	Arguments []interface{} `json:"arguments,omitempty"`
}

type SemanticHighlightingOptions struct {
	Scopes [][]string `json:"scopes,omitempty"`
}

type CompletionItemKind int

const (
	_ CompletionItemKind = iota
	CIKText
	CIKMethod
	CIKFunction
	CIKConstructor
	CIKField
	CIKVariable
	CIKClass
	CIKInterface
	CIKModule
	CIKProperty
	CIKUnit
	CIKValue
	CIKEnum
	CIKKeyword
	CIKSnippet
	CIKColor
	CIKFile
	CIKReference
	CIKFolder
	CIKEnumMember
	CIKConstant
	CIKStruct
	CIKEvent
	CIKOperator
	CIKTypeParameter
)

func (c CompletionItemKind) String() string {
	return completionItemKindName[c]
}

var completionItemKindName = map[CompletionItemKind]string{
	CIKText:          "text",
	CIKMethod:        "method",
	CIKFunction:      "function",
	CIKConstructor:   "constructor",
	CIKField:         "field",
	CIKVariable:      "variable",
	CIKClass:         "class",
	CIKInterface:     "interface",
	CIKModule:        "module",
	CIKProperty:      "property",
	CIKUnit:          "unit",
	CIKValue:         "value",
	CIKEnum:          "enum",
	CIKKeyword:       "keyword",
	CIKSnippet:       "snippet",
	CIKColor:         "color",
	CIKFile:          "file",
	CIKReference:     "reference",
	CIKFolder:        "folder",
	CIKEnumMember:    "enumMember",
	CIKConstant:      "constant",
	CIKStruct:        "struct",
	CIKEvent:         "event",
	CIKOperator:      "operator",
	CIKTypeParameter: "typeParameter",
}

type CompletionItem struct {
	Label            string             `json:"label"`
	Kind             CompletionItemKind `json:"kind,omitempty"`
	Detail           string             `json:"detail,omitempty"`
	Documentation    string             `json:"documentation,omitempty"`
	SortText         string             `json:"sortText,omitempty"`
	FilterText       string             `json:"filterText,omitempty"`
	InsertText       string             `json:"insertText,omitempty"`
	InsertTextFormat InsertTextFormat   `json:"insertTextFormat,omitempty"`
	TextEdit         *TextEdit          `json:"textEdit,omitempty"`
	Data             interface{}        `json:"data,omitempty"`
}

type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

type CompletionTriggerKind int

const (
	CTKInvoked          CompletionTriggerKind = 1
	CTKTriggerCharacter                       = 2
)

type DocumentationFormat string

const (
	DFPlainText DocumentationFormat = "plaintext"
)

type CodeActionKind string

const (
	CAKEmpty                 CodeActionKind = ""
	CAKQuickFix              CodeActionKind = "quickfix"
	CAKRefactor              CodeActionKind = "refactor"
	CAKRefactorExtract       CodeActionKind = "refactor.extract"
	CAKRefactorInline        CodeActionKind = "refactor.inline"
	CAKRefactorRewrite       CodeActionKind = "refactor.rewrite"
	CAKSource                CodeActionKind = "source"
	CAKSourceOrganizeImports CodeActionKind = "source.organizeImports"
)

type InsertTextFormat int

const (
	ITFPlainText InsertTextFormat = 1
	ITFSnippet                    = 2
)

type CompletionContext struct {
	TriggerKind      CompletionTriggerKind `json:"triggerKind"`
	TriggerCharacter string                `json:"triggerCharacter,omitempty"`
}

type CompletionParams struct {
	TextDocumentPositionParams
	Context CompletionContext `json:"context,omitempty"`
}

type Hover struct {
	Contents []MarkedString `json:"contents"`
	Range    *Range         `json:"range,omitempty"`
}

type hover Hover

func (h Hover) MarshalJSON() ([]byte, error) {
	if h.Contents == nil {
		return json.Marshal(hover{
			Contents: []MarkedString{},
			Range:    h.Range,
		})
	}
	return json.Marshal(hover(h))
}

type MarkedString markedString

type markedString struct {
	Language string `json:"language"`
	Value    string `json:"value"`

	isRawString bool
}

func (m *MarkedString) UnmarshalJSON(data []byte) error {
	if d := strings.TrimSpace(string(data)); len(d) > 0 && d[0] == '"' {
		// Raw string
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		m.Value = s
		m.isRawString = true
		return nil
	}
	// Language string
	ms := (*markedString)(m)
	return json.Unmarshal(data, ms)
}

func (m MarkedString) MarshalJSON() ([]byte, error) {
	if m.isRawString {
		return json.Marshal(m.Value)
	}
	return json.Marshal((markedString)(m))
}

// RawMarkedString returns a MarkedString consisting of only a raw
// string (i.e., "foo" instead of {"value":"foo", "language":"bar"}).
func RawMarkedString(s string) MarkedString {
	return MarkedString{Value: s, isRawString: true}
}

type SignatureHelp struct {
	Signatures      []SignatureInformation `json:"signatures"`
	ActiveSignature int                    `json:"activeSignature"`
	ActiveParameter int                    `json:"activeParameter"`
}

type SignatureInformation struct {
	Label         string                 `json:"label"`
	Documentation string                 `json:"documentation,omitempty"`
	Parameters    []ParameterInformation `json:"parameters,omitempty"`
}

type ParameterInformation struct {
	Label         string `json:"label"`
	Documentation string `json:"documentation,omitempty"`
}

type ReferenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`

	// Sourcegraph extension
	XLimit int `json:"xlimit,omitempty"`
}

type ReferenceParams struct {
	TextDocumentPositionParams
	Context ReferenceContext `json:"context"`
}

type DocumentHighlightKind int

const (
	Text  DocumentHighlightKind = 1
	Read                        = 2
	Write                       = 3
)

type DocumentHighlight struct {
	Range Range `json:"range"`
	Kind  int   `json:"kind,omitempty"`
}

type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type SymbolKind int

// The SymbolKind values are defined at https://microsoft.github.io/language-server-protocol/specification.
const (
	SKFile          SymbolKind = 1
	SKModule        SymbolKind = 2
	SKNamespace     SymbolKind = 3
	SKPackage       SymbolKind = 4
	SKClass         SymbolKind = 5
	SKMethod        SymbolKind = 6
	SKProperty      SymbolKind = 7
	SKField         SymbolKind = 8
	SKConstructor   SymbolKind = 9
	SKEnum          SymbolKind = 10
	SKInterface     SymbolKind = 11
	SKFunction      SymbolKind = 12
	SKVariable      SymbolKind = 13
	SKConstant      SymbolKind = 14
	SKString        SymbolKind = 15
	SKNumber        SymbolKind = 16
	SKBoolean       SymbolKind = 17
	SKArray         SymbolKind = 18
	SKObject        SymbolKind = 19
	SKKey           SymbolKind = 20
	SKNull          SymbolKind = 21
	SKEnumMember    SymbolKind = 22
	SKStruct        SymbolKind = 23
	SKEvent         SymbolKind = 24
	SKOperator      SymbolKind = 25
	SKTypeParameter SymbolKind = 26
)

func (s SymbolKind) String() string {
	return symbolKindName[s]
}

var symbolKindName = map[SymbolKind]string{
	SKFile:          "File",
	SKModule:        "Module",
	SKNamespace:     "Namespace",
	SKPackage:       "Package",
	SKClass:         "Class",
	SKMethod:        "Method",
	SKProperty:      "Property",
	SKField:         "Field",
	SKConstructor:   "Constructor",
	SKEnum:          "Enum",
	SKInterface:     "Interface",
	SKFunction:      "Function",
	SKVariable:      "Variable",
	SKConstant:      "Constant",
	SKString:        "String",
	SKNumber:        "Number",
	SKBoolean:       "Boolean",
	SKArray:         "Array",
	SKObject:        "Object",
	SKKey:           "Key",
	SKNull:          "Null",
	SKEnumMember:    "EnumMember",
	SKStruct:        "Struct",
	SKEvent:         "Event",
	SKOperator:      "Operator",
	SKTypeParameter: "TypeParameter",
}

type SymbolInformation struct {
	Name          string     `json:"name"`
	Kind          SymbolKind `json:"kind"`
	Location      Location   `json:"location"`
	ContainerName string     `json:"containerName,omitempty"`
}

type WorkspaceSymbolParams struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

type ConfigurationParams struct {
	Items []ConfigurationItem `json:"items"`
}

type ConfigurationItem struct {
	ScopeURI string `json:"scopeUri,omitempty"`
	Section  string `json:"section,omitempty"`
}

type ConfigurationResult []interface{}

type CodeActionContext struct {
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type CodeActionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
	Context      CodeActionContext      `json:"context"`
}

type CodeLensParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type CodeLens struct {
	Range   Range       `json:"range"`
	Command Command     `json:"command,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Options      FormattingOptions      `json:"options"`
}

type FormattingOptions struct {
	TabSize      int    `json:"tabSize"`
	InsertSpaces bool   `json:"insertSpaces"`
	Key          string `json:"key"`
}

type RenameParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	NewName      string                 `json:"newName"`
}

type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

type TextDocumentContentChangeEvent struct {
	Range       *Range `json:"range,omitEmpty"`
	RangeLength uint   `json:"rangeLength,omitEmpty"`
	Text        string `json:"text"`
}

type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type DidSaveTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type MessageType int

const (
	MTError   MessageType = 1
	MTWarning             = 2
	Info                  = 3
	Log                   = 4
)

type ShowMessageParams struct {
	Type    MessageType `json:"type"`
	Message string      `json:"message"`
}

type MessageActionItem struct {
	Title string `json:"title"`
}

type ShowMessageRequestParams struct {
	Type    MessageType         `json:"type"`
	Message string              `json:"message"`
	Actions []MessageActionItem `json:"actions"`
}

type LogMessageParams struct {
	Type    MessageType `json:"type"`
	Message string      `json:"message"`
}

type DidChangeConfigurationParams struct {
	Settings interface{} `json:"settings"`
}

type FileChangeType int

const (
	Created FileChangeType = 1
	Changed                = 2
	Deleted                = 3
)

type FileEvent struct {
	URI  DocumentURI `json:"uri"`
	Type int         `json:"type"`
}

type DidChangeWatchedFilesParams struct {
	Changes []FileEvent `json:"changes"`
}

type PublishDiagnosticsParams struct {
	URI         DocumentURI  `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type DocumentRangeFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
	Options      FormattingOptions      `json:"options"`
}

type DocumentOnTypeFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
	Ch           string                 `json:"ch"`
	Options      FormattingOptions      `json:"formattingOptions"`
}

type CancelParams struct {
	ID ID `json:"id"`
}

type SemanticHighlightingParams struct {
	TextDocument VersionedTextDocumentIdentifier   `json:"textDocument"`
	Lines        []SemanticHighlightingInformation `json:"lines"`
}

// SemanticHighlightingInformation represents a semantic highlighting
// information that has to be applied on a specific line of the text
// document.
type SemanticHighlightingInformation struct {
	// Line is the zero-based line position in the text document.
	Line int `json:"line"`

	// Tokens is a base64 encoded string representing every single highlighted
	// characters with its start position, length and the "lookup table" index of
	// the semantic highlighting [TextMate scopes](https://manual.macromates.com/en/language_grammars).
	// If the `tokens` is empty or not defined, then no highlighted positions are
	// available for the line.
	Tokens SemanticHighlightingTokens `json:"tokens,omitempty"`
}

type semanticHighlightingInformation struct {
	Line   int     `json:"line"`
	Tokens *string `json:"tokens"`
}

// MarshalJSON implements json.Marshaler.
func (v *SemanticHighlightingInformation) MarshalJSON() ([]byte, error) {
	tokens := string(v.Tokens.Serialize())
	return json.Marshal(&semanticHighlightingInformation{
		Line:   v.Line,
		Tokens: &tokens,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (v *SemanticHighlightingInformation) UnmarshalJSON(data []byte) error {
	var info semanticHighlightingInformation
	err := json.Unmarshal(data, &info)
	if err != nil {
		return err
	}

	if info.Tokens != nil {
		v.Tokens, err = DeserializeSemanticHighlightingTokens([]byte(*info.Tokens))
		if err != nil {
			return err
		}
	}

	v.Line = info.Line
	return nil
}

type SemanticHighlightingTokens []SemanticHighlightingToken

func (v SemanticHighlightingTokens) Serialize() []byte {
	var chunks [][]byte

	// Writes each token to `tokens` in the byte format specified by the LSP
	// proposal. Described below:
	// |<---- 4 bytes ---->|<-- 2 bytes -->|<--- 2 bytes -->|
	// |    character      |  length       |    index       |
	for _, token := range v {
		chunk := make([]byte, 8)
		binary.BigEndian.PutUint32(chunk[:4], token.Character)
		binary.BigEndian.PutUint16(chunk[4:6], token.Length)
		binary.BigEndian.PutUint16(chunk[6:], token.Scope)
		chunks = append(chunks, chunk)
	}

	src := make([]byte, len(chunks)*8)
	for i, chunk := range chunks {
		copy(src[i*8:i*8+8], chunk)
	}

	dst := make([]byte, base64.StdEncoding.EncodedLen(len(src)))
	base64.StdEncoding.Encode(dst, src)
	return dst
}

func DeserializeSemanticHighlightingTokens(src []byte) (SemanticHighlightingTokens, error) {
	dst := make([]byte, base64.StdEncoding.DecodedLen(len(src)))
	n, err := base64.StdEncoding.Decode(dst, src)
	if err != nil {
		return nil, err
	}

	var chunks [][]byte
	for i := 7; i < len(dst[:n]); i += 8 {
		chunks = append(chunks, dst[i-7:i+1])
	}

	var tokens SemanticHighlightingTokens
	for _, chunk := range chunks {
		tokens = append(tokens, SemanticHighlightingToken{
			Character: binary.BigEndian.Uint32(chunk[:4]),
			Length:    binary.BigEndian.Uint16(chunk[4:6]),
			Scope:     binary.BigEndian.Uint16(chunk[6:]),
		})
	}

	return tokens, nil
}

type SemanticHighlightingToken struct {
	Character uint32
	Length    uint16
	Scope     uint16
}
