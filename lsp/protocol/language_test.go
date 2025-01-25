// SPDX-FileCopyrightText: 2020 The Go Language Server Authors
// SPDX-License-Identifier: BSD-3-Clause

package protocol

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"encoding/json"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/a-h/templ/lsp/uri"
)

func TestCompletionParams(t *testing.T) {
	t.Parallel()

	const (
		wantWorkDoneToken      = "156edea9-9d8d-422f-b7ee-81a84594afbb"
		wantPartialResultToken = "dd134d84-c134-4d7a-a2a3-f8af3ef4a568"
	)
	const (
		want        = `{"textDocument":{"uri":"file:///path/to/test.go"},"position":{"line":25,"character":1},"workDoneToken":"` + wantWorkDoneToken + `","partialResultToken":"` + wantPartialResultToken + `","context":{"triggerCharacter":".","triggerKind":1}}`
		wantInvalid = `{"textDocument":{"uri":"file:///path/to/test.go"},"position":{"line":2,"character":0},"workDoneToken":"` + wantPartialResultToken + `","partialResultToken":"` + wantWorkDoneToken + `","context":{"triggerCharacter":".","triggerKind":1}}`
	)
	wantType := CompletionParams{
		TextDocumentPositionParams: TextDocumentPositionParams{
			TextDocument: TextDocumentIdentifier{
				URI: uri.File("/path/to/test.go"),
			},
			Position: Position{
				Line:      25,
				Character: 1,
			},
		},
		WorkDoneProgressParams: WorkDoneProgressParams{
			WorkDoneToken: NewProgressToken(wantWorkDoneToken),
		},
		PartialResultParams: PartialResultParams{
			PartialResultToken: NewProgressToken(wantPartialResultToken),
		},
		Context: &CompletionContext{
			TriggerCharacter: ".",
			TriggerKind:      CompletionTriggerKindInvoked,
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          CompletionParams
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             CompletionParams
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got CompletionParams
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreTypes(WorkDoneProgressParams{}, PartialResultParams{})); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}

				if workDoneToken := got.WorkDoneToken; workDoneToken != nil {
					if diff := cmp.Diff(fmt.Sprint(workDoneToken), wantWorkDoneToken); (diff != "") != tt.wantErr {
						t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
					}
				}

				if partialResultToken := got.PartialResultToken; partialResultToken != nil {
					if diff := cmp.Diff(fmt.Sprint(partialResultToken), wantPartialResultToken); (diff != "") != tt.wantErr {
						t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
					}
				}
			})
		}
	})
}

func TestCompletionTriggerKind_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		k    CompletionTriggerKind
		want string
	}{
		{
			name: "Invoked",
			k:    CompletionTriggerKindInvoked,
			want: "Invoked",
		},
		{
			name: "TriggerCharacter",
			k:    CompletionTriggerKindTriggerCharacter,
			want: "TriggerCharacter",
		},
		{
			name: "TriggerForIncompleteCompletions",
			k:    CompletionTriggerKindTriggerForIncompleteCompletions,
			want: "TriggerForIncompleteCompletions",
		},
		{
			name: "Unknown",
			k:    CompletionTriggerKind(0),
			want: "0",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.k.String(); got != tt.want {
				t.Errorf("CompletionTriggerKind.String() = %v, want %v", tt.want, got)
			}
		})
	}
}

func TestCompletionContext(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"triggerCharacter":".","triggerKind":1}`
		wantInvalid = `{"triggerCharacter":" ","triggerKind":0}`
	)
	wantType := CompletionContext{
		TriggerCharacter: ".",
		TriggerKind:      CompletionTriggerKindInvoked,
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          CompletionContext
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             CompletionContext
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got CompletionContext
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestCompletionList(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"isIncomplete":true,"items":[{"tags":[1],"detail":"string","documentation":"Detail a human-readable string with additional information about this item, like type or symbol information.","filterText":"Detail","insertTextFormat":2,"kind":5,"label":"Detail","preselect":true,"sortText":"00000","textEdit":{"range":{"start":{"line":255,"character":4},"end":{"line":255,"character":10}},"newText":"Detail: ${1:},"}}]}`
		wantInvalid = `{"isIncomplete":false,"items":[]}`
	)
	wantType := CompletionList{
		IsIncomplete: true,
		Items: []CompletionItem{
			{
				AdditionalTextEdits: nil,
				Command:             nil,
				CommitCharacters:    nil,
				Tags: []CompletionItemTag{
					CompletionItemTagDeprecated,
				},
				Deprecated:       false,
				Detail:           "string",
				Documentation:    "Detail a human-readable string with additional information about this item, like type or symbol information.",
				FilterText:       "Detail",
				InsertText:       "",
				InsertTextFormat: InsertTextFormatSnippet,
				Kind:             CompletionItemKindField,
				Label:            "Detail",
				Preselect:        true,
				SortText:         "00000",
				TextEdit: &TextEditOrInsertReplaceEdit{
					TextEdit: &TextEdit{
						Range: Range{
							Start: Position{
								Line:      255,
								Character: 4,
							},
							End: Position{
								Line:      255,
								Character: 10,
							},
						},
						NewText: "Detail: ${1:},",
					},
				},
			},
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          CompletionList
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             CompletionList
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got CompletionList
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestInsertTextFormat_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		k    InsertTextFormat
		want string
	}{
		{
			name: "PlainText",
			k:    InsertTextFormatPlainText,
			want: "PlainText",
		},
		{
			name: "Snippet",
			k:    InsertTextFormatSnippet,
			want: "Snippet",
		},
		{
			name: "Unknown",
			k:    InsertTextFormat(0),
			want: "0",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.k.String(); got != tt.want {
				t.Errorf("InsertTextFormat.String() = %v, want %v", tt.want, got)
			}
		})
	}
}

func TestInsertReplaceEdit(t *testing.T) {
	t.Parallel()

	const (
		want = `{"newText":"testNewText","insert":{"start":{"line":255,"character":4},"end":{"line":255,"character":10}},"replace":{"start":{"line":255,"character":4},"end":{"line":255,"character":10}}}`
	)
	wantType := InsertReplaceEdit{
		NewText: "testNewText",
		Insert: Range{
			Start: Position{
				Line:      255,
				Character: 4,
			},
			End: Position{
				Line:      255,
				Character: 10,
			},
		},
		Replace: Range{
			Start: Position{
				Line:      255,
				Character: 4,
			},
			End: Position{
				Line:      255,
				Character: 10,
			},
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          InsertReplaceEdit
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             InsertReplaceEdit
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got InsertReplaceEdit
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestInsertTextMode_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		k    InsertTextMode
		want string
	}{
		{
			name: "AsIs",
			k:    InsertTextModeAsIs,
			want: "AsIs",
		},
		{
			name: "AdjustIndentation",
			k:    InsertTextModeAdjustIndentation,
			want: "AdjustIndentation",
		},
		{
			name: "Unknown",
			k:    InsertTextMode(0),
			want: "0",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.k.String(); got != tt.want {
				t.Errorf("InsertTextMode.String() = %v, want %v", tt.want, got)
			}
		})
	}
}

