// SPDX-FileCopyrightText: 2020 The Go Language Server Authors
// SPDX-License-Identifier: BSD-3-Clause

package protocol

import (
	"testing"

	"encoding/json"
	"github.com/google/go-cmp/cmp"
)

func TestRegistration(t *testing.T) {
	t.Parallel()

	const (
		want           = `{"id":"1","method":"testMethod","registerOptions":{"foo":"bar"}}`
		wantInterfaces = `{"id":"1","method":"testMethod","registerOptions":["foo","bar"]}`
		wantNil        = `{"id":"1","method":"testMethod"}`
		wantInvalid    = `{"id":"2","method":"invalidMethod","registerOptions":{"baz":"qux"}}`
	)
	wantTypeStringInterface := Registration{
		ID:     "1",
		Method: "testMethod",
		RegisterOptions: map[string]any{
			"foo": "bar",
		},
	}
	wantTypeStringString := Registration{
		ID:     "1",
		Method: "testMethod",
		RegisterOptions: map[string]string{
			"foo": "bar",
		},
	}
	wantTypeInterfaces := Registration{
		ID:     "1",
		Method: "testMethod",
		RegisterOptions: []any{
			"foo",
			"bar",
		},
	}
	wantTypeNil := Registration{
		ID:     "1",
		Method: "testMethod",
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          Registration
			want           string
			wantMarshalErr bool
			wantErr        bool
		}{
			{
				name:           "ValidStringInterface",
				field:          wantTypeStringInterface,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidStringString",
				field:          wantTypeStringString,
				want:           want,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidInterfaces",
				field:          wantTypeInterfaces,
				want:           wantInterfaces,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "ValidNilAll",
				field:          wantTypeNil,
				want:           wantNil,
				wantMarshalErr: false,
				wantErr:        false,
			},
			{
				name:           "Invalid",
				field:          wantTypeStringInterface,
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
			want             Registration
			wantUnmarshalErr bool
			wantErr          bool
		}{
			{
				name:             "ValidStringInterface",
				field:            want,
				want:             wantTypeStringInterface,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "ValidInterfaces",
				field:            wantInterfaces,
				want:             wantTypeInterfaces,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "ValidNilAll",
				field:            wantNil,
				want:             wantTypeNil,
				wantUnmarshalErr: false,
				wantErr:          false,
			},
			{
				name:             "Invalid",
				field:            wantInvalid,
				want:             wantTypeStringInterface,
				wantUnmarshalErr: false,
				wantErr:          true,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				var got Registration
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

func TestRegistrationParams(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"registrations":[{"id":"1","method":"testMethod","registerOptions":{"foo":"bar"}}]}`
		wantNil     = `{"registrations":[{"id":"1","method":"testMethod"}]}`
		wantInvalid = `{"registrations":[{"id":"2","method":"invalidMethod","registerOptions":{"baz":"qux"}}]}`
	)
	wantType := RegistrationParams{
		Registrations: []Registration{
			{
				ID:     "1",
				Method: "testMethod",
				RegisterOptions: map[string]any{
					"foo": "bar",
				},
			},
		},
	}
	wantTypeNil := RegistrationParams{
		Registrations: []Registration{
			{
				ID:     "1",
				Method: "testMethod",
			},
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          RegistrationParams
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
		tests := []struct {
			name             string
			field            string
			want             RegistrationParams
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

				var got RegistrationParams
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

func TestTextDocumentRegistrationOptions(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"documentSelector":[{"language":"go","scheme":"file","pattern":"*.go"},{"language":"cpp","scheme":"untitled","pattern":"*.{cpp,hpp}"}]}`
		wantInvalid = `{"documentSelector":[{"language":"typescript","scheme":"file","pattern":"*.{ts,js}"},{"language":"c","scheme":"untitled","pattern":"*.{c,h}"}]}`
	)
	wantType := TextDocumentRegistrationOptions{
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
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          TextDocumentRegistrationOptions
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
			want             TextDocumentRegistrationOptions
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

				var got TextDocumentRegistrationOptions
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

func TestUnregistration(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"id":"1","method":"testMethod"}`
		wantInvalid = `{"id":"2","method":"invalidMethod"}`
	)
	wantType := Unregistration{
		ID:     "1",
		Method: "testMethod",
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          Unregistration
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
			want             Unregistration
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

				var got Unregistration
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

func TestUnregistrationParams(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"unregisterations":[{"id":"1","method":"testMethod"}]}`
		wantInvalid = `{"unregisterations":[{"id":"2","method":"invalidMethod"}]}`
	)
	wantType := UnregistrationParams{
		Unregisterations: []Unregistration{
			{
				ID:     "1",
				Method: "testMethod",
			},
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          UnregistrationParams
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
			want             UnregistrationParams
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

				var got UnregistrationParams
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
