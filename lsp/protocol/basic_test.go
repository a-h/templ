// SPDX-FileCopyrightText: 2019 The Go Language Server Authors
// SPDX-License-Identifier: BSD-3-Clause

package protocol

import (
	"testing"

	"encoding/json"
	"github.com/google/go-cmp/cmp"

	"github.com/a-h/templ/lsp/uri"
)

func TestPosition(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"line":25,"character":1}`
		wantInvalid = `{"line":2,"character":0}`
	)
	wantType := Position{
		Line:      25,
		Character: 1,
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          Position
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
			want             Position
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

				var got Position
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

func TestRange(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"start":{"line":25,"character":1},"end":{"line":27,"character":3}}`
		wantInvalid = `{"start":{"line":2,"character":1},"end":{"line":3,"character":2}}`
	)
	wantType := Range{
		Start: Position{
			Line:      25,
			Character: 1,
		},
		End: Position{
			Line:      27,
			Character: 3,
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          Range
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
			want             Range
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

				var got Range
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

func TestLocation(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"uri":"file:///Users/gopher/go/src/github.com/a-h/templ/lsp/protocol/basic_test.go","range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}}}`
		wantInvalid = `{"uri":"file:///Users/gopher/go/src/github.com/a-h/templ/lsp/protocol/basic_test.go","range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}}}`
	)
	wantType := Location{
		URI: uri.File("/Users/gopher/go/src/github.com/a-h/templ/lsp/protocol/basic_test.go"),
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
			field          Location
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
			want             Location
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

				var got Location
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

func TestLocationLink(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"originSelectionRange":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"targetUri":"file:///path/to/test.go","targetRange":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"targetSelectionRange":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}}}`
		wantNil     = `{"targetUri":"file:///path/to/test.go","targetRange":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"targetSelectionRange":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}}}`
		wantInvalid = `{"originSelectionRange":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"targetUri":"file:///path/to/test.go","targetRange":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"targetSelectionRange":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}}}`
	)
	wantType := LocationLink{
		OriginSelectionRange: &Range{
			Start: Position{
				Line:      25,
				Character: 1,
			},
			End: Position{
				Line:      27,
				Character: 3,
			},
		},
		TargetURI: uri.File("/path/to/test.go"),
		TargetRange: Range{
			Start: Position{
				Line:      25,
				Character: 1,
			},
			End: Position{
				Line:      27,
				Character: 3,
			},
		},
		TargetSelectionRange: Range{
			Start: Position{
				Line: 25, Character: 1,
			},
			End: Position{
				Line:      27,
				Character: 3,
			},
		},
	}
	wantTypeNil := LocationLink{
		TargetURI: uri.File("/path/to/test.go"),
		TargetRange: Range{
			Start: Position{
				Line:      25,
				Character: 1,
			},
			End: Position{
				Line:      27,
				Character: 3,
			},
		},
		TargetSelectionRange: Range{
			Start: Position{
				Line: 25, Character: 1,
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
			field          LocationLink
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
				name:           "ValidNilOriginSelectionRange",
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
			want             LocationLink
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
				name:             "ValidNilOriginSelectionRange",
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

				var got LocationLink
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

func TestCodeDescription(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"href":"file:///path/to/test.go"}`
		wantInvalid = `{"href":"file:///path/to/invalid.go"}`
	)
	wantType := CodeDescription{
		Href: uri.File("/path/to/test.go"),
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          CodeDescription
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
			want             CodeDescription
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

				var got CodeDescription
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

func TestCommand(t *testing.T) {
	t.Parallel()

	const (
		want             = `{"title":"exec echo","command":"echo","arguments":["hello"]}`
		wantNilArguments = `{"title":"exec echo","command":"echo"}`
		wantInvalid      = `{"title":"exec echo","command":"true","arguments":["hello"]}`
	)
	wantType := Command{
		Title:     "exec echo",
		Command:   "echo",
		Arguments: []any{"hello"},
	}
	wantTypeNilArguments := Command{
		Title:   "exec echo",
		Command: "echo",
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          Command
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
				name:           "ValidNilArguments",
				field:          wantTypeNilArguments,
				want:           wantNilArguments,
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
			want             Command
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
				name:             "ValidNilArguments",
				field:            wantNilArguments,
				want:             wantTypeNilArguments,
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

				var got Command
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

func TestChangeAnnotation(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"label":"testLabel","needsConfirmation":true,"description":"testDescription"}`
		wantNilAll  = `{"label":"testLabel"}`
		wantInvalid = `{"label":"invalidLabel","needsConfirmation":false,"description":"invalidDescription"}`
	)
	wantType := ChangeAnnotation{
		Label:             "testLabel",
		NeedsConfirmation: true,
		Description:       "testDescription",
	}
	wantTypeNilAll := ChangeAnnotation{
		Label: "testLabel",
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          ChangeAnnotation
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
				name:           "ValidNilArguments",
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
			want             ChangeAnnotation
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
				name:             "ValidNilArguments",
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

				var got ChangeAnnotation
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

func TestAnnotatedTextEdit(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"newText":"foo bar","annotationId":"testAnnotationIdentifier"}`
		wantInvalid = `{"range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"newText":"foo bar","annotationId":"invalidAnnotationIdentifier"}`
	)
	wantType := AnnotatedTextEdit{
		TextEdit: TextEdit{
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
		AnnotationID: ChangeAnnotationIdentifier("testAnnotationIdentifier"),
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          AnnotatedTextEdit
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
			want             AnnotatedTextEdit
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

				var got AnnotatedTextEdit
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

func TestTextEdit(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"newText":"foo bar"}`
		wantInvalid = `{"range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"newText":"foo bar"}`
	)
	wantType := TextEdit{
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
	}
	wantInvalidType := TextEdit{
		Range: Range{
			Start: Position{
				Line:      2,
				Character: 1,
			},
			End: Position{
				Line:      3,
				Character: 2,
			},
		},
		NewText: "foo bar",
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          TextEdit
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
				name: "Invalid",
				field: TextEdit{
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
			want             TextEdit
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
				field:            want,
				want:             wantInvalidType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got TextEdit
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

func TestTextDocumentEdit(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"textDocument":{"uri":"file:///path/to/basic.go","version":10},"edits":[{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"newText":"foo bar"}]}`
		wantInvalid = `{"textDocument":{"uri":"file:///path/to/basic_gen.go","version":10},"edits":[{"range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"newText":"foo bar"}]}`
	)
	wantType := TextDocumentEdit{
		TextDocument: OptionalVersionedTextDocumentIdentifier{
			TextDocumentIdentifier: TextDocumentIdentifier{
				URI: "file:///path/to/basic.go",
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
	}
	wantInvalidType := TextDocumentEdit{
		TextDocument: OptionalVersionedTextDocumentIdentifier{
			TextDocumentIdentifier: TextDocumentIdentifier{
				URI: "file:///path/to/basic.go",
			},
			Version: NewVersion(int32(10)),
		},
		Edits: []TextEdit{
			{
				Range: Range{
					Start: Position{
						Line:      2,
						Character: 1,
					},
					End: Position{
						Line:      3,
						Character: 2,
					},
				},
				NewText: "foo bar",
			},
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          TextDocumentEdit
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
			want             TextDocumentEdit
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
				field:            want,
				want:             wantInvalidType,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got TextDocumentEdit
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

func TestCreateFileOptions(t *testing.T) {
	t.Parallel()

	const (
		want                  = `{"overwrite":true,"ignoreIfExists":true}`
		wantNilIgnoreIfExists = `{"overwrite":true}`
		wantNilOverwrite      = `{"ignoreIfExists":true}`
		wantValidNilAll       = `{}`
		wantInvalid           = `{"overwrite":false,"ignoreIfExists":false}`
	)
	wantType := CreateFileOptions{
		Overwrite:      true,
		IgnoreIfExists: true,
	}
	wantTypeNilOverwrite := CreateFileOptions{
		IgnoreIfExists: true,
	}
	wantTypeNilIgnoreIfExists := CreateFileOptions{
		Overwrite: true,
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          CreateFileOptions
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
				name:           "ValidNilOverwrite",
				field:          wantTypeNilIgnoreIfExists,
				want:           wantNilIgnoreIfExists,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidNilIgnoreIfExists",
				field:          wantTypeNilOverwrite,
				want:           wantNilOverwrite,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidNilAll",
				field:          CreateFileOptions{},
				want:           wantValidNilAll,
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
			want             CreateFileOptions
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:  "Valid",
				field: `{"overwrite":true,"ignoreIfExists":true}`,
				want: CreateFileOptions{
					Overwrite:      true,
					IgnoreIfExists: true,
				},
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:  "ValidNilOverwrite",
				field: `{"ignoreIfExists":true}`,
				want: CreateFileOptions{
					IgnoreIfExists: true,
				},
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:  "ValidNilIgnoreIfExists",
				field: `{"overwrite":true}`,
				want: CreateFileOptions{
					Overwrite: true,
				},
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "ValidNilAll",
				field:            `{}`,
				want:             CreateFileOptions{},
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:  "Invalid",
				field: `{"overwrite":true,"ignoreIfExists":true}`,
				want: CreateFileOptions{
					Overwrite:      false,
					IgnoreIfExists: false,
				},
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got CreateFileOptions
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

func TestCreateFile(t *testing.T) {
	t.Parallel()

	const (
		want           = `{"kind":"create","uri":"file:///path/to/basic.go","options":{"overwrite":true,"ignoreIfExists":true},"annotationId":"testAnnotationIdentifier"}`
		wantNilOptions = `{"kind":"create","uri":"file:///path/to/basic.go"}`
		wantInvalid    = `{"kind":"create","uri":"file:///path/to/basic_gen.go","options":{"overwrite":false,"ignoreIfExists":false},"annotationId":"invalidAnnotationIdentifier"}`
	)
	wantType := CreateFile{
		Kind: "create",
		URI:  uri.File("/path/to/basic.go"),
		Options: &CreateFileOptions{
			Overwrite:      true,
			IgnoreIfExists: true,
		},
		AnnotationID: ChangeAnnotationIdentifier("testAnnotationIdentifier"),
	}
	wantTypeNilOptions := CreateFile{
		Kind: "create",
		URI:  uri.File("/path/to/basic.go"),
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          CreateFile
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
				name:           "ValidNilOptions",
				field:          wantTypeNilOptions,
				want:           wantNilOptions,
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
			want             CreateFile
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
				name:             "ValidNilOptions",
				field:            wantNilOptions,
				want:             wantTypeNilOptions,
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

				var got CreateFile
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

func TestRenameFileOptions(t *testing.T) {
	t.Parallel()

	const (
		want                  = `{"overwrite":true,"ignoreIfExists":true}`
		wantNilOverwrite      = `{"ignoreIfExists":true}`
		wantNilIgnoreIfExists = `{"overwrite":true}`
		wantNilAll            = `{}`
		wantInvalid           = `{"overwrite":false,"ignoreIfExists":false}`
	)
	wantType := RenameFileOptions{
		Overwrite:      true,
		IgnoreIfExists: true,
	}
	wantTypeNilOverwrite := RenameFileOptions{
		IgnoreIfExists: true,
	}
	wantTypeNilIgnoreIfExists := RenameFileOptions{
		Overwrite: true,
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          RenameFileOptions
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
				name:           "ValidNilOverwrite",
				field:          wantTypeNilOverwrite,
				want:           wantNilOverwrite,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidNilIgnoreIfExists",
				field:          wantTypeNilIgnoreIfExists,
				want:           wantNilIgnoreIfExists,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidNilAll",
				field:          RenameFileOptions{},
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
			want             RenameFileOptions
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:  "Valid",
				field: `{"overwrite":true,"ignoreIfExists":true}`,
				want: RenameFileOptions{
					Overwrite:      true,
					IgnoreIfExists: true,
				},
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:  "ValidNilOverwrite",
				field: `{"ignoreIfExists":true}`,
				want: RenameFileOptions{
					IgnoreIfExists: true,
				},
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:  "ValidNilIgnoreIfExists",
				field: `{"overwrite":true}`,
				want: RenameFileOptions{
					Overwrite: true,
				},
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "ValidNilAll",
				field:            `{}`,
				want:             RenameFileOptions{},
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:  "Invalid",
				field: `{"overwrite":true,"ignoreIfExists":true}`,
				want: RenameFileOptions{
					Overwrite:      false,
					IgnoreIfExists: false,
				},
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got RenameFileOptions
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

func TestRenameFile(t *testing.T) {
	t.Parallel()

	const (
		want           = `{"kind":"rename","oldUri":"file:///path/to/old.go","newUri":"file:///path/to/new.go","options":{"overwrite":true,"ignoreIfExists":true},"annotationId":"testAnnotationIdentifier"}`
		wantNilOptions = `{"kind":"rename","oldUri":"file:///path/to/old.go","newUri":"file:///path/to/new.go"}`
		wantInvalid    = `{"kind":"rename","oldUri":"file:///path/to/old2.go","newUri":"file:///path/to/new2.go","options":{"overwrite":false,"ignoreIfExists":false},"annotationId":"invalidAnnotationIdentifier"}`
	)
	wantType := RenameFile{
		Kind:   "rename",
		OldURI: uri.File("/path/to/old.go"),
		NewURI: uri.File("/path/to/new.go"),
		Options: &RenameFileOptions{
			Overwrite:      true,
			IgnoreIfExists: true,
		},
		AnnotationID: ChangeAnnotationIdentifier("testAnnotationIdentifier"),
	}
	wantTypeNilOptions := RenameFile{
		Kind:   "rename",
		OldURI: uri.File("/path/to/old.go"),
		NewURI: uri.File("/path/to/new.go"),
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          RenameFile
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
				name:           "ValidNilOptions",
				field:          wantTypeNilOptions,
				want:           wantNilOptions,
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
			want             RenameFile
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
				name:             "ValidNilOptions",
				field:            wantNilOptions,
				want:             wantTypeNilOptions,
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

				var got RenameFile
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

func TestDeleteFileOptions(t *testing.T) {
	t.Parallel()

	const (
		want                    = `{"recursive":true,"ignoreIfNotExists":true}`
		wantNilRecursive        = `{"ignoreIfNotExists":true}`
		wantNiIgnoreIfNotExists = `{"recursive":true}`
		wantNilAll              = `{}`
		wantInvalid             = `{"recursive":false,"ignoreIfNotExists":false}`
	)
	wantType := DeleteFileOptions{
		Recursive:         true,
		IgnoreIfNotExists: true,
	}
	wantTypeNilRecursive := DeleteFileOptions{
		IgnoreIfNotExists: true,
	}
	wantTypeNiIgnoreIfNotExists := DeleteFileOptions{
		Recursive: true,
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          DeleteFileOptions
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
				name:           "ValidNilRecursive",
				field:          wantTypeNilRecursive,
				want:           wantNilRecursive,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidNiIgnoreIfNotExists",
				field:          wantTypeNiIgnoreIfNotExists,
				want:           wantNiIgnoreIfNotExists,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidNilAll",
				field:          DeleteFileOptions{},
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
			want             DeleteFileOptions
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
				name:             "ValidNilRecursive",
				field:            wantNilRecursive,
				want:             wantTypeNilRecursive,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "ValidNilIgnoreIfNotExists",
				field:            wantNiIgnoreIfNotExists,
				want:             wantTypeNiIgnoreIfNotExists,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "ValidNilAll",
				field:            wantNilAll,
				want:             DeleteFileOptions{},
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

				var got DeleteFileOptions
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

func TestDeleteFile(t *testing.T) {
	t.Parallel()

	const (
		want           = `{"kind":"delete","uri":"file:///path/to/delete.go","options":{"recursive":true,"ignoreIfNotExists":true},"annotationId":"testAnnotationIdentifier"}`
		wantNilOptions = `{"kind":"delete","uri":"file:///path/to/delete.go"}`
		wantInvalid    = `{"kind":"delete","uri":"file:///path/to/delete2.go","options":{"recursive":false,"ignoreIfNotExists":false},"annotationId":"invalidAnnotationIdentifier"}`
	)
	wantType := DeleteFile{
		Kind: "delete",
		URI:  uri.File("/path/to/delete.go"),
		Options: &DeleteFileOptions{
			Recursive:         true,
			IgnoreIfNotExists: true,
		},
		AnnotationID: ChangeAnnotationIdentifier("testAnnotationIdentifier"),
	}
	wantTypeNilOptions := DeleteFile{
		Kind: "delete",
		URI:  uri.File("/path/to/delete.go"),
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          DeleteFile
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
				name:           "ValidNilOptions",
				field:          wantTypeNilOptions,
				want:           wantNilOptions,
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
			want             DeleteFile
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
				name:             "ValidNilOptions",
				field:            wantNilOptions,
				want:             wantTypeNilOptions,
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

				var got DeleteFile
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

func TestWorkspaceEdit(t *testing.T) {
	t.Parallel()

	const (
		want                   = `{"changes":{"file:///path/to/basic.go":[{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"newText":"foo bar"}]},"documentChanges":[{"textDocument":{"uri":"file:///path/to/basic.go","version":10},"edits":[{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"newText":"foo bar"}]}],"changeAnnotations":{"testAnnotationIdentifier":{"label":"testLabel","needsConfirmation":true,"description":"testDescription"}}}`
		wantNilChanges         = `{"documentChanges":[{"textDocument":{"uri":"file:///path/to/basic.go","version":10},"edits":[{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"newText":"foo bar"}]}]}`
		wantNilDocumentChanges = `{"changes":{"file:///path/to/basic.go":[{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"newText":"foo bar"}]}}`
		wantInvalid            = `{"changes":{"file:///path/to/basic_gen.go":[{"range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"newText":"foo bar"}]},"documentChanges":[{"textDocument":{"uri":"file:///path/to/basic_gen.go","version":10},"edits":[{"range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"newText":"foo bar"}]}]}`
	)
	wantType := WorkspaceEdit{
		Changes: map[uri.URI][]TextEdit{
			uri.File("/path/to/basic.go"): {
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
						URI: uri.File("/path/to/basic.go"),
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
		ChangeAnnotations: map[ChangeAnnotationIdentifier]ChangeAnnotation{
			ChangeAnnotationIdentifier("testAnnotationIdentifier"): {
				Label:             "testLabel",
				NeedsConfirmation: true,
				Description:       "testDescription",
			},
		},
	}
	wantTypeNilChanges := WorkspaceEdit{
		DocumentChanges: []TextDocumentEdit{
			{
				TextDocument: OptionalVersionedTextDocumentIdentifier{
					TextDocumentIdentifier: TextDocumentIdentifier{
						URI: uri.File("/path/to/basic.go"),
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
	}
	wantTypeNilDocumentChanges := WorkspaceEdit{
		Changes: map[uri.URI][]TextEdit{
			uri.File("/path/to/basic.go"): {
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
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          WorkspaceEdit
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
				name:           "ValidNilChanges",
				field:          wantTypeNilChanges,
				want:           wantNilChanges,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidNilDocumentChanges",
				field:          wantTypeNilDocumentChanges,
				want:           wantNilDocumentChanges,
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
			want             WorkspaceEdit
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
				name:             "ValidNilChanges",
				field:            wantNilChanges,
				want:             wantTypeNilChanges,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "ValidNilDocumentChanges",
				field:            wantNilDocumentChanges,
				want:             wantTypeNilDocumentChanges,
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

				var got WorkspaceEdit
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

func TestTextDocumentIdentifier(t *testing.T) {
	t.Parallel()

	const (
		want             = `{"uri":"file:///path/to/basic.go"}`
		wantInvalid      = `{"uri":"file:///path/to/unknown.go"}`
		wantInvalidEmpty = `{}`
	)
	wantType := TextDocumentIdentifier{
		URI: uri.File("/path/to/basic.go"),
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          TextDocumentIdentifier
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
			{
				name:           "InvalidEmpty",
				field:          TextDocumentIdentifier{},
				want:           wantInvalidEmpty,
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
			want             TextDocumentIdentifier
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

				var got TextDocumentIdentifier
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

func TestTextDocumentItem(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"uri":"file:///path/to/basic.go","languageId":"go","version":10,"text":"Go Language"}`
		wantInvalid = `{"uri":"file:///path/to/basic_gen.go","languageId":"cpp","version":10,"text":"C++ Language"}`
	)
	wantType := TextDocumentItem{
		URI:        uri.File("/path/to/basic.go"),
		LanguageID: GoLanguage,
		Version:    int32(10),
		Text:       "Go Language",
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          TextDocumentItem
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
			want             TextDocumentItem
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

				var got TextDocumentItem
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

func TestToLanguageIdentifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ft   string
		want LanguageIdentifier
	}{
		{
			name: "Go",
			ft:   "go",
			want: GoLanguage,
		},
		{
			name: "C",
			ft:   "c",
			want: CLanguage,
		},
		{
			name: "lsif",
			ft:   "lsif",
			want: LanguageIdentifier("lsif"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := ToLanguageIdentifier(tt.ft); got != tt.want {
				t.Errorf("ToLanguageIdentifier(%v) = %v, want %v", tt.ft, tt.want, got)
			}
		})
	}
}

func TestVersionedTextDocumentIdentifier(t *testing.T) {
	t.Parallel()

	const (
		want            = `{"uri":"file:///path/to/basic.go","version":10}`
		wantZeroVersion = `{"uri":"file:///path/to/basic.go","version":0}`
		wantInvalid     = `{"uri":"file:///path/to/basic_gen.go","version":50}`
	)
	wantType := VersionedTextDocumentIdentifier{
		TextDocumentIdentifier: TextDocumentIdentifier{
			URI: uri.File("/path/to/basic.go"),
		},
		Version: int32(10),
	}
	wantTypeNullVersion := VersionedTextDocumentIdentifier{
		TextDocumentIdentifier: TextDocumentIdentifier{
			URI: uri.File("/path/to/basic.go"),
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          VersionedTextDocumentIdentifier
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
				name:           "ValidNullVersion",
				field:          wantTypeNullVersion,
				want:           wantZeroVersion,
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
			want             VersionedTextDocumentIdentifier
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
				name:             "ValidNullVersion",
				field:            wantZeroVersion,
				want:             wantTypeNullVersion,
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

				var got VersionedTextDocumentIdentifier
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

func TestOptionalVersionedTextDocumentIdentifier(t *testing.T) {
	t.Parallel()

	const (
		want            = `{"uri":"file:///path/to/basic.go","version":10}`
		wantNullVersion = `{"uri":"file:///path/to/basic.go","version":null}`
		wantInvalid     = `{"uri":"file:///path/to/basic_gen.go","version":50}`
	)
	wantType := OptionalVersionedTextDocumentIdentifier{
		TextDocumentIdentifier: TextDocumentIdentifier{
			URI: uri.File("/path/to/basic.go"),
		},
		Version: NewVersion(10),
	}
	wantTypeNullVersion := OptionalVersionedTextDocumentIdentifier{
		TextDocumentIdentifier: TextDocumentIdentifier{
			URI: uri.File("/path/to/basic.go"),
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          OptionalVersionedTextDocumentIdentifier
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
				name:           "ValidNullVersion",
				field:          wantTypeNullVersion,
				want:           wantNullVersion,
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
			want             OptionalVersionedTextDocumentIdentifier
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
				name:             "ValidNullVersion",
				field:            wantNullVersion,
				want:             wantTypeNullVersion,
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

				var got OptionalVersionedTextDocumentIdentifier
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

func TestTextDocumentPositionParams(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"textDocument":{"uri":"file:///path/to/basic.go"},"position":{"line":25,"character":1}}`
		wantInvalid = `{"textDocument":{"uri":"file:///path/to/basic_gen.go"},"position":{"line":2,"character":1}}`
	)
	wantType := TextDocumentPositionParams{
		TextDocument: TextDocumentIdentifier{
			URI: uri.File("/path/to/basic.go"),
		},
		Position: Position{
			Line:      25,
			Character: 1,
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          TextDocumentPositionParams
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
			want             TextDocumentPositionParams
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

				var got TextDocumentPositionParams
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

func TestDocumentFilter(t *testing.T) {
	t.Parallel()

	const (
		want            = `{"language":"go","scheme":"file","pattern":"*"}`
		wantNilLanguage = `{"scheme":"file","pattern":"*"}`
		wantNilScheme   = `{"language":"go","pattern":"*"}`
		wantNilPattern  = `{"language":"go","scheme":"file"}`
		wantNilAll      = `{}`
		wantInvalid     = `{"language":"typescript","scheme":"file","pattern":"?"}`
	)
	wantType := DocumentFilter{
		Language: "go",
		Scheme:   "file",
		Pattern:  "*",
	}
	wantTypeNilLanguage := DocumentFilter{
		Scheme:  "file",
		Pattern: "*",
	}
	wantTypeNilScheme := DocumentFilter{
		Language: "go",
		Pattern:  "*",
	}
	wantTypeNilPattern := DocumentFilter{
		Language: "go",
		Scheme:   "file",
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          DocumentFilter
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
				name:           "ValidNilLanguage",
				field:          wantTypeNilLanguage,
				want:           wantNilLanguage,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidNilScheme",
				field:          wantTypeNilScheme,
				want:           wantNilScheme,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidNilPattern",
				field:          wantTypeNilPattern,
				want:           wantNilPattern,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidNilAll",
				field:          DocumentFilter{},
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
			want             DocumentFilter
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
				name:             "ValidNilLanguage",
				field:            wantNilLanguage,
				want:             wantTypeNilLanguage,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "ValidNilScheme",
				field:            wantNilScheme,
				want:             wantTypeNilScheme,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "ValidNilPattern",
				field:            wantNilPattern,
				want:             wantTypeNilPattern,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "ValidNilAll",
				field:            wantNilAll,
				want:             DocumentFilter{},
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

				var got DocumentFilter
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

func TestDocumentSelector(t *testing.T) {
	t.Parallel()

	const (
		want        = `[{"language":"go","scheme":"file","pattern":"*.go"},{"language":"cpp","scheme":"untitled","pattern":"*.{cpp,hpp}"}]`
		wantInvalid = `[{"language":"typescript","scheme":"file","pattern":"*.{ts,js}"},{"language":"c","scheme":"untitled","pattern":"*.{c,h}"}]`
	)
	wantType := DocumentSelector{
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
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          DocumentSelector
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
			want             DocumentSelector
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

				var got DocumentSelector
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

func TestMarkupContent(t *testing.T) {
	t.Parallel()

	const (
		want        = "{\"kind\":\"markdown\",\"value\":\"# Header\\nSome text\\n```typescript\\nsomeCode();\\n'```\\n\"}"
		wantInvalid = "{\"kind\":\"plaintext\",\"value\":\"Header\\nSome text\\ntypescript\\nsomeCode();\\n\"}"
	)
	wantType := MarkupContent{
		Kind:  Markdown,
		Value: "# Header\nSome text\n```typescript\nsomeCode();\n'```\n",
	}

	t.Run("Marshal", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name           string
			field          MarkupContent
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
			want             MarkupContent
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

				var got MarkupContent
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