func TestCompletionItem(t *testing.T) {
	t.Parallel()

	const (
		wantTextEdit = `{
  "additionalTextEdits": [
    {
      "range": {
        "start": {
          "line": 255,
          "character": 4
        },
        "end": {
          "line": 255,
          "character": 10
        }
      },
      "newText": "Detail: ${1:},"
    }
  ],
  "command": {
    "title": "exec echo",
    "command": "echo",
    "arguments": [
      "hello"
    ]
  },
  "commitCharacters": [
    "a"
  ],
  "tags": [
    1
  ],
  "data": "testData",
  "deprecated": true,
  "detail": "string",
  "documentation": "Detail a human-readable string with additional information about this item, like type or symbol information.",
  "filterText": "Detail",
  "insertText": "testInsert",
  "insertTextFormat": 2,
  "insertTextMode": 1,
  "kind": 5,
  "label": "Detail",
  "preselect": true,
  "sortText": "00000",
  "textEdit": {
    "range": {
      "start": {
        "line": 255,
        "character": 4
      },
      "end": {
        "line": 255,
        "character": 10
      }
    },
    "newText": "Detail: ${1:},"
  }
}`
		wantInsertReplaceEdit = `{
  "additionalTextEdits": [
    {
      "range": {
        "start": {
          "line": 255,
          "character": 4
        },
        "end": {
          "line": 255,
          "character": 10
        }
      },
      "newText": "Detail: ${1:},"
    }
  ],
  "command": {
    "title": "exec echo",
    "command": "echo",
    "arguments": [
      "hello"
    ]
  },
  "commitCharacters": [
    "a"
  ],
  "tags": [
    1
  ],
  "data": "testData",
  "deprecated": true,
  "detail": "string",
  "documentation": "Detail a human-readable string with additional information about this item, like type or symbol information.",
  "filterText": "Detail",
  "insertText": "testInsert",
  "insertTextFormat": 2,
  "insertTextMode": 1,
  "kind": 5,
  "label": "Detail",
  "preselect": true,
  "sortText": "00000",
  "textEdit": {
    "newText": "Detail: ${1:},",
    "insert": {
      "start": {
        "line": 105,
        "character": 65
      },
      "end": {
        "line": 105,
        "character": 72
      }
    },
    "replace": {
      "start": {
        "line": 105,
        "character": 65
      },
      "end": {
        "line": 105,
        "character": 76
      }
    }
  }
}`
		wantNilAll = `{
  "label": "Detail"
}`
		wantInvalid = `{"items":[]}`
	)
	wantTypeTextEdit := CompletionItem{
		AdditionalTextEdits: []TextEdit{
			{
				Range: Range{
					Start: Position{
						Line:      255,
						Character: 4,
					},
					End: Position{
						Line:      255,
						Character: 10,
					},
				},
				NewText: "Detail: ${1:},",
			},
		},
		Command: &Command{
			Title:     "exec echo",
			Command:   "echo",
			Arguments: []any{"hello"},
		},
		CommitCharacters: []string{"a"},
		Tags: []CompletionItemTag{
			CompletionItemTagDeprecated,
		},
		Data:             "testData",
		Deprecated:       true,
		Detail:           "string",
		Documentation:    "Detail a human-readable string with additional information about this item, like type or symbol information.",
		FilterText:       "Detail",
		InsertText:       "testInsert",
		InsertTextFormat: InsertTextFormatSnippet,
		InsertTextMode:   InsertTextModeAsIs,
		Kind:             CompletionItemKindField,
		Label:            "Detail",
		Preselect:        true,
		SortText:         "00000",
		TextEdit: &TextEditOrInsertReplaceEdit{
			TextEdit: &TextEdit{
				Range: Range{
					Start: Position{
						Line:      255,
						Character: 4,
					},
					End: Position{
						Line:      255,
						Character: 10,
					},
				},
				NewText: "Detail: ${1:},",
			},
		},
	}
	wantTypeInsertReplaceEdit := CompletionItem{
		AdditionalTextEdits: []TextEdit{
			{
				Range: Range{
					Start: Position{
						Line:      255,
						Character: 4,
					},
					End: Position{
						Line:      255,
						Character: 10,
					},
				},
				NewText: "Detail: ${1:},",
			},
		},
		Command: &Command{
			Title:     "exec echo",
			Command:   "echo",
			Arguments: []any{"hello"},
		},
		CommitCharacters: []string{"a"},
		Tags: []CompletionItemTag{
			CompletionItemTagDeprecated,
		},
		Data:             "testData",
		Deprecated:       true,
		Detail:           "string",
		Documentation:    "Detail a human-readable string with additional information about this item, like type or symbol information.",
		FilterText:       "Detail",
		InsertText:       "testInsert",
		InsertTextFormat: InsertTextFormatSnippet,
		InsertTextMode:   InsertTextModeAsIs,
		Kind:             CompletionItemKindField,
		Label:            "Detail",
		Preselect:        true,
		SortText:         "00000",
		TextEdit: &TextEditOrInsertReplaceEdit{
			InsertReplaceEdit: &InsertReplaceEdit{
				NewText: "Detail: ${1:},",
				Insert: Range{
					Start: Position{
						Line:      105,
						Character: 65,
					},
					End: Position{
						Line:      105,
						Character: 72,
					},
				},
				Replace: Range{
					Start: Position{
						Line:      105,
						Character: 65,
					},
					End: Position{
						Line:      105,
						Character: 76,
					},
				},
			},
		},
	}
	wantTypeNilAll := CompletionItem{
		Label: "Detail",
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          CompletionItem
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "ValidTextEdit",
				field:          wantTypeTextEdit,
				want:           wantTextEdit,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidInsertReplaceEdit",
				field:          wantTypeInsertReplaceEdit,
				want:           wantInsertReplaceEdit,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidNilAll",
				field:          wantTypeNilAll,
				want:           wantNilAll,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantTypeTextEdit,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				b := new(bytes.Buffer)
				enc := json.NewEncoder(b)
				enc.SetIndent("", "  ")
				err := enc.Encode(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}
				got := strings.TrimSpace(b.String())

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             CompletionItem
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "ValidTextEdit",
				field:            wantTextEdit,
				want:             wantTypeTextEdit,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "ValidInsertReplaceEdit",
				field:            wantInsertReplaceEdit,
				want:             wantTypeInsertReplaceEdit,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "ValidNilAll",
				field:            wantNilAll,
				want:             wantTypeNilAll,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantTypeTextEdit,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got CompletionItem
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestCompletionItemKind_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		k    CompletionItemKind
		want string
	}{
		{
			name: "Text",
			k:    CompletionItemKindText,
			want: "Text",
		},
		{
			name: "Method",
			k:    CompletionItemKindMethod,
			want: "Method",
		},
		{
			name: "Function",
			k:    CompletionItemKindFunction,
			want: "Function",
		},
		{
			name: "Constructor",
			k:    CompletionItemKindConstructor,
			want: "Constructor",
		},
		{
			name: "Field",
			k:    CompletionItemKindField,
			want: "Field",
		},
		{
			name: "Variable",
			k:    CompletionItemKindVariable,
			want: "Variable",
		},
		{
			name: "Class",
			k:    CompletionItemKindClass,
			want: "Class",
		},
		{
			name: "Interface",
			k:    CompletionItemKindInterface,
			want: "Interface",
		},
		{
			name: "Module",
			k:    CompletionItemKindModule,
			want: "Module",
		},
		{
			name: "Property",
			k:    CompletionItemKindProperty,
			want: "Property",
		},
		{
			name: "Unit",
			k:    CompletionItemKindUnit,
			want: "Unit",
		},
		{
			name: "Value",
			k:    CompletionItemKindValue,
			want: "Value",
		},
		{
			name: "Enum",
			k:    CompletionItemKindEnum,
			want: "Enum",
		},
		{
			name: "Keyword",
			k:    CompletionItemKindKeyword,
			want: "Keyword",
		},
		{
			name: "Snippet",
			k:    CompletionItemKindSnippet,
			want: "Snippet",
		},
		{
			name: "Color",
			k:    CompletionItemKindColor,
			want: "Color",
		},
		{
			name: "File",
			k:    CompletionItemKindFile,
			want: "File",
		},
		{
			name: "Reference",
			k:    CompletionItemKindReference,
			want: "Reference",
		},
		{
			name: "Folder",
			k:    CompletionItemKindFolder,
			want: "Folder",
		},
		{
			name: "EnumMember",
			k:    CompletionItemKindEnumMember,
			want: "EnumMember",
		},
		{
			name: "Constant",
			k:    CompletionItemKindConstant,
			want: "Constant",
		},
		{
			name: "Struct",
			k:    CompletionItemKindStruct,
			want: "Struct",
		},
		{
			name: "Event",
			k:    CompletionItemKindEvent,
			want: "Event",
		},
		{
			name: "Operator",
			k:    CompletionItemKindOperator,
			want: "Operator",
		},
		{
			name: "TypeParameter",
			k:    CompletionItemKindTypeParameter,
			want: "TypeParameter",
		},
		{
			name: "Unknown",
			k:    CompletionItemKind(0),
			want: "0",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.k.String(); got != tt.want {
				t.Errorf("CompletionItemKind.String() = %v, want %v", tt.want, got)
			}
		})
	}
}

func TestCompletionItemTag_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		k    CompletionItemTag
		want string
	}{
		{
			name: "Deprecated",
			k:    CompletionItemTagDeprecated,
			want: "Deprecated",
		},
		{
			name: "Unknown",
			k:    CompletionItemTag(0),
			want: "0",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.k.String(); got != tt.want {
				t.Errorf("CompletionItemTag.String() = %v, want %v", tt.want, got)
			}
		})
	}
}

