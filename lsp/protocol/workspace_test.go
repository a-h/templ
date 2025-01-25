// SPDX-FileCopyrightText: 2020 The Go Language Server Authors
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

func TestWorkspaceFolder(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"uri":"/path/to/workspace","name":"testWorkspace"}`
		wantInvalid = `{"uri":"/path/to/invalid","name":"invalidWorkspace"}`
	)
	wantType := WorkspaceFolder{
		URI:  "/path/to/workspace",
		Name: "testWorkspace",
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          WorkspaceFolder
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
			want             WorkspaceFolder
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

				var got WorkspaceFolder
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

func TestDidChangeWorkspaceFoldersParams(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"event":{"added":[{"uri":"/path/to/addedWorkspace","name":"testAddedWorkspace"},{"uri":"/path/to/addedWorkspace2","name":"testAddedWorkspace2"}],"removed":[{"uri":"/path/to/removedWorkspace","name":"testRemovedWorkspace"},{"uri":"/path/to/removedWorkspace2","name":"testRemovedWorkspace2"}]}}`
		wantInvalid = `{"event":{"added":[{"uri":"/path/to/addedInvalidWorkspace","name":"invalidAddedWorkspace"},{"uri":"/path/to/addedInvalidWorkspace2","name":"invalidAddedWorkspace2"}],"removed":[{"uri":"/path/to/removedInvalidWorkspace","name":"invalidRemovedWorkspace"},{"uri":"/path/to/removedInvalidWorkspace2","name":"invalidRemovedWorkspace2"}]}}`
	)
	wantType := DidChangeWorkspaceFoldersParams{
		Event: WorkspaceFoldersChangeEvent{
			Added: []WorkspaceFolder{
				{
					URI:  "/path/to/addedWorkspace",
					Name: "testAddedWorkspace",
				},
				{
					URI:  "/path/to/addedWorkspace2",
					Name: "testAddedWorkspace2",
				},
			},
			Removed: []WorkspaceFolder{
				{
					URI:  "/path/to/removedWorkspace",
					Name: "testRemovedWorkspace",
				},
				{
					URI:  "/path/to/removedWorkspace2",
					Name: "testRemovedWorkspace2",
				},
			},
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          DidChangeWorkspaceFoldersParams
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
			want             DidChangeWorkspaceFoldersParams
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

				var got DidChangeWorkspaceFoldersParams
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

func TestWorkspaceFoldersChangeEvent(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"added":[{"uri":"/path/to/addedWorkspace","name":"testAddedWorkspace"},{"uri":"/path/to/addedWorkspace2","name":"testAddedWorkspace2"}],"removed":[{"uri":"/path/to/removedWorkspace","name":"testRemovedWorkspace"},{"uri":"/path/to/removedWorkspace2","name":"testRemovedWorkspace2"}]}`
		wantInvalid = `{"added":[{"uri":"/path/to/addedInvalidWorkspace","name":"invalidAddedWorkspace"},{"uri":"/path/to/addedInvalidWorkspace2","name":"invalidAddedWorkspace2"}],"removed":[{"uri":"/path/to/removedInvalidWorkspace","name":"invalidRemovedWorkspace"},{"uri":"/path/to/removedInvalidWorkspace2","name":"invalidRemovedWorkspace2"}]}`
	)
	wantType := WorkspaceFoldersChangeEvent{
		Added: []WorkspaceFolder{
			{
				URI:  "/path/to/addedWorkspace",
				Name: "testAddedWorkspace",
			},
			{
				URI:  "/path/to/addedWorkspace2",
				Name: "testAddedWorkspace2",
			},
		},
		Removed: []WorkspaceFolder{
			{
				URI:  "/path/to/removedWorkspace",
				Name: "testRemovedWorkspace",
			},
			{
				URI:  "/path/to/removedWorkspace2",
				Name: "testRemovedWorkspace2",
			},
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          WorkspaceFoldersChangeEvent
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
			want             WorkspaceFoldersChangeEvent
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

				var got WorkspaceFoldersChangeEvent
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

func TestDidChangeConfigurationParams(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"settings":"testSettings"}`
		wantNilAll  = `{}`
		wantInvalid = `{"settings":"invalidSettings"}`
	)
	wantType := DidChangeConfigurationParams{
		Settings: "testSettings",
	}
	wantTypeNilAll := DidChangeConfigurationParams{}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          DidChangeConfigurationParams
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
			want             DidChangeConfigurationParams
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

				var got DidChangeConfigurationParams
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

func TestConfigurationParams(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"items":[{"scopeUri":"file:///path/to/test.go","section":"testSection"}]}`
		wantNilAll  = `{"items":[]}`
		wantInvalid = `{"items":[{"scopeUri":"file:///path/to/invalid.go","section":"invalidSection"}]}`
	)
	wantType := ConfigurationParams{
		Items: []ConfigurationItem{
			{
				ScopeURI: uri.File("/path/to/test.go"),
				Section:  "testSection",
			},
		},
	}
	wantTypeNilAll := ConfigurationParams{
		Items: []ConfigurationItem{},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          ConfigurationParams
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
			want             ConfigurationParams
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

				var got ConfigurationParams
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

func TestConfigurationItem(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"scopeUri":"file:///path/to/test.go","section":"testSection"}`
		wantNilAll  = `{}`
		wantInvalid = `{"scopeUri":"file:///path/to/invalid.go","section":"invalidSection"}`
	)
	wantType := ConfigurationItem{
		ScopeURI: uri.File("/path/to/test.go"),
		Section:  "testSection",
	}
	wantTypeNilAll := ConfigurationItem{}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          ConfigurationItem
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
			want             ConfigurationItem
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

				var got ConfigurationItem
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

