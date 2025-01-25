// Copyright 2019 The Go Language Server Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package uri

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFile(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    URI
		wantErr bool
	}{
		{
			name:    "ValidFileScheme",
			path:    "/users/me/c#-projects/",
			want:    URI(FileScheme + hierPart + "/users/me/c%23-projects"),
			wantErr: false,
		},
		{
			name:    "Invalid",
			path:    "users-me-c#-projects",
			want:    URI(""),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if diff := cmp.Diff(File(tt.path), tt.want); (diff != "") != tt.wantErr {
				t.Errorf("%s: (-got, +want)\n%s", tt.name, diff)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want URI
	}{
		{
			name: "ValidFileScheme",
			s:    "file://code.visualstudio.com/docs/extensions/overview.md",
			want: URI(FileScheme + hierPart + "/docs/extensions/overview.md"),
		},
		{
			name: "ValidHTTPScheme",
			s:    "http://code.visualstudio.com/docs/extensions/overview#frag",
			want: URI(HTTPScheme + hierPart + "code.visualstudio.com/docs/extensions/overview#frag"),
		},
		{
			name: "ValidHTTPSScheme",
			s:    "https://code.visualstudio.com/docs/extensions/overview#frag",
			want: URI(HTTPSScheme + hierPart + "code.visualstudio.com/docs/extensions/overview#frag"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Parse(tt.s)
			if err != nil {
				t.Error(err)
				return
			}

			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("%s: (-got, +want)\n%s", tt.name, diff)
			}
		})
	}
}

func TestFrom(t *testing.T) {
	type args struct {
		scheme    string
		authority string
		path      string
		query     string
		fragment  string
	}
	tests := []struct {
		name string
		args args
		want URI
	}{
		{
			name: "ValidFileScheme",
			args: args{
				scheme:    "file",
				authority: "example.com",
				path:      "/over/there",
				query:     "name=ferret",
				fragment:  "nose",
			},
			want: URI(FileScheme + hierPart + "/over/there"),
		},
		{
			name: "ValidHTTPScheme",
			args: args{
				scheme:    "http",
				authority: "example.com:8042",
				path:      "/over/there",
				query:     "name=ferret",
				fragment:  "nose",
			},
			want: URI(HTTPScheme + hierPart + "example.com:8042/over/there?name%3Dferret#nose"),
		},
		{
			name: "ValidHTTPSScheme",
			args: args{
				scheme:    "https",
				authority: "example.com:8042",
				path:      "/over/there",
				query:     "name=ferret",
				fragment:  "nose",
			},
			want: URI(HTTPSScheme + hierPart + "example.com:8042/over/there?name%3Dferret#nose"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if diff := cmp.Diff(From(tt.args.scheme, tt.args.authority, tt.args.path, tt.args.query, tt.args.fragment), tt.want); diff != "" {
				t.Errorf("%s: (-got, +want)\n%s", tt.name, diff)
			}
		})
	}
}