func TestCompletionRegistrationOptions(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"documentSelector":[{"language":"go","scheme":"file","pattern":"*.go"}],"triggerCharacters":["."],"resolveProvider":true}`
		wantInvalid = `{"documentSelector":[{"language":"typescript","scheme":"file","pattern":"*.{ts,js}"}],"triggerCharacters":[" "],"resolveProvider":true}`
	)
	wantType := CompletionRegistrationOptions{
		TextDocumentRegistrationOptions: TextDocumentRegistrationOptions{
			DocumentSelector: DocumentSelector{
				{
					Language: "go",
					Scheme:   "file",
					Pattern:  "*.go",
				},
			},
		},
		TriggerCharacters: []string{"."},
		ResolveProvider:   true,
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          CompletionRegistrationOptions
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             CompletionRegistrationOptions
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got CompletionRegistrationOptions
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestHoverParams(t *testing.T) {
	t.Parallel()

	const (
		wantWorkDoneToken    = "156edea9-9d8d-422f-b7ee-81a84594afbb"
		invalidWorkDoneToken = "dd134d84-c134-4d7a-a2a3-f8af3ef4a568"
	)
	const (
		want        = `{"textDocument":{"uri":"file:///path/to/basic.go"},"position":{"line":25,"character":1},"workDoneToken":"` + wantWorkDoneToken + `"}`
		wantNilAll  = `{"textDocument":{"uri":"file:///path/to/basic.go"},"position":{"line":25,"character":1}}`
		wantInvalid = `{"textDocument":{"uri":"file:///path/to/basic_gen.go"},"position":{"line":2,"character":1},"workDoneToken":"` + invalidWorkDoneToken + `"}`
	)
	wantType := HoverParams{
		TextDocumentPositionParams: TextDocumentPositionParams{
			TextDocument: TextDocumentIdentifier{
				URI: uri.File("/path/to/basic.go"),
			},
			Position: Position{
				Line:      25,
				Character: 1,
			},
		},
		WorkDoneProgressParams: WorkDoneProgressParams{
			WorkDoneToken: NewProgressToken(wantWorkDoneToken),
		},
	}
	wantTypeNilAll := HoverParams{
		TextDocumentPositionParams: TextDocumentPositionParams{
			TextDocument: TextDocumentIdentifier{
				URI: uri.File("/path/to/basic.go"),
			},
			Position: Position{
				Line:      25,
				Character: 1,
			},
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          HoverParams
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidNilAll",
				field:          wantTypeNilAll,
				want:           wantNilAll,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name             string
			field            string
			want             HoverParams
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "ValidNilAll",
				field:            wantNilAll,
				want:             wantTypeNilAll,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got HoverParams
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreTypes(WorkDoneProgressParams{})); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}

				if workDoneToken := got.WorkDoneToken; workDoneToken != nil {
					if diff := cmp.Diff(fmt.Sprint(workDoneToken), wantWorkDoneToken); (diff != "") != tt.wantErr {
						t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
					}
				}
			})
		}
	})
}

func TestHover(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"contents":{"kind":"markdown","value":"example value"},"range":{"start":{"line":255,"character":4},"end":{"line":255,"character":10}}}`
		wantInvalid = `{"contents":{"kind":"markdown","value":"example value"},"range":{"start":{"line":25,"character":2},"end":{"line":25,"character":5}}}`
	)
	wantType := Hover{
		Contents: MarkupContent{
			Kind:  Markdown,
			Value: "example value",
		},
		Range: &Range{
			Start: Position{
				Line:      255,
				Character: 4,
			},
			End: Position{
				Line:      255,
				Character: 10,
			},
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          Hover
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             Hover
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got Hover
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestSignatureHelpParams(t *testing.T) {
	t.Parallel()

	const (
		wantWorkDoneToken    = "156edea9-9d8d-422f-b7ee-81a84594afbb"
		invalidWorkDoneToken = "dd134d84-c134-4d7a-a2a3-f8af3ef4a568"
	)
	const (
		want        = `{"textDocument":{"uri":"file:///path/to/basic.go"},"position":{"line":25,"character":1},"workDoneToken":"` + wantWorkDoneToken + `","context":{"triggerKind":1,"triggerCharacter":".","isRetrigger":true,"activeSignatureHelp":{"signatures":[{"label":"testLabel","documentation":"testDocumentation","parameters":[{"label":"test label","documentation":"test documentation"}]}],"activeParameter":10,"activeSignature":5}}}`
		wantNilAll  = `{"textDocument":{"uri":"file:///path/to/basic.go"},"position":{"line":25,"character":1}}`
		wantInvalid = `{"textDocument":{"uri":"file:///path/to/basic_gen.go"},"position":{"line":2,"character":1},"workDoneToken":"` + invalidWorkDoneToken + `","context":{"triggerKind":0,"triggerCharacter":"aaa","isRetrigger":false,"activeSignatureHelp":{"signatures":[{"documentationFormat":["markdown"],"parameterInformation":{"label":"test label","documentation":"test documentation"}}],"activeParameter":1,"activeSignature":0}}}`
	)
	wantType := SignatureHelpParams{
		TextDocumentPositionParams: TextDocumentPositionParams{
			TextDocument: TextDocumentIdentifier{
				URI: uri.File("/path/to/basic.go"),
			},
			Position: Position{
				Line:      25,
				Character: 1,
			},
		},
		WorkDoneProgressParams: WorkDoneProgressParams{
			WorkDoneToken: NewProgressToken(wantWorkDoneToken),
		},
		Context: &SignatureHelpContext{
			TriggerKind:      SignatureHelpTriggerKindInvoked,
			TriggerCharacter: ".",
			IsRetrigger:      true,
			ActiveSignatureHelp: &SignatureHelp{
				Signatures: []SignatureInformation{
					{
						Label:         "testLabel",
						Documentation: "testDocumentation",
						Parameters: []ParameterInformation{
							{
								Label:         "test label",
								Documentation: "test documentation",
							},
						},
					},
				},
				ActiveParameter: 10,
				ActiveSignature: 5,
			},
		},
	}
	wantTypeNilAll := SignatureHelpParams{
		TextDocumentPositionParams: TextDocumentPositionParams{
			TextDocument: TextDocumentIdentifier{
				URI: uri.File("/path/to/basic.go"),
			},
			Position: Position{
				Line:      25,
				Character: 1,
			},
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          SignatureHelpParams
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidNilAll",
				field:          wantTypeNilAll,
				want:           wantNilAll,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name             string
			field            string
			want             SignatureHelpParams
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "ValidNilAll",
				field:            wantNilAll,
				want:             wantTypeNilAll,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got SignatureHelpParams
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreTypes(WorkDoneProgressParams{})); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}

				if workDoneToken := got.WorkDoneToken; workDoneToken != nil {
					if diff := cmp.Diff(fmt.Sprint(workDoneToken), wantWorkDoneToken); (diff != "") != tt.wantErr {
						t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
					}
				}
			})
		}
	})
}

func TestSignatureHelpTriggerKind_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		k    SignatureHelpTriggerKind
		want string
	}{
		{
			name: "Invoked",
			k:    SignatureHelpTriggerKindInvoked,
			want: "Invoked",
		},
		{
			name: "TriggerCharacter",
			k:    SignatureHelpTriggerKindTriggerCharacter,
			want: "TriggerCharacter",
		},
		{
			name: "ContentChange",
			k:    SignatureHelpTriggerKindContentChange,
			want: "ContentChange",
		},
		{
			name: "Unknown",
			k:    SignatureHelpTriggerKind(0),
			want: "0",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.k.String(); got != tt.want {
				t.Errorf("SignatureHelpTriggerKind.String() = %v, want %v", tt.want, got)
			}
		})
	}
}

