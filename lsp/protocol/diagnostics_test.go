// SPDX-FileCopyrightText: 2019 The Go Language Server Authors
// SPDX-License-Identifier: BSD-3-Clause

package protocol

import (
	"testing"

	"encoding/json"
	"github.com/google/go-cmp/cmp"

	"github.com/a-h/templ/lsp/uri"
)

func TestDiagnostic(t *testing.T) {
	t.Parallel()

	const (
		want                      = `{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"severity":1,"code":"foo/bar","codeDescription":{"href":"file:///path/to/test.go"},"source":"test foo bar","message":"foo bar","tags":[1,2],"relatedInformation":[{"location":{"uri":"file:///path/to/basic.go","range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}}},"message":"basic_gen.go"}],"data":"testData"}`
		wantNilSeverity           = `{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"code":"foo/bar","codeDescription":{"href":"file:///path/to/test.go"},"source":"test foo bar","message":"foo bar","relatedInformation":[{"location":{"uri":"file:///path/to/basic.go","range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}}},"message":"basic_gen.go"}],"data":"testData"}`
		wantNilCode               = `{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"severity":1,"codeDescription":{"href":"file:///path/to/test.go"},"source":"test foo bar","message":"foo bar","relatedInformation":[{"location":{"uri":"file:///path/to/basic.go","range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}}},"message":"basic_gen.go"}],"data":"testData"}`
		wantNilRelatedInformation = `{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"severity":1,"code":"foo/bar","codeDescription":{"href":"file:///path/to/test.go"},"source":"test foo bar","message":"foo bar","data":"testData"}`
		wantNilAll                = `{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"message":"foo bar"}`
		wantInvalid               = `{"range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"severity":1,"code":"foo/bar","codeDescription":{"href":"file:///path/to/test.go"},"source":"test foo bar","message":"foo bar","relatedInformation":[{"location":{"uri":"file:///path/to/basic.go","range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}}},"message":"basic_gen.go"}],"data":"invalidData"}`
	)
	wantType := Diagnostic{
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
		CodeDescription: &CodeDescription{
			Href: uri.File("/path/to/test.go"),
		},
		Source:  "test foo bar",
		Message: "foo bar",
		Tags: []DiagnosticTag{
			DiagnosticTagUnnecessary,
			DiagnosticTagDeprecated,
		},
		RelatedInformation: []DiagnosticRelatedInformation{
			{
				Location: Location{
					URI: uri.File("/path/to/basic.go"),
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
				Message: "basic_gen.go",
			},
		},
		Data: "testData",
	}
	wantTypeNilSeverity := Diagnostic{
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
		Code: "foo/bar",
		CodeDescription: &CodeDescription{
			Href: uri.File("/path/to/test.go"),
		},
		Source:  "test foo bar",
		Message: "foo bar",
		RelatedInformation: []DiagnosticRelatedInformation{
			{
				Location: Location{
					URI: uri.File("/path/to/basic.go"),
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
				Message: "basic_gen.go",
			},
		},
		Data: "testData",
	}
	wantTypeNilCode := Diagnostic{
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
		CodeDescription: &CodeDescription{
			Href: uri.File("/path/to/test.go"),
		},
		Source:  "test foo bar",
		Message: "foo bar",
		RelatedInformation: []DiagnosticRelatedInformation{
			{
				Location: Location{
					URI: uri.File("/path/to/basic.go"),
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
				Message: "basic_gen.go",
			},
		},
		Data: "testData",
	}
	wantTypeNilRelatedInformation := Diagnostic{
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
		CodeDescription: &CodeDescription{
			Href: uri.File("/path/to/test.go"),
		},
		Source:  "test foo bar",
		Message: "foo bar",
		Data:    "testData",
	}
	wantTypeNilAll := Diagnostic{
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
		Message: "foo bar",
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          Diagnostic
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
				name:           "ValidNilSeverity",
				field:          wantTypeNilSeverity,
				want:           wantNilSeverity,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidNilCode",
				field:          wantTypeNilCode,
				want:           wantNilCode,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidNilRelatedInformation",
				field:          wantTypeNilRelatedInformation,
				want:           wantNilRelatedInformation,
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
			want             Diagnostic
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
				name:             "ValidNilSeverity",
				field:            wantNilSeverity,
				want:             wantTypeNilSeverity,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "ValidNilCode",
				field:            wantNilCode,
				want:             wantTypeNilCode,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "ValidNilRelatedInformation",
				field:            wantNilRelatedInformation,
				want:             wantTypeNilRelatedInformation,
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

				var got Diagnostic
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

func TestDiagnosticSeverity_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		d    DiagnosticSeverity
		want string
	}{
		{
			name: "Error",
			d:    DiagnosticSeverityError,
			want: "Error",
		},
		{
			name: "Warning",
			d:    DiagnosticSeverityWarning,
			want: "Warning",
		},
		{
			name: "Information",
			d:    DiagnosticSeverityInformation,
			want: "Information",
		},
		{
			name: "Hint",
			d:    DiagnosticSeverityHint,
			want: "Hint",
		},
		{
			name: "Unknown",
			d:    DiagnosticSeverity(0),
			want: "0",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.d.String(); got != tt.want {
				t.Errorf("DiagnosticSeverity.String() = %v, want %v", tt.want, got)
			}
		})
	}
}

func TestDiagnosticTag_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		d    DiagnosticTag
		want string
	}{
		{
			name: "Unnecessary",
			d:    DiagnosticTagUnnecessary,
			want: "Unnecessary",
		},
		{
			name: "Deprecated",
			d:    DiagnosticTagDeprecated,
			want: "Deprecated",
		},
		{
			name: "Unknown",
			d:    DiagnosticTag(0),
			want: "0",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.d.String(); got != tt.want {
				t.Errorf("DiagnosticSeverity.String() = %v, want %v", tt.want, got)
			}
		})
	}
}

func TestDiagnosticRelatedInformation(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"location":{"uri":"file:///path/to/basic.go","range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}}},"message":"basic_gen.go"}`
		wantInvalid = `{"location":{"uri":"file:///path/to/basic.go","range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}}},"message":"basic_gen.go"}`
	)
	wantType := DiagnosticRelatedInformation{
		Location: Location{
			URI: uri.File("/path/to/basic.go"),
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
		Message: "basic_gen.go",
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          DiagnosticRelatedInformation
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
			want             DiagnosticRelatedInformation
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

				var got DiagnosticRelatedInformation
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

func TestPublishDiagnosticsParams(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"uri":"file:///path/to/diagnostics.go","version":1,"diagnostics":[{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"severity":1,"code":"foo/bar","source":"test foo bar","message":"foo bar","relatedInformation":[{"location":{"uri":"file:///path/to/diagnostics.go","range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}}},"message":"diagnostics.go"}]}]}`
		wantInvalid = `{"uri":"file:///path/to/diagnostics_gen.go","version":2,"diagnostics":[{"range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"severity":1,"code":"foo/bar","source":"test foo bar","message":"foo bar","relatedInformation":[{"location":{"uri":"file:///path/to/diagnostics_gen.go","range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}}},"message":"diagnostics_gen.go"}]}]}`
	)
	wantType := PublishDiagnosticsParams{
		URI:     DocumentURI("file:///path/to/diagnostics.go"),
		Version: 1,
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
							URI: uri.File("/path/to/diagnostics.go"),
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
						Message: "diagnostics.go",
					},
				},
			},
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          PublishDiagnosticsParams
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
			want             PublishDiagnosticsParams
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

				var got PublishDiagnosticsParams
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