func TestDidChangeWatchedFilesParams(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"changes":[{"type":2,"uri":"file:///path/to/test.go"}]}`
		wantNilAll  = `{}`
		wantInvalid = `{"changes":[{"type":3,"uri":"file:///path/to/invalid.go"}]}`
	)
	wantType := DidChangeWatchedFilesParams{
		Changes: []*FileEvent{
			{
				Type: FileChangeTypeChanged,
				URI:  uri.File("/path/to/test.go"),
			},
		},
	}
	wantTypeNilAll := DidChangeWatchedFilesParams{}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          DidChangeWatchedFilesParams
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
			want             DidChangeWatchedFilesParams
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

				var got DidChangeWatchedFilesParams
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

func TestFileEvent(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"type":2,"uri":"file:///path/to/test.go"}`
		wantInvalid = `{"type":3,"uri":"file:///path/to/invalid.go"}`
	)
	wantType := FileEvent{
		Type: FileChangeTypeChanged,
		URI:  uri.File("/path/to/test.go"),
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          FileEvent
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
			want             FileEvent
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

				var got FileEvent
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

func TestFileChangeType_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		k    FileChangeType
		want string
	}{
		{
			name: "Created",
			k:    FileChangeTypeCreated,
			want: "Created",
		},
		{
			name: "Changed",
			k:    FileChangeTypeChanged,
			want: "Changed",
		},
		{
			name: "Deleted",
			k:    FileChangeTypeDeleted,
			want: "Deleted",
		},
		{
			name: "Unknown",
			k:    FileChangeType(0),
			want: "0",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.k.String(); got != tt.want {
				t.Errorf("FileChangeType.String() = %v, want %v", tt.want, got)
			}
		})
	}
}