func TestSignatureHelp(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"signatures":[{"label":"testLabel","documentation":"testDocumentation","parameters":[{"label":"test label","documentation":"test documentation"}]}],"activeParameter":10,"activeSignature":5}`
		wantNilAll  = `{"signatures":[]}`
		wantInvalid = `{"signatures":[{"label":"invalidLabel","documentation":"invalidDocumentation","parameters":[{"label":"test label","documentation":"test documentation"}]}],"activeParameter":1,"activeSignature":0}`
	)
	wantType := SignatureHelp{
		Signatures: []SignatureInformation{
			{
				Label:         "testLabel",
				Documentation: "testDocumentation",
				Parameters: []ParameterInformation{
					{
						Label:         "test label",
						Documentation: "test documentation",
					},
				},
			},
		},
		ActiveParameter: 10,
		ActiveSignature: 5,
	}
	wantTypeNilAll := SignatureHelp{
		Signatures: []SignatureInformation{},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          SignatureHelp
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidNilAll",
				field:          wantTypeNilAll,
				want:           wantNilAll,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             SignatureHelp
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "ValidNilAll",
				field:            wantNilAll,
				want:             wantTypeNilAll,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got SignatureHelp
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestSignatureInformation(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"label":"testLabel","documentation":"testDocumentation","parameters":[{"label":"test label","documentation":"test documentation"}],"activeParameter":5}`
		wantInvalid = `{"label":"testLabel","documentation":"invalidDocumentation","parameters":[{"label":"test label","documentation":"test documentation"}],"activeParameter":50}`
	)
	wantType := SignatureInformation{
		Label:         "testLabel",
		Documentation: "testDocumentation",
		Parameters: []ParameterInformation{
			{
				Label:         "test label",
				Documentation: "test documentation",
			},
		},
		ActiveParameter: uint32(5),
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          SignatureInformation
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             SignatureInformation
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got SignatureInformation
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestParameterInformation(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"label":"test label","documentation":"test documentation"}`
		wantInvalid = `{"label":"invalid","documentation":"invalid"}`
	)
	wantType := ParameterInformation{
		Label:         "test label",
		Documentation: "test documentation",
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          ParameterInformation
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             ParameterInformation
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got ParameterInformation
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestSignatureHelpRegistrationOptions(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"documentSelector":[{"language":"go","scheme":"file","pattern":"*.go"}],"triggerCharacters":["."]}`
		wantInvalid = `{"documentSelector":[{"language":"typescript","scheme":"file","pattern":"*.{ts,js}"}],"triggerCharacters":[" "]}`
	)
	wantType := SignatureHelpRegistrationOptions{
		TextDocumentRegistrationOptions: TextDocumentRegistrationOptions{
			DocumentSelector: DocumentSelector{
				{
					Language: "go",
					Scheme:   "file",
					Pattern:  "*.go",
				},
			},
		},
		TriggerCharacters: []string{"."},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          SignatureHelpRegistrationOptions
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             SignatureHelpRegistrationOptions
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got SignatureHelpRegistrationOptions
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestReferenceParams(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"textDocument":{"uri":"file:///path/to/test.go"},"position":{"line":25,"character":1},"context":{"includeDeclaration":true}}`
		wantInvalid = `{"textDocument":{"uri":"file:///path/to/test.go"},"position":{"line":2,"character":0},"context":{"includeDeclaration":false}}`
	)
	wantType := ReferenceParams{
		TextDocumentPositionParams: TextDocumentPositionParams{
			TextDocument: TextDocumentIdentifier{
				URI: uri.File("/path/to/test.go"),
			},
			Position: Position{
				Line:      25,
				Character: 1,
			},
		},
		Context: ReferenceContext{
			IncludeDeclaration: true,
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          ReferenceParams
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             ReferenceParams
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got ReferenceParams
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestReferenceContext(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"includeDeclaration":true}`
		wantInvalid = `{"includeDeclaration":false}`
	)
	wantType := ReferenceContext{
		IncludeDeclaration: true,
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          ReferenceContext
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             ReferenceContext
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got ReferenceContext
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestDocumentHighlight(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"range":{"start":{"line":255,"character":4},"end":{"line":255,"character":10}},"kind":1}`
		wantInvalid = `{"range":{"start":{"line":25,"character":2},"end":{"line":25,"character":5}},"kind":1}`
	)
	wantType := DocumentHighlight{
		Range: Range{
			Start: Position{
				Line:      255,
				Character: 4,
			},
			End: Position{
				Line:      255,
				Character: 10,
			},
		},
		Kind: DocumentHighlightKindText,
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          DocumentHighlight
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             DocumentHighlight
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got DocumentHighlight
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestDocumentHighlightKind_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		k    DocumentHighlightKind
		want string
	}{
		{
			name: "Text",
			k:    DocumentHighlightKindText,
			want: "Text",
		},
		{
			name: "Read",
			k:    DocumentHighlightKindRead,
			want: "Read",
		},
		{
			name: "Write",
			k:    DocumentHighlightKindWrite,
			want: "Write",
		},
		{
			name: "Unknown",
			k:    DocumentHighlightKind(0),
			want: "0",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.k.String(); got != tt.want {
				t.Errorf("DocumentHighlightKind.String() = %v, want %v", tt.want, got)
			}
		})
	}
}

func TestDocumentSymbolParams(t *testing.T) {
	t.Parallel()

	const (
		wantWorkDoneToken      = "156edea9-9d8d-422f-b7ee-81a84594afbb"
		wantPartialResultToken = "dd134d84-c134-4d7a-a2a3-f8af3ef4a568"
	)
	wantType := DocumentSymbolParams{
		WorkDoneProgressParams: WorkDoneProgressParams{
			WorkDoneToken: NewProgressToken(wantWorkDoneToken),
		},
		PartialResultParams: PartialResultParams{
			PartialResultToken: NewProgressToken(wantPartialResultToken),
		},
		TextDocument: TextDocumentIdentifier{
			URI: uri.File("/path/to/test.go"),
		},
	}
	const (
		want        = `{"workDoneToken":"` + wantWorkDoneToken + `","partialResultToken":"` + wantPartialResultToken + `","textDocument":{"uri":"file:///path/to/test.go"}}`
		wantInvalid = `{"workDoneToken":"` + wantPartialResultToken + `","partialResultToken":"` + wantWorkDoneToken + `","textDocument":{"uri":"file:///path/to/nottest.go"}}`
	)

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          DocumentSymbolParams
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             DocumentSymbolParams
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got DocumentSymbolParams
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreTypes(WorkDoneProgressParams{}, PartialResultParams{})); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}

				if workDoneToken := got.WorkDoneToken; workDoneToken != nil {
					if diff := cmp.Diff(fmt.Sprint(workDoneToken), wantWorkDoneToken); (diff != "") != tt.wantErr {
						t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
					}
				}

				if partialResultToken := got.PartialResultToken; partialResultToken != nil {
					if diff := cmp.Diff(fmt.Sprint(partialResultToken), wantPartialResultToken); (diff != "") != tt.wantErr {
						t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
					}
				}
			})
		}
	})
}

func TestSymbolKind_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		k    SymbolKind
		want string
	}{
		{
			name: "File",
			k:    SymbolKindFile,
			want: "File",
		},
		{
			name: "Module",
			k:    SymbolKindModule,
			want: "Module",
		},
		{
			name: "Namespace",
			k:    SymbolKindNamespace,
			want: "Namespace",
		},
		{
			name: "Package",
			k:    SymbolKindPackage,
			want: "Package",
		},
		{
			name: "Class",
			k:    SymbolKindClass,
			want: "Class",
		},
		{
			name: "Method",
			k:    SymbolKindMethod,
			want: "Method",
		},
		{
			name: "Property",
			k:    SymbolKindProperty,
			want: "Property",
		},
		{
			name: "Field",
			k:    SymbolKindField,
			want: "Field",
		},
		{
			name: "Constructor",
			k:    SymbolKindConstructor,
			want: "Constructor",
		},
		{
			name: "Enum",
			k:    SymbolKindEnum,
			want: "Enum",
		},
		{
			name: "Interface",
			k:    SymbolKindInterface,
			want: "Interface",
		},
		{
			name: "Function",
			k:    SymbolKindFunction,
			want: "Function",
		},
		{
			name: "Variable",
			k:    SymbolKindVariable,
			want: "Variable",
		},
		{
			name: "Constant",
			k:    SymbolKindConstant,
			want: "Constant",
		},
		{
			name: "String",
			k:    SymbolKindString,
			want: "String",
		},
		{
			name: "Number",
			k:    SymbolKindNumber,
			want: "Number",
		},
		{
			name: "Boolean",
			k:    SymbolKindBoolean,
			want: "Boolean",
		},
		{
			name: "Array",
			k:    SymbolKindArray,
			want: "Array",
		},
		{
			name: "Object",
			k:    SymbolKindObject,
			want: "Object",
		},
		{
			name: "Key",
			k:    SymbolKindKey,
			want: "Key",
		},
		{
			name: "Null",
			k:    SymbolKindNull,
			want: "Null",
		},
		{
			name: "EnumMember",
			k:    SymbolKindEnumMember,
			want: "EnumMember",
		},
		{
			name: "Struct",
			k:    SymbolKindStruct,
			want: "Struct",
		},
		{
			name: "Event",
			k:    SymbolKindEvent,
			want: "Event",
		},
		{
			name: "Operator",
			k:    SymbolKindOperator,
			want: "Operator",
		},
		{
			name: "TypeParameter",
			k:    SymbolKindTypeParameter,
			want: "TypeParameter",
		},
		{
			name: "Unknown",
			k:    SymbolKind(0),
			want: "0",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.k.String(); got != tt.want {
				t.Errorf("SymbolKind.String() = %v, want %v", tt.want, got)
			}
		})
	}
}

func TestSymbolTag_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		k    SymbolTag
		want string
	}{
		{
			name: "Deprecated",
			k:    SymbolTagDeprecated,
			want: "Deprecated",
		},
		{
			name: "Unknown",
			k:    SymbolTag(0),
			want: "0",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.k.String(); got != tt.want {
				t.Errorf("SymbolTag.String() = %v, want %v", tt.want, got)
			}
		})
	}
}

func TestDocumentSymbol(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"name":"test symbol","detail":"test detail","kind":1,"tags":[1],"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":6}},"selectionRange":{"start":{"line":25,"character":3},"end":{"line":26,"character":10}},"children":[{"name":"child symbol","detail":"child detail","kind":11,"deprecated":true,"range":{"start":{"line":255,"character":4},"end":{"line":255,"character":10}},"selectionRange":{"start":{"line":255,"character":5},"end":{"line":255,"character":8}}}]}`
		wantInvalid = `{"name":"invalid symbol","detail":"invalid detail","kind":1,"tags":[0],"range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"selectionRange":{"start":{"line":2,"character":5},"end":{"line":3,"character":1}},"children":[{"name":"invalid child symbol","kind":1,"detail":"invalid child detail","range":{"start":{"line":255,"character":4},"end":{"line":255,"character":10}},"selectionRange":{"start":{"line":255,"character":5},"end":{"line":255,"character":8}}}]}`
	)
	wantType := DocumentSymbol{
		Name:   "test symbol",
		Detail: "test detail",
		Kind:   SymbolKindFile,
		Tags: []SymbolTag{
			SymbolTagDeprecated,
		},
		Deprecated: false,
		Range: Range{
			Start: Position{
				Line:      25,
				Character: 1,
			},
			End: Position{
				Line:      27,
				Character: 6,
			},
		},
		SelectionRange: Range{
			Start: Position{
				Line:      25,
				Character: 3,
			},
			End: Position{
				Line:      26,
				Character: 10,
			},
		},
		Children: []DocumentSymbol{
			{
				Name:       "child symbol",
				Detail:     "child detail",
				Kind:       SymbolKindInterface,
				Deprecated: true,
				Range: Range{
					Start: Position{
						Line:      255,
						Character: 4,
					},
					End: Position{
						Line:      255,
						Character: 10,
					},
				},
				SelectionRange: Range{
					Start: Position{
						Line:      255,
						Character: 5,
					},
					End: Position{
						Line:      255,
						Character: 8,
					},
				},
			},
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          DocumentSymbol
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             DocumentSymbol
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got DocumentSymbol
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestSymbolInformation(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"name":"test symbol","kind":1,"tags":[1],"deprecated":true,"location":{"uri":"file:///path/to/test.go","range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}}},"containerName":"testContainerName"}`
		wantNilAll  = `{"name":"test symbol","kind":1,"location":{"uri":"file:///path/to/test.go","range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}}}}`
		wantInvalid = `{"name":"invalid symbol","kind":1,"tags":[0],"deprecated":false,"location":{"uri":"file:///path/to/test_test.go","range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}}},"containerName":"invalidContainerName"}`
	)
	wantType := SymbolInformation{
		Name: "test symbol",
		Kind: 1,
		Tags: []SymbolTag{
			SymbolTagDeprecated,
		},
		Deprecated: true,
		Location: Location{
			URI: uri.File("/path/to/test.go"),
			Range: Range{
				Start: Position{
					Line:      25,
					Character: 1,
				},
				End: Position{
					Line:      27,
					Character: 3,
				},
			},
		},
		ContainerName: "testContainerName",
	}
	wantTypeNilAll := SymbolInformation{
		Name:       "test symbol",
		Kind:       1,
		Deprecated: false,
		Location: Location{
			URI: uri.File("/path/to/test.go"),
			Range: Range{
				Start: Position{
					Line:      25,
					Character: 1,
				},
				End: Position{
					Line:      27,
					Character: 3,
				},
			},
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          SymbolInformation
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidNilAll",
				field:          wantTypeNilAll,
				want:           wantNilAll,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             SymbolInformation
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "ValidNilAll",
				field:            wantNilAll,
				want:             wantTypeNilAll,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got SymbolInformation
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestCodeActionParams(t *testing.T) {
	t.Parallel()

	const (
		wantWorkDoneToken      = "156edea9-9d8d-422f-b7ee-81a84594afbb"
		wantPartialResultToken = "dd134d84-c134-4d7a-a2a3-f8af3ef4a568"
	)
	const (
		want        = `{"workDoneToken":"` + wantWorkDoneToken + `","partialResultToken":"` + wantPartialResultToken + `","textDocument":{"uri":"file:///path/to/test.go"},"context":{"diagnostics":[{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"severity":1,"code":"foo/bar","source":"test foo bar","message":"foo bar","relatedInformation":[{"location":{"uri":"file:///path/to/test.go","range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}}},"message":"test.go"}]}],"only":["quickfix"]},"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":6}}}`
		wantInvalid = `{"workDoneToken":"` + wantPartialResultToken + `","partialResultToken":"` + wantWorkDoneToken + `","textDocument":{"uri":"file:///path/to/test.go"},"context":{"diagnostics":[{"range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"severity":1,"code":"foo/bar","source":"test foo bar","message":"foo bar","relatedInformation":[{"location":{"uri":"file:///path/to/test.go","range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}}},"message":"test.go"}]}],"only":["quickfix"]},"range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}}}`
	)
	wantType := CodeActionParams{
		WorkDoneProgressParams: WorkDoneProgressParams{
			WorkDoneToken: NewProgressToken(wantWorkDoneToken),
		},
		PartialResultParams: PartialResultParams{
			PartialResultToken: NewProgressToken(wantPartialResultToken),
		},
		TextDocument: TextDocumentIdentifier{
			URI: uri.File("/path/to/test.go"),
		},
		Context: CodeActionContext{
			Diagnostics: []Diagnostic{
				{
					Range: Range{
						Start: Position{
							Line:      25,
							Character: 1,
						},
						End: Position{
							Line:      27,
							Character: 3,
						},
					},
					Severity: DiagnosticSeverityError,
					Code:     "foo/bar",
					Source:   "test foo bar",
					Message:  "foo bar",
					RelatedInformation: []DiagnosticRelatedInformation{
						{
							Location: Location{
								URI: uri.File("/path/to/test.go"),
								Range: Range{
									Start: Position{
										Line:      25,
										Character: 1,
									},
									End: Position{
										Line:      27,
										Character: 3,
									},
								},
							},
							Message: "test.go",
						},
					},
				},
			},
			Only: []CodeActionKind{
				QuickFix,
			},
		},
		Range: Range{
			Start: Position{
				Line:      25,
				Character: 1,
			},
			End: Position{
				Line:      27,
				Character: 6,
			},
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          CodeActionParams
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             CodeActionParams
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got CodeActionParams
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreTypes(WorkDoneProgressParams{}, PartialResultParams{})); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}

				if workDoneToken := got.WorkDoneToken; workDoneToken != nil {
					if diff := cmp.Diff(fmt.Sprint(workDoneToken), wantWorkDoneToken); (diff != "") != tt.wantErr {
						t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
					}
				}

				if partialResultToken := got.PartialResultToken; partialResultToken != nil {
					if diff := cmp.Diff(fmt.Sprint(partialResultToken), wantPartialResultToken); (diff != "") != tt.wantErr {
						t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
					}
				}
			})
		}
	})
}

