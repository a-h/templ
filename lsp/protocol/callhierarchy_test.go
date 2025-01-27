// SPDX-FileCopyrightText: 2021 The Go Language Server Authors
// SPDX-License-Identifier: BSD-3-Clause

package protocol

import (
	"fmt"
	"testing"

	"encoding/json"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/a-h/templ/lsp/uri"
)

func TestCallHierarchy(t *testing.T) {
	const (
		want        = `{"dynamicRegistration":true}`
		wantNil     = `{}`
		wantInvalid = `{"dynamicRegistration":false}`
	)
	wantType := CallHierarchy{
		DynamicRegistration: true,
	}
	wantTypeNil := CallHierarchy{}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          CallHierarchy
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
				name:           "Nil",
				field:          wantTypeNil,
				want:           wantNil,
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
			want             CallHierarchy
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
				name:             "Valid",
				field:            wantNil,
				want:             wantTypeNil,
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

				var got CallHierarchy
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

func TestCallHierarchyOptions(t *testing.T) {
	const (
		want        = `{"workDoneProgress":true}`
		wantNil     = `{}`
		wantInvalid = `{"workDoneProgress":false}`
	)
	wantType := CallHierarchyOptions{
		WorkDoneProgressOptions: WorkDoneProgressOptions{
			WorkDoneProgress: true,
		},
	}
	wantTypeNil := CallHierarchyOptions{}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          CallHierarchyOptions
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
				name:           "Nil",
				field:          wantTypeNil,
				want:           wantNil,
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
			want             CallHierarchyOptions
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
				name:             "Valid",
				field:            wantNil,
				want:             wantTypeNil,
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

				var got CallHierarchyOptions
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

func TestCallHierarchyRegistrationOptions(t *testing.T) {
	const (
		want        = `{"documentSelector":[{"language":"go","scheme":"file","pattern":"*.go"}],"workDoneProgress":true,"id":"testID"}`
		wantNil     = `{"documentSelector":[]}`
		wantInvalid = `{"workDoneProgress":false}`
	)
	wantType := CallHierarchyRegistrationOptions{
		TextDocumentRegistrationOptions: TextDocumentRegistrationOptions{
			DocumentSelector: DocumentSelector{
				{
					Language: "go",
					Scheme:   "file",
					Pattern:  "*.go",
				},
			},
		},
		CallHierarchyOptions: CallHierarchyOptions{
			WorkDoneProgressOptions: WorkDoneProgressOptions{
				WorkDoneProgress: true,
			},
		},
		StaticRegistrationOptions: StaticRegistrationOptions{
			ID: "testID",
		},
	}
	wantTypeNil := CallHierarchyRegistrationOptions{
		TextDocumentRegistrationOptions: TextDocumentRegistrationOptions{
			DocumentSelector: DocumentSelector{},
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          CallHierarchyRegistrationOptions
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
				name:           "Nil",
				field:          wantTypeNil,
				want:           wantNil,
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
			want             CallHierarchyRegistrationOptions
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
				name:             "Valid",
				field:            wantNil,
				want:             wantTypeNil,
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

				var got CallHierarchyRegistrationOptions
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

func TestCallHierarchyPrepareParams(t *testing.T) {
	const (
		wantWorkDoneToken    = "156edea9-9d8d-422f-b7ee-81a84594afbb"
		invalidWorkDoneToken = "dd134d84-c134-4d7a-a2a3-f8af3ef4a568"
	)
	const (
		want        = `{"textDocument":{"uri":"file:///path/to/basic.go"},"position":{"line":25,"character":1},"workDoneToken":"` + wantWorkDoneToken + `"}`
		wantNil     = `{"textDocument":{"uri":"file:///path/to/basic.go"},"position":{"line":25,"character":1}}`
		wantInvalid = `{"textDocument":{"uri":"file:///path/to/basic_gen.go"},"position":{"line":2,"character":1},"workDoneToken":"` + invalidWorkDoneToken + `"}`
	)
	wantType := CallHierarchyPrepareParams{
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
	wantTypeNil := CallHierarchyPrepareParams{
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
			field          CallHierarchyPrepareParams
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
				name:           "Nil",
				field:          wantTypeNil,
				want:           wantNil,
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
			want             CallHierarchyPrepareParams
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
				name:             "Valid",
				field:            wantNil,
				want:             wantTypeNil,
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

				var got CallHierarchyPrepareParams
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

func TestCallHierarchyItem(t *testing.T) {
	const (
		want        = `{"name":"testName","kind":1,"tags":[1],"detail":"testDetail","uri":"file:///path/to/basic.go","range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"selectionRange":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"data":"testData"}`
		wantNil     = `{"name":"testName","kind":1,"uri":"file:///path/to/basic.go","range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"selectionRange":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}}}`
		wantInvalid = `{"name":"invalidName","kind":0,"tags":[0],"detail":"invalidDetail","uri":"file:///path/to/invalid.go","range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"selectionRange":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"data":"invalidData"}`
	)
	wantType := CallHierarchyItem{
		Name: "testName",
		Kind: SymbolKindFile,
		Tags: []SymbolTag{
			SymbolTagDeprecated,
		},
		Detail: "testDetail",
		URI:    uri.File("/path/to/basic.go"),
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
		SelectionRange: Range{
			Start: Position{
				Line:      25,
				Character: 1,
			},
			End: Position{
				Line:      27,
				Character: 3,
			},
		},
		Data: "testData",
	}
	wantTypeNil := CallHierarchyItem{
		Name: "testName",
		Kind: SymbolKindFile,
		URI:  uri.File("/path/to/basic.go"),
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
		SelectionRange: Range{
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
		t.Parallel()

		tests := []struct {
			name           string
			field          CallHierarchyItem
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
				name:           "Nil",
				field:          wantTypeNil,
				want:           wantNil,
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
			want             CallHierarchyItem
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
				name:             "Valid",
				field:            wantNil,
				want:             wantTypeNil,
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

				var got CallHierarchyItem
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

func TestCallHierarchyIncomingCallsParams(t *testing.T) {
	const (
		wantWorkDoneToken      = "156edea9-9d8d-422f-b7ee-81a84594afbb"
		wantPartialResultToken = "dd134d84-c134-4d7a-a2a3-f8af3ef4a568"
	)
	const (
		want        = `{"workDoneToken":"` + wantWorkDoneToken + `","partialResultToken":"` + wantPartialResultToken + `","item":{"name":"testName","kind":1,"tags":[1],"detail":"testDetail","uri":"file:///path/to/basic.go","range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"selectionRange":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"data":"testData"}}`
		wantNil     = `{"item":{"name":"testName","kind":1,"uri":"file:///path/to/basic.go","range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"selectionRange":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}}}}`
		wantInvalid = `{"workDoneToken":"` + wantPartialResultToken + `","partialResultToken":"` + wantWorkDoneToken + `","item":{"name":"invalidName","kind":0,"tags":[0],"detail":"invalidDetail","uri":"file:///path/to/invalid.go","range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"selectionRange":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"data":"invalidData"}}`
	)
	wantType := CallHierarchyIncomingCallsParams{
		WorkDoneProgressParams: WorkDoneProgressParams{
			WorkDoneToken: NewProgressToken(wantWorkDoneToken),
		},
		PartialResultParams: PartialResultParams{
			PartialResultToken: NewProgressToken(wantPartialResultToken),
		},
		Item: CallHierarchyItem{
			Name: "testName",
			Kind: SymbolKindFile,
			Tags: []SymbolTag{
				SymbolTagDeprecated,
			},
			Detail: "testDetail",
			URI:    uri.File("/path/to/basic.go"),
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
			SelectionRange: Range{
				Start: Position{
					Line:      25,
					Character: 1,
				},
				End: Position{
					Line:      27,
					Character: 3,
				},
			},
			Data: "testData",
		},
	}
	wantTypeNil := CallHierarchyIncomingCallsParams{
		Item: CallHierarchyItem{
			Name: "testName",
			Kind: SymbolKindFile,
			URI:  uri.File("/path/to/basic.go"),
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
			SelectionRange: Range{
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
		t.Parallel()

		tests := []struct {
			name           string
			field          CallHierarchyIncomingCallsParams
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
				name:           "Nil",
				field:          wantTypeNil,
				want:           wantNil,
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
			want             CallHierarchyIncomingCallsParams
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
				name:             "Nil",
				field:            wantNil,
				want:             wantTypeNil,
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

				var got CallHierarchyIncomingCallsParams
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

func TestCallHierarchyIncomingCall(t *testing.T) {
	const (
		want        = `{"from":{"name":"testName","kind":1,"tags":[1],"detail":"testDetail","uri":"file:///path/to/basic.go","range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"selectionRange":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"data":"testData"},"fromRanges":[{"start":{"line":25,"character":1},"end":{"line":27,"character":3}}]}`
		wantInvalid = `{"from":{"name":"invalidName","kind":0,"tags":[0],"detail":"invalidDetail","uri":"file:///path/to/invalid.go","range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"selectionRange":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"data":"invalidData"},"fromRanges":[{"start":{"line":2,"character":1},"end":{"line":3,"character":2}}]}`
	)
	wantType := CallHierarchyIncomingCall{
		From: CallHierarchyItem{
			Name: "testName",
			Kind: SymbolKindFile,
			Tags: []SymbolTag{
				SymbolTagDeprecated,
			},
			Detail: "testDetail",
			URI:    uri.File("/path/to/basic.go"),
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
			SelectionRange: Range{
				Start: Position{
					Line:      25,
					Character: 1,
				},
				End: Position{
					Line:      27,
					Character: 3,
				},
			},
			Data: "testData",
		},
		FromRanges: []Range{
			{
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
		t.Parallel()

		tests := []struct {
			name           string
			field          CallHierarchyIncomingCall
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
		t.Parallel()

		tests := []struct {
			name             string
			field            string
			want             CallHierarchyIncomingCall
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

				var got CallHierarchyIncomingCall
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

func TestCallHierarchyOutgoingCallsParams(t *testing.T) {
	const (
		wantWorkDoneToken      = "156edea9-9d8d-422f-b7ee-81a84594afbb"
		wantPartialResultToken = "dd134d84-c134-4d7a-a2a3-f8af3ef4a568"
	)
	const (
		want        = `{"workDoneToken":"` + wantWorkDoneToken + `","partialResultToken":"` + wantPartialResultToken + `","item":{"name":"testName","kind":1,"tags":[1],"detail":"testDetail","uri":"file:///path/to/basic.go","range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"selectionRange":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"data":"testData"}}`
		wantNil     = `{"item":{"name":"testName","kind":1,"uri":"file:///path/to/basic.go","range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"selectionRange":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}}}}`
		wantInvalid = `{"workDoneToken":"` + wantPartialResultToken + `","partialResultToken":"` + wantWorkDoneToken + `","item":{"name":"invalidName","kind":0,"tags":[0],"detail":"invalidDetail","uri":"file:///path/to/invalid.go","range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"selectionRange":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"data":"invalidData"}}`
	)
	wantType := CallHierarchyOutgoingCallsParams{
		WorkDoneProgressParams: WorkDoneProgressParams{
			WorkDoneToken: NewProgressToken(wantWorkDoneToken),
		},
		PartialResultParams: PartialResultParams{
			PartialResultToken: NewProgressToken(wantPartialResultToken),
		},
		Item: CallHierarchyItem{
			Name: "testName",
			Kind: SymbolKindFile,
			Tags: []SymbolTag{
				SymbolTagDeprecated,
			},
			Detail: "testDetail",
			URI:    uri.File("/path/to/basic.go"),
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
			SelectionRange: Range{
				Start: Position{
					Line:      25,
					Character: 1,
				},
				End: Position{
					Line:      27,
					Character: 3,
				},
			},
			Data: "testData",
		},
	}
	wantTypeNil := CallHierarchyOutgoingCallsParams{
		Item: CallHierarchyItem{
			Name: "testName",
			Kind: SymbolKindFile,
			URI:  uri.File("/path/to/basic.go"),
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
			SelectionRange: Range{
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
		t.Parallel()

		tests := []struct {
			name           string
			field          CallHierarchyOutgoingCallsParams
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
				name:           "Nil",
				field:          wantTypeNil,
				want:           wantNil,
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
			want             CallHierarchyOutgoingCallsParams
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
				name:             "Nil",
				field:            wantNil,
				want:             wantTypeNil,
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

				var got CallHierarchyOutgoingCallsParams
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

func TestCallHierarchyOutgoingCall(t *testing.T) {
	const (
		want        = `{"to":{"name":"testName","kind":1,"tags":[1],"detail":"testDetail","uri":"file:///path/to/basic.go","range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"selectionRange":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"data":"testData"},"fromRanges":[{"start":{"line":25,"character":1},"end":{"line":27,"character":3}}]}`
		wantInvalid = `{"to":{"name":"invalidName","kind":0,"tags":[0],"detail":"invalidDetail","uri":"file:///path/to/invalid.go","range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"selectionRange":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"data":"invalidData"},"fromRanges":[{"start":{"line":2,"character":1},"end":{"line":3,"character":2}}]}`
	)
	wantType := CallHierarchyOutgoingCall{
		To: CallHierarchyItem{
			Name: "testName",
			Kind: SymbolKindFile,
			Tags: []SymbolTag{
				SymbolTagDeprecated,
			},
			Detail: "testDetail",
			URI:    uri.File("/path/to/basic.go"),
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
			SelectionRange: Range{
				Start: Position{
					Line:      25,
					Character: 1,
				},
				End: Position{
					Line:      27,
					Character: 3,
				},
			},
			Data: "testData",
		},
		FromRanges: []Range{
			{
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
		t.Parallel()

		tests := []struct {
			name           string
			field          CallHierarchyOutgoingCall
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
		t.Parallel()

		tests := []struct {
			name             string
			field            string
			want             CallHierarchyOutgoingCall
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

				var got CallHierarchyOutgoingCall
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