func TestDidChangeWatchedFilesRegistrationOptions(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"watchers":[{"globPattern":"*","kind":2}]}`
		wantNilAll  = `{"watchers":[{"globPattern":"*"}]}`
		wantInvalid = `{"watchers":[{"globPattern":"?","kind":1}]}`
	)
	wantType := DidChangeWatchedFilesRegistrationOptions{
		Watchers: []FileSystemWatcher{
			{
				GlobPattern: "*",
				Kind:        WatchKindChange,
			},
		},
	}
	wantTypeNilAll := DidChangeWatchedFilesRegistrationOptions{
		Watchers: []FileSystemWatcher{
			{
				GlobPattern: "*",
			},
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          DidChangeWatchedFilesRegistrationOptions
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
			want             DidChangeWatchedFilesRegistrationOptions
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

				var got DidChangeWatchedFilesRegistrationOptions
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

func TestWatchKind_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		k    WatchKind
		want string
	}{
		{
			name: "CreateWatch",
			k:    WatchKindCreate,
			want: "Create",
		},
		{
			name: "ChangeWatch",
			k:    WatchKindChange,
			want: "Change",
		},
		{
			name: "DeleteWatch",
			k:    WatchKindDelete,
			want: "Delete",
		},
		{
			name: "Unknown",
			k:    WatchKind(0),
			want: "0",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.k.String(); got != tt.want {
				t.Errorf("WatchKind.String() = %v, want %v", tt.want, got)
			}
		})
	}
}

func TestWorkspaceSymbolParams(t *testing.T) {
	t.Parallel()

	const (
		wantWorkDoneToken      = "156edea9-9d8d-422f-b7ee-81a84594afbb"
		wantPartialResultToken = "dd134d84-c134-4d7a-a2a3-f8af3ef4a568"
	)
	const (
		want        = `{"workDoneToken":"` + wantWorkDoneToken + `","partialResultToken":"` + wantPartialResultToken + `","query":"testQuery"}`
		wantInvalid = `{"workDoneToken":"` + wantPartialResultToken + `","partialResultToken":"` + wantWorkDoneToken + `","query":"invalidQuery"}`
	)
	wantType := WorkspaceSymbolParams{
		WorkDoneProgressParams: WorkDoneProgressParams{
			WorkDoneToken: NewProgressToken(wantWorkDoneToken),
		},
		PartialResultParams: PartialResultParams{
			PartialResultToken: NewProgressToken(wantPartialResultToken),
		},
		Query: "testQuery",
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          WorkspaceSymbolParams
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
			want             WorkspaceSymbolParams
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

				var got WorkspaceSymbolParams
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

func TestExecuteCommandParams(t *testing.T) {
	t.Parallel()

	const (
		wantWorkDoneToken    = "156edea9-9d8d-422f-b7ee-81a84594afbb"
		invalidWorkDoneToken = "dd134d84-c134-4d7a-a2a3-f8af3ef4a568"
	)
	const (
		want        = `{"workDoneToken":"` + wantWorkDoneToken + `","command":"testCommand","arguments":["testArguments"]}`
		wantNilAll  = `{"command":"testCommand"}`
		wantInvalid = `{"workDoneToken":"` + invalidWorkDoneToken + `","command":"invalidCommand","arguments":["invalidArguments"]}`
	)
	wantType := ExecuteCommandParams{
		WorkDoneProgressParams: WorkDoneProgressParams{
			WorkDoneToken: NewProgressToken(wantWorkDoneToken),
		},
		Command: "testCommand",
		Arguments: []any{
			"testArguments",
		},
	}
	wantTypeNilAll := ExecuteCommandParams{
		Command: "testCommand",
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          ExecuteCommandParams
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
			want             ExecuteCommandParams
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

				var got ExecuteCommandParams
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
			})
		}
	})
}

func TestExecuteCommandRegistrationOptions(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"commands":["testCommand","testCommand2"]}`
		wantInvalid = `{"commands":["invalidCommand"]}`
	)
	wantType := ExecuteCommandRegistrationOptions{
		Commands: []string{
			"testCommand",
			"testCommand2",
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          ExecuteCommandRegistrationOptions
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
			want             ExecuteCommandRegistrationOptions
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

				var got ExecuteCommandRegistrationOptions
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

func TestApplyWorkspaceEditParams(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"label":"testLabel","edit":{"changes":{"file:///path/to/basic.go":[{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"newText":"foo bar"}]},"documentChanges":[{"textDocument":{"uri":"file:///path/to/basic.go","version":10},"edits":[{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"newText":"foo bar"}]}]}}`
		wantNilAll  = `{"edit":{"changes":{"file:///path/to/basic.go":[{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"newText":"foo bar"}]},"documentChanges":[{"textDocument":{"uri":"file:///path/to/basic.go","version":10},"edits":[{"range":{"start":{"line":25,"character":1},"end":{"line":27,"character":3}},"newText":"foo bar"}]}]}}`
		wantInvalid = `{"label":"testLabel","edit":{"changes":{"file:///path/to/basic_gen.go":[{"range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"newText":"foo bar"}]},"documentChanges":[{"textDocument":{"uri":"file:///path/to/basic_gen.go","version":10},"edits":[{"range":{"start":{"line":2,"character":1},"end":{"line":3,"character":2}},"newText":"foo bar"}]}]}}`
	)
	wantType := ApplyWorkspaceEditParams{
		Label: "testLabel",
		Edit: WorkspaceEdit{
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
		},
	}
	wantTypeNilAll := ApplyWorkspaceEditParams{
		Edit: WorkspaceEdit{
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
		},
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          ApplyWorkspaceEditParams
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
			want             ApplyWorkspaceEditParams
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

				var got ApplyWorkspaceEditParams
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

func TestApplyWorkspaceEditResponse(t *testing.T) {
	t.Parallel()

	const (
		want        = `{"applied":true,"failureReason":"testFailureReason","failedChange":1}`
		wantInvalid = `{"applied":false}`
	)
	wantType := ApplyWorkspaceEditResponse{
		Applied:       true,
		FailureReason: "testFailureReason",
		FailedChange:  1,
	}

	t.Run("Marshal", func(t *testing.T) {
		tests := []struct {
			name           string
			field          ApplyWorkspaceEditResponse
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
			want             ApplyWorkspaceEditResponse
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

				var got ApplyWorkspaceEditResponse
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