func TestCodeActionKind_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		k    CodeActionKind
		want string
	}{
		{
			name: "QuickFix",
			k:    QuickFix,
			want: "quickfix",
		},
		{
			name: "Refactor",
			k:    Refactor,
			want: "refactor",
		},
		{
			name: "RefactorExtract",
			k:    RefactorExtract,
			want: "refactor.extract",
		},
		{
			name: "RefactorInline",
			k:    RefactorInline,
			want: "refactor.inline",
		},
		{
			name: "RefactorRewrite",
			k:    RefactorRewrite,
			want: "refactor.rewrite",
		},
		{
			name: "Source",
			k:    Source,
			want: "source",
		},
		{
			name: "SourceOrganizeImports",
			k:    SourceOrganizeImports,
			want: "source.organizeImports",
		},
		{
			name: "Unknown",
			k:    CodeActionKind(""),
			want: "",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.k; got != CodeActionKind(tt.want) {
				t.Errorf("CodeActionKind = %v, want %v", tt.want, got)
			}
		})
	}
}

func TestCodeActionContext(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"diagnostics":[{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"severity":1,"code":"foo/bar","source":"test foo bar","message":"foo bar","relatedInformation":[{"location":{"uri":"file:///path/to/test.go","range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}}},"message":"test.go"}]}],"only":["quickfix"]}`
		wantInvalid = `{"diagnostics":[{"range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"severity":1,"code":"foo/bar","source":"test foo bar","message":"foo bar","relatedInformation":[{"location":{"uri":"file:///path/to/test.go","range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}}},"message":"test.go"}]}],"only":["quickfix"]}`
	)
	wantType := CodeActionContext{
		Diagnostics: []Diagnostic{
			{
				Range: Range{
					Start: Position{
						Line:      25,
						Character: 1,
					},
					End: Position{
						Line:      27,
						Character: 3,
					},
				},
				Severity: DiagnosticSeverityError,
				Code:     "foo/bar",
				Source:   "test foo bar",
				Message:  "foo bar",
				RelatedInformation: []DiagnosticRelatedInformation{
					{
						Location: Location{
							URI: uri.File("/path/to/test.go"),
							Range: Range{
								Start: Position{
									Line:      25,
									Character: 1,
								},
								End: Position{
									Line:      27,
									Character: 3,
								},
							},
						},
						Message: "test.go",
					},
				},
			},
		},
		Only: []CodeActionKind{
			QuickFix,
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          CodeActionContext
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             CodeActionContext
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got CodeActionContext
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestCodeAction(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"title":"Refactoring","kind":"refactor.rewrite","diagnostics":[{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"severity":1,"code":"foo/bar","source":"test foo bar","message":"foo bar","relatedInformation":[{"location":{"uri":"file:///path/to/test.go","range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}}},"message":"test.go"}]}],"isPreferred":true,"disabled":{"reason":"testReason"},"edit":{"changes":{"file:///path/to/test.go":[{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"newText":"foo bar"}]},"documentChanges":[{"textDocument":{"uri":"file:///path/to/test.go","version":10},"edits":[{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"newText":"foo bar"}]}]},"command":{"title":"rewrite","command":"rewriter","arguments":["-w"]},"data":"testData"}`
		wantInvalid = `{"title":"Refactoring","kind":"refactor","diagnostics":[{"range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"severity":1,"code":"foo/bar","source":"test foo bar","message":"foo bar","relatedInformation":[{"location":{"uri":"file:///path/to/test.go","range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}}},"message":"test.go"}]}],"isPreferred":false,"disabled":{"reason":"invalidReason"},"edit":{"changes":{"file:///path/to/test.go":[{"range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"newText":"foo bar"}]},"documentChanges":[{"textDocument":{"uri":"file:///path/to/test.go","version":10},"edits":[{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"newText":"foo bar"}]}]},"command":{"title":"rewrite","command":"rewriter","arguments":["-w"]}}`
	)
	wantType := CodeAction{
		Title: "Refactoring",
		Kind:  RefactorRewrite,
		Diagnostics: []Diagnostic{
			{
				Range: Range{
					Start: Position{
						Line:      25,
						Character: 1,
					},
					End: Position{
						Line:      27,
						Character: 3,
					},
				},
				Severity: DiagnosticSeverityError,
				Code:     "foo/bar",
				Source:   "test foo bar",
				Message:  "foo bar",
				RelatedInformation: []DiagnosticRelatedInformation{
					{
						Location: Location{
							URI: uri.File("/path/to/test.go"),
							Range: Range{
								Start: Position{
									Line:      25,
									Character: 1,
								},
								End: Position{
									Line:      27,
									Character: 3,
								},
							},
						},
						Message: "test.go",
					},
				},
			},
		},
		IsPreferred: true,
		Disabled: &CodeActionDisable{
			Reason: "testReason",
		},
		Edit: &WorkspaceEdit{
			Changes: map[uri.URI][]TextEdit{
				uri.File("/path/to/test.go"): {
					{
						Range: Range{
							Start: Position{
								Line:      25,
								Character: 1,
							},
							End: Position{
								Line:      27,
								Character: 3,
							},
						},
						NewText: "foo bar",
					},
				},
			},
			DocumentChanges: []TextDocumentEdit{
				{
					TextDocument: OptionalVersionedTextDocumentIdentifier{
						TextDocumentIdentifier: TextDocumentIdentifier{
							URI: uri.File("/path/to/test.go"),
						},
						Version: NewVersion(int32(10)),
					},
					Edits: []TextEdit{
						{
							Range: Range{
								Start: Position{
									Line:      25,
									Character: 1,
								},
								End: Position{
									Line:      27,
									Character: 3,
								},
							},
							NewText: "foo bar",
						},
					},
				},
			},
		},
		Command: &Command{
			Title:     "rewrite",
			Command:   "rewriter",
			Arguments: []any{"-w"},
		},
		Data: "testData",
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          CodeAction
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             CodeAction
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got CodeAction
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestCodeActionRegistrationOptions(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"documentSelector":[{"language":"go","scheme":"file","pattern":"*.go"},{"language":"cpp","scheme":"untitled","pattern":"*.{cpp,hpp}"}],"codeActionKinds":["quickfix","refactor"]}`
		wantInvalid = `{"documentSelector":[{"language":"typescript","scheme":"file","pattern":"*.{ts,js}"},{"language":"c","scheme":"untitled","pattern":"*.{c,h}"}],"codeActionKinds":["quickfix","refactor"]}`
	)
	wantType := CodeActionRegistrationOptions{
		TextDocumentRegistrationOptions: TextDocumentRegistrationOptions{
			DocumentSelector: DocumentSelector{
				{
					Language: "go",
					Scheme:   "file",
					Pattern:  "*.go",
				},
				{
					Language: "cpp",
					Scheme:   "untitled",
					Pattern:  "*.{cpp,hpp}",
				},
			},
		},
		CodeActionOptions: CodeActionOptions{
			CodeActionKinds: []CodeActionKind{
				QuickFix,
				Refactor,
			},
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          CodeActionRegistrationOptions
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             CodeActionRegistrationOptions
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got CodeActionRegistrationOptions
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestCodeLensParams(t *testing.T) {
	t.Parallel()

	const (
		wantWorkDoneToken      = "156edea9-9d8d-422f-b7ee-81a84594afbb"
		wantPartialResultToken = "dd134d84-c134-4d7a-a2a3-f8af3ef4a568"
	)
	const (
		want        = `{"workDoneToken":"` + wantWorkDoneToken + `","partialResultToken":"` + wantPartialResultToken + `","textDocument":{"uri":"file:///path/to/test.go"}}`
		wantInvalid = `{"workDoneToken":"` + wantPartialResultToken + `","partialResultToken":"` + wantWorkDoneToken + `","textDocument":{"uri":"file:///path/to/invalid.go"}}`
	)
	wantType := CodeLensParams{
		WorkDoneProgressParams: WorkDoneProgressParams{
			WorkDoneToken: NewProgressToken(wantWorkDoneToken),
		},
		PartialResultParams: PartialResultParams{
			PartialResultToken: NewProgressToken(wantPartialResultToken),
		},
		TextDocument: TextDocumentIdentifier{
			URI: uri.File("/path/to/test.go"),
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          CodeLensParams
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             CodeLensParams
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got CodeLensParams
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreTypes(WorkDoneProgressParams{}, PartialResultParams{})); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}

				if workDoneToken := got.WorkDoneToken; workDoneToken != nil {
					if diff := cmp.Diff(fmt.Sprint(workDoneToken), wantWorkDoneToken); (diff != "") != tt.wantErr {
						t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
					}
				}

				if partialResultToken := got.PartialResultToken; partialResultToken != nil {
					if diff := cmp.Diff(fmt.Sprint(partialResultToken), wantPartialResultToken); (diff != "") != tt.wantErr {
						t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
					}
				}
			})
		}
	})
}

func TestCodeLens(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"command":{"title":"rewrite","command":"rewriter","arguments":["-w"]},"data":"testData"}`
		wantNilAll  = `{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}}}`
		wantInvalid = `{"range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"command":{"title":"rewrite","command":"rewriter","arguments":["-w"]},"data":"invalidData"}`
	)
	wantType := CodeLens{
		Range: Range{
			Start: Position{
				Line:      25,
				Character: 1,
			},
			End: Position{
				Line:      27,
				Character: 3,
			},
		},
		Command: &Command{
			Title:     "rewrite",
			Command:   "rewriter",
			Arguments: []any{"-w"},
		},
		Data: "testData",
	}
	wantTypeNilAll := CodeLens{
		Range: Range{
			Start: Position{
				Line:      25,
				Character: 1,
			},
			End: Position{
				Line:      27,
				Character: 3,
			},
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          CodeLens
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidNilAll",
				field:          wantTypeNilAll,
				want:           wantNilAll,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             CodeLens
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "ValidNilAll",
				field:            wantNilAll,
				want:             wantTypeNilAll,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got CodeLens
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestCodeLensRegistrationOptions(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"documentSelector":[{"language":"go","scheme":"file","pattern":"*.go"},{"language":"cpp","scheme":"untitled","pattern":"*.{cpp,hpp}"}],"resolveProvider":true}`
		wantNilAll  = `{"documentSelector":[{"language":"go","scheme":"file","pattern":"*.go"},{"language":"cpp","scheme":"untitled","pattern":"*.{cpp,hpp}"}]}`
		wantInvalid = `{"documentSelector":[{"language":"typescript","scheme":"file","pattern":"*.{ts,js}"},{"language":"c","scheme":"untitled","pattern":"*.{c,h}"}],"resolveProvider":false}`
	)
	wantType := CodeLensRegistrationOptions{
		TextDocumentRegistrationOptions: TextDocumentRegistrationOptions{
			DocumentSelector: DocumentSelector{
				{
					Language: "go",
					Scheme:   "file",
					Pattern:  "*.go",
				},
				{
					Language: "cpp",
					Scheme:   "untitled",
					Pattern:  "*.{cpp,hpp}",
				},
			},
		},
		ResolveProvider: true,
	}
	wantTypeNilAll := CodeLensRegistrationOptions{
		TextDocumentRegistrationOptions: TextDocumentRegistrationOptions{
			DocumentSelector: DocumentSelector{
				{
					Language: "go",
					Scheme:   "file",
					Pattern:  "*.go",
				},
				{
					Language: "cpp",
					Scheme:   "untitled",
					Pattern:  "*.{cpp,hpp}",
				},
			},
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          CodeLensRegistrationOptions
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidNilAll",
				field:          wantTypeNilAll,
				want:           wantNilAll,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             CodeLensRegistrationOptions
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "ValidNilAll",
				field:            wantNilAll,
				want:             wantTypeNilAll,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got CodeLensRegistrationOptions
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestDocumentLinkParams(t *testing.T) {
	t.Parallel()

	const (
		wantWorkDoneToken      = "156edea9-9d8d-422f-b7ee-81a84594afbb"
		wantPartialResultToken = "dd134d84-c134-4d7a-a2a3-f8af3ef4a568"
	)
	const (
		want        = `{"workDoneToken":"` + wantWorkDoneToken + `","partialResultToken":"` + wantPartialResultToken + `","textDocument":{"uri":"file:///path/to/test.go"}}`
		wantNilAll  = `{"textDocument":{"uri":"file:///path/to/test.go"}}`
		wantInvalid = `{"workDoneToken":"` + wantPartialResultToken + `","partialResultToken":"` + wantWorkDoneToken + `","textDocument":{"uri":"file:///path/to/invalid.go"}}`
	)
	wantType := DocumentLinkParams{
		WorkDoneProgressParams: WorkDoneProgressParams{
			WorkDoneToken: NewProgressToken(wantWorkDoneToken),
		},
		PartialResultParams: PartialResultParams{
			PartialResultToken: NewProgressToken(wantPartialResultToken),
		},
		TextDocument: TextDocumentIdentifier{
			URI: uri.File("/path/to/test.go"),
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          DocumentLinkParams
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             DocumentLinkParams
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got DocumentLinkParams
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreTypes(WorkDoneProgressParams{}, PartialResultParams{})); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}

				if workDoneToken := got.WorkDoneToken; workDoneToken != nil {
					if diff := cmp.Diff(fmt.Sprint(workDoneToken), wantWorkDoneToken); (diff != "") != tt.wantErr {
						t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
					}
				}

				if partialResultToken := got.PartialResultToken; partialResultToken != nil {
					if diff := cmp.Diff(fmt.Sprint(partialResultToken), wantPartialResultToken); (diff != "") != tt.wantErr {
						t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
					}
				}
			})
		}
	})
}

func TestDocumentLink(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"target":"file:///path/to/test.go","tooltip":"testTooltip","data":"testData"}`
		wantNilAll  = `{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}}}`
		wantInvalid = `{"range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"target":"file:///path/to/test.go","tooltip":"invalidTooltip","data":"testData"}`
	)
	wantType := DocumentLink{
		Range: Range{
			Start: Position{
				Line:      25,
				Character: 1,
			},
			End: Position{
				Line:      27,
				Character: 3,
			},
		},
		Target:  uri.File("/path/to/test.go"),
		Tooltip: "testTooltip",
		Data:    "testData",
	}
	wantTypeNilAll := DocumentLink{
		Range: Range{
			Start: Position{
				Line:      25,
				Character: 1,
			},
			End: Position{
				Line:      27,
				Character: 3,
			},
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          DocumentLink
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidNilAll",
				field:          wantTypeNilAll,
				want:           wantNilAll,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             DocumentLink
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "ValidNilAll",
				field:            wantNilAll,
				want:             wantTypeNilAll,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got DocumentLink
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestDocumentColorParams(t *testing.T) {
	t.Parallel()

	const (
		wantWorkDoneToken      = "156edea9-9d8d-422f-b7ee-81a84594afbb"
		wantPartialResultToken = "dd134d84-c134-4d7a-a2a3-f8af3ef4a568"
	)
	const (
		want        = `{"workDoneToken":"` + wantWorkDoneToken + `","partialResultToken":"` + wantPartialResultToken + `","textDocument":{"uri":"file:///path/to/test.go"}}`
		wantInvalid = `{"workDoneToken":"` + wantPartialResultToken + `","partialResultToken":"` + wantWorkDoneToken + `","textDocument":{"uri":"file:///path/to/invalid.go"}}`
	)
	wantType := DocumentColorParams{
		WorkDoneProgressParams: WorkDoneProgressParams{
			WorkDoneToken: NewProgressToken(wantWorkDoneToken),
		},
		PartialResultParams: PartialResultParams{
			PartialResultToken: NewProgressToken(wantPartialResultToken),
		},
		TextDocument: TextDocumentIdentifier{
			URI: uri.File("/path/to/test.go"),
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          DocumentColorParams
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             DocumentColorParams
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got DocumentColorParams
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreTypes(WorkDoneProgressParams{}, PartialResultParams{})); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}

				if workDoneToken := got.WorkDoneToken; workDoneToken != nil {
					if diff := cmp.Diff(fmt.Sprint(workDoneToken), wantWorkDoneToken); (diff != "") != tt.wantErr {
						t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
					}
				}

				if partialResultToken := got.PartialResultToken; partialResultToken != nil {
					if diff := cmp.Diff(fmt.Sprint(partialResultToken), wantPartialResultToken); (diff != "") != tt.wantErr {
						t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
					}
				}
			})
		}
	})
}

func TestColorInformation(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"color":{"alpha":1,"blue":0.2,"green":0.3,"red":0.4}}`
		wantInvalid = `{"range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"color":{"alpha":0,"blue":0.4,"green":0.3,"red":0.2}}`
	)
	wantType := ColorInformation{
		Range: Range{
			Start: Position{
				Line:      25,
				Character: 1,
			},
			End: Position{
				Line:      27,
				Character: 3,
			},
		},
		Color: Color{
			Alpha: 1,
			Blue:  0.2,
			Green: 0.3,
			Red:   0.4,
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          ColorInformation
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             ColorInformation
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got ColorInformation
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestColor(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"alpha":1,"blue":0.2,"green":0.3,"red":0.4}`
		wantInvalid = `{"alpha":0,"blue":0.4,"green":0.3,"red":0.2}`
	)
	wantType := Color{
		Alpha: 1,
		Blue:  0.2,
		Green: 0.3,
		Red:   0.4,
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          Color
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             Color
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got Color
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestColorPresentationParams(t *testing.T) {
	t.Parallel()

	const (
		wantWorkDoneToken      = "156edea9-9d8d-422f-b7ee-81a84594afbb"
		wantPartialResultToken = "dd134d84-c134-4d7a-a2a3-f8af3ef4a568"
	)
	const (
		want        = `{"workDoneToken":"` + wantWorkDoneToken + `","partialResultToken":"` + wantPartialResultToken + `","textDocument":{"uri":"file:///path/to/test.go"},"color":{"alpha":1,"blue":0.2,"green":0.3,"red":0.4},"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}}}`
		wantInvalid = `{"workDoneToken":"` + wantPartialResultToken + `","partialResultToken":"` + wantWorkDoneToken + `","textDocument":{"uri":"file:///path/to/test.go"},"color":{"alpha":0,"blue":0.4,"green":0.3,"red":0.2},"range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}}}`
	)
	wantType := ColorPresentationParams{
		WorkDoneProgressParams: WorkDoneProgressParams{
			WorkDoneToken: NewProgressToken(wantWorkDoneToken),
		},
		PartialResultParams: PartialResultParams{
			PartialResultToken: NewProgressToken(wantPartialResultToken),
		},
		TextDocument: TextDocumentIdentifier{
			URI: uri.File("/path/to/test.go"),
		},
		Color: Color{
			Alpha: 1,
			Blue:  0.2,
			Green: 0.3,
			Red:   0.4,
		},
		Range: Range{
			Start: Position{
				Line:      25,
				Character: 1,
			},
			End: Position{
				Line:      27,
				Character: 3,
			},
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          ColorPresentationParams
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             ColorPresentationParams
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got ColorPresentationParams
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreTypes(WorkDoneProgressParams{}, PartialResultParams{})); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}

				if workDoneToken := got.WorkDoneToken; workDoneToken != nil {
					if diff := cmp.Diff(fmt.Sprint(workDoneToken), wantWorkDoneToken); (diff != "") != tt.wantErr {
						t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
					}
				}

				if partialResultToken := got.PartialResultToken; partialResultToken != nil {
					if diff := cmp.Diff(fmt.Sprint(partialResultToken), wantPartialResultToken); (diff != "") != tt.wantErr {
						t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
					}
				}
			})
		}
	})
}

func TestColorPresentation(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"label":"testLabel","textEdit":{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"newText":"foo bar"},"additionalTextEdits":[{"range":{"start":{"line":100,"character":10},"end":{"line":102,"character":15}},"newText":"baz qux"}]}`
		wantNilAll  = `{"label":"testLabel"}`
		wantInvalid = `{"label":"invalidLabel","textEdit":{"range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"newText":"quux quuz"},"additionalTextEdits":[{"range":{"start":{"line":105,"character":15},"end":{"line":107,"character":20}},"newText":"corge grault"}]}`
	)
	wantType := ColorPresentation{
		Label: "testLabel",
		TextEdit: &TextEdit{
			Range: Range{
				Start: Position{
					Line:      25,
					Character: 1,
				},
				End: Position{
					Line:      27,
					Character: 3,
				},
			},
			NewText: "foo bar",
		},
		AdditionalTextEdits: []TextEdit{
			{
				Range: Range{
					Start: Position{
						Line:      100,
						Character: 10,
					},
					End: Position{
						Line:      102,
						Character: 15,
					},
				},
				NewText: "baz qux",
			},
		},
	}
	wantTypeNilAll := ColorPresentation{
		Label: "testLabel",
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          ColorPresentation
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidNilAll",
				field:          wantTypeNilAll,
				want:           wantNilAll,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             ColorPresentation
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "ValidNilAll",
				field:            wantNilAll,
				want:             wantTypeNilAll,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got ColorPresentation
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestDocumentFormattingParams(t *testing.T) {
	t.Parallel()

	const (
		wantWorkDoneToken    = "156edea9-9d8d-422f-b7ee-81a84594afbb"
		invalidWorkDoneToken = "dd134d84-c134-4d7a-a2a3-f8af3ef4a568"
	)
	const (
		want        = `{"workDoneToken":"` + wantWorkDoneToken + `","options":{"insertSpaces":true,"tabSize":4},"textDocument":{"uri":"file:///path/to/test.go"}}`
		wantInvalid = `{"workDoneToken":"` + invalidWorkDoneToken + `","options":{"insertSpaces":false,"tabSize":2},"textDocument":{"uri":"file:///path/to/invalid.go"}}`
	)
	wantType := DocumentFormattingParams{
		WorkDoneProgressParams: WorkDoneProgressParams{
			WorkDoneToken: NewProgressToken(wantWorkDoneToken),
		},
		Options: FormattingOptions{
			InsertSpaces: true,
			TabSize:      4,
		},
		TextDocument: TextDocumentIdentifier{
			URI: uri.File("/path/to/test.go"),
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          DocumentFormattingParams
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             DocumentFormattingParams
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got DocumentFormattingParams
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreTypes(WorkDoneProgressParams{})); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}

				if workDoneToken := got.WorkDoneToken; workDoneToken != nil {
					if diff := cmp.Diff(fmt.Sprint(workDoneToken), wantWorkDoneToken); (diff != "") != tt.wantErr {
						t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
					}
				}
			})
		}
	})
}

func TestFormattingOptions(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"insertSpaces":true,"tabSize":4,"trimTrailingWhitespace":true,"insertFinalNewline":true,"trimFinalNewlines":true,"key":{"test":"key"}}`
		wantInvalid = `{"insertSpaces":false,"tabSize":2,"trimTrailingWhitespace":false,"insertFinalNewline":false,"trimFinalNewlines":false}`
	)
	wantType := FormattingOptions{
		InsertSpaces:           true,
		TabSize:                4,
		TrimTrailingWhitespace: true,
		InsertFinalNewline:     true,
		TrimFinalNewlines:      true,
		Key: map[string]any{
			"test": "key",
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          FormattingOptions
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             FormattingOptions
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got FormattingOptions
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestDocumentRangeFormattingParams(t *testing.T) {
	t.Parallel()

	const (
		wantWorkDoneToken    = "156edea9-9d8d-422f-b7ee-81a84594afbb"
		invalidWorkDoneToken = "dd134d84-c134-4d7a-a2a3-f8af3ef4a568"
	)
	const (
		want        = `{"workDoneToken":"` + wantWorkDoneToken + `","textDocument":{"uri":"file:///path/to/test.go"},"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"options":{"insertSpaces":true,"tabSize":4}}`
		wantInvalid = `{"workDoneToken":"` + invalidWorkDoneToken + `","textDocument":{"uri":"file:///path/to/invalid.go"},"range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"options":{"insertSpaces":false,"tabSize":2}}`
	)
	wantType := DocumentRangeFormattingParams{
		WorkDoneProgressParams: WorkDoneProgressParams{
			WorkDoneToken: NewProgressToken(wantWorkDoneToken),
		},
		TextDocument: TextDocumentIdentifier{
			URI: uri.File("/path/to/test.go"),
		},
		Range: Range{
			Start: Position{
				Line:      25,
				Character: 1,
			},
			End: Position{
				Line:      27,
				Character: 3,
			},
		},
		Options: FormattingOptions{
			InsertSpaces: true,
			TabSize:      4,
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          DocumentRangeFormattingParams
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             DocumentRangeFormattingParams
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got DocumentRangeFormattingParams
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreTypes(WorkDoneProgressParams{})); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}

				if workDoneToken := got.WorkDoneToken; workDoneToken != nil {
					if diff := cmp.Diff(fmt.Sprint(workDoneToken), wantWorkDoneToken); (diff != "") != tt.wantErr {
						t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
					}
				}
			})
		}
	})
}

func TestDocumentOnTypeFormattingParams(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"textDocument":{"uri":"file:///path/to/test.go"},"position":{"line":25,"character":1},"ch":"character","options":{"insertSpaces":true,"tabSize":4}}`
		wantInvalid = `{"textDocument":{"uri":"file:///path/to/invalid.go"},"position":{"line":2,"character":1},"ch":"invalidChar","options":{"insertSpaces":false,"tabSize":2}}`
	)
	wantType := DocumentOnTypeFormattingParams{
		TextDocument: TextDocumentIdentifier{
			URI: uri.File("/path/to/test.go"),
		},
		Position: Position{
			Line:      25,
			Character: 1,
		},
		Ch: "character",
		Options: FormattingOptions{
			InsertSpaces: true,
			TabSize:      4,
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          DocumentOnTypeFormattingParams
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             DocumentOnTypeFormattingParams
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got DocumentOnTypeFormattingParams
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestDocumentOnTypeFormattingRegistrationOptions(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"documentSelector":[{"language":"go","scheme":"file","pattern":"*.go"},{"language":"cpp","scheme":"untitled","pattern":"*.{cpp,hpp}"}],"firstTriggerCharacter":"}","moreTriggerCharacter":[".","{"]}`
		wantInvalid = `{"documentSelector":[{"language":"typescript","scheme":"file","pattern":"*.{ts,js}"},{"language":"c","scheme":"untitled","pattern":"*.{c,h}"}],"firstTriggerCharacter":"{","moreTriggerCharacter":[" ","("]}`
	)
	wantType := DocumentOnTypeFormattingRegistrationOptions{
		TextDocumentRegistrationOptions: TextDocumentRegistrationOptions{
			DocumentSelector: DocumentSelector{
				{
					Language: "go",
					Scheme:   "file",
					Pattern:  "*.go",
				},
				{
					Language: "cpp",
					Scheme:   "untitled",
					Pattern:  "*.{cpp,hpp}",
				},
			},
		},
		FirstTriggerCharacter: "}",
		MoreTriggerCharacter:  []string{".", "{"},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          DocumentOnTypeFormattingRegistrationOptions
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             DocumentOnTypeFormattingRegistrationOptions
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got DocumentOnTypeFormattingRegistrationOptions
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestRenameParams(t *testing.T) {
	t.Parallel()

	const (
		wantPartialResultToken    = "156edea9-9d8d-422f-b7ee-81a84594afbb"
		invalidPartialResultToken = "dd134d84-c134-4d7a-a2a3-f8af3ef4a568"
	)
	const (
		want        = `{"textDocument":{"uri":"file:///path/to/test.go"},"position":{"line":25,"character":1},"partialResultToken":"` + wantPartialResultToken + `","newName":"newNameSymbol"}`
		wantInvalid = `{"textDocument":{"uri":"file:///path/to/invalid.go"},"position":{"line":2,"character":1},"partialResultToken":"` + invalidPartialResultToken + `","newName":"invalidSymbol"}`
	)
	wantType := RenameParams{
		TextDocumentPositionParams: TextDocumentPositionParams{
			TextDocument: TextDocumentIdentifier{
				URI: uri.File("/path/to/test.go"),
			},
			Position: Position{
				Line:      25,
				Character: 1,
			},
		},
		PartialResultParams: PartialResultParams{
			PartialResultToken: NewProgressToken(wantPartialResultToken),
		},
		NewName: "newNameSymbol",
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          RenameParams
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             RenameParams
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got RenameParams
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreTypes(PartialResultParams{})); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}

				if partialResultToken := got.PartialResultToken; partialResultToken != nil {
					if diff := cmp.Diff(fmt.Sprint(partialResultToken), wantPartialResultToken); (diff != "") != tt.wantErr {
						t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
					}
				}
			})
		}
	})
}

func TestRenameRegistrationOptions(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"documentSelector":[{"language":"go","scheme":"file","pattern":"*.go"},{"language":"cpp","scheme":"untitled","pattern":"*.{cpp,hpp}"}],"prepareProvider":true}`
		wantNilAll  = `{"documentSelector":[{"language":"go","scheme":"file","pattern":"*.go"},{"language":"cpp","scheme":"untitled","pattern":"*.{cpp,hpp}"}]}`
		wantInvalid = `{"documentSelector":[{"language":"typescript","scheme":"file","pattern":"*.{ts,js}"},{"language":"c","scheme":"untitled","pattern":"*.{c,h}"}],"prepareProvider":false}`
	)
	wantType := RenameRegistrationOptions{
		TextDocumentRegistrationOptions: TextDocumentRegistrationOptions{
			DocumentSelector: DocumentSelector{
				{
					Language: "go",
					Scheme:   "file",
					Pattern:  "*.go",
				},
				{
					Language: "cpp",
					Scheme:   "untitled",
					Pattern:  "*.{cpp,hpp}",
				},
			},
		},
		PrepareProvider: true,
	}
	wantTypeNilAll := RenameRegistrationOptions{
		TextDocumentRegistrationOptions: TextDocumentRegistrationOptions{
			DocumentSelector: DocumentSelector{
				{
					Language: "go",
					Scheme:   "file",
					Pattern:  "*.go",
				},
				{
					Language: "cpp",
					Scheme:   "untitled",
					Pattern:  "*.{cpp,hpp}",
				},
			},
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          RenameRegistrationOptions
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidNilAll",
				field:          wantTypeNilAll,
				want:           wantNilAll,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             RenameRegistrationOptions
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "ValidNilAll",
				field:            wantNilAll,
				want:             wantTypeNilAll,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got RenameRegistrationOptions
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestPrepareRenameParams(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"textDocument":{"uri":"file:///path/to/test.go"},"position":{"line":25,"character":1}}`
		wantInvalid = `{"textDocument":{"uri":"file:///path/to/invalid.go"},"position":{"line":2,"character":0}}`
	)
	wantType := PrepareRenameParams{
		TextDocumentPositionParams: TextDocumentPositionParams{
			TextDocument: TextDocumentIdentifier{
				URI: uri.File("/path/to/test.go"),
			},
			Position: Position{
				Line:      25,
				Character: 1,
			},
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          PrepareRenameParams
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             PrepareRenameParams
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got PrepareRenameParams
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreTypes(PartialResultParams{})); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}

func TestFoldingRangeParams(t *testing.T) {
	t.Parallel()

	const (
		wantPartialResultToken    = "156edea9-9d8d-422f-b7ee-81a84594afbb"
		invalidPartialResultToken = "dd134d84-c134-4d7a-a2a3-f8af3ef4a568"
	)
	const (
		want        = `{"textDocument":{"uri":"file:///path/to/test.go"},"position":{"line":25,"character":1},"partialResultToken":"` + wantPartialResultToken + `"}`
		wantInvalid = `{"textDocument":{"uri":"file:///path/to/invalid.go"},"position":{"line":2,"character":0},"partialResultToken":"` + invalidPartialResultToken + `"}`
	)
	wantType := FoldingRangeParams{
		TextDocumentPositionParams: TextDocumentPositionParams{
			TextDocument: TextDocumentIdentifier{
				URI: uri.File("/path/to/test.go"),
			},
			Position: Position{
				Line:      25,
				Character: 1,
			},
		},
		PartialResultParams: PartialResultParams{
			PartialResultToken: NewProgressToken(wantPartialResultToken),
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          FoldingRangeParams
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             FoldingRangeParams
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got FoldingRangeParams
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreTypes(PartialResultParams{})); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}

				if partialResultToken := got.PartialResultToken; partialResultToken != nil {
					if diff := cmp.Diff(fmt.Sprint(partialResultToken), wantPartialResultToken); (diff != "") != tt.wantErr {
						t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
					}
				}
			})
		}
	})
}

func TestFoldingRangeKind_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    FoldingRangeKind
		want string
	}{
		{
			name: "Comment",
			s:    CommentFoldingRange,
			want: "comment",
		},
		{
			name: "Imports",
			s:    ImportsFoldingRange,
			want: "imports",
		},
		{
			name: "Region",
			s:    RegionFoldingRange,
			want: "region",
		},
		{
			name: "Unknown",
			s:    FoldingRangeKind(""),
			want: "",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.s; !strings.EqualFold(tt.want, string(got)) {
				t.Errorf("FoldingRangeKind(%v), want %v", tt.want, got)
			}
		})
	}
}

func TestFoldingRange(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"startLine":10,"startCharacter":1,"endLine":10,"endCharacter":8,"kind":"imports"}`
		wantNilAll  = `{"startLine":10,"endLine":10}`
		wantInvalid = `{"startLine":0,"startCharacter":1,"endLine":0,"endCharacter":8,"kind":"comment"}`
	)
	wantType := FoldingRange{
		StartLine:      10,
		StartCharacter: 1,
		EndLine:        10,
		EndCharacter:   8,
		Kind:           ImportsFoldingRange,
	}
	wantTypeNilAll := FoldingRange{
		StartLine: 10,
		EndLine:   10,
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          FoldingRange
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "Valid",
				field:          wantType,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidNilAll",
				field:          wantTypeNilAll,
				want:           wantNilAll,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantType,
				want:           wantInvalid,
				wantMarshalErr: false,
				wantErr:        true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := json.Marshal(&tt.field)
				if (err != nil) != tt.wantMarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, string(got)); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		tests := []struct {
			name             string
			field            string
			want             FoldingRange
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "Valid",
				field:            want,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "ValidNilAll",
				field:            wantNilAll,
				want:             wantTypeNilAll,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got FoldingRange
				if err := json.Unmarshal([]byte(tt.field), &got); (err != nil) != tt.wantUnmarshalErr {
					t.Fatal(err)
				}

				if diff := cmp.Diff(tt.want, got); (diff != "") != tt.wantErr {
					t.Errorf("%s: wantErr: %t\n(-want +got)\n%s", tt.name, tt.wantErr, diff)
				}
			})
		}
	})
}
