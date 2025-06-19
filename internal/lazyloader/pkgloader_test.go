package lazyloader

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/packages"
)

func TestGoPkgLoaderLoad(t *testing.T) {
	tests := []struct {
		name            string
		filename        string
		loader          goPkgLoader
		wantPkg         *packages.Package
		wantErrContains string
	}{
		{
			name:     "loadPackages returns error",
			filename: "/bad.go",
			loader: goPkgLoader{
				loadPackages: func(cfg *packages.Config, patterns ...string) ([]*packages.Package, error) {
					return nil, errors.New("load failed")
				},
			},
			wantErrContains: "load failed",
		},
		{
			name:     "returns multiple packages",
			filename: "/multi.go",
			loader: goPkgLoader{
				loadPackages: func(cfg *packages.Config, patterns ...string) ([]*packages.Package, error) {
					return []*packages.Package{
						{Name: "a"},
						{Name: "b"},
					}, nil
				},
			},
			wantErrContains: "expected 1 package, loaded 2 packages",
		},
		{
			name:     "returns zero packages",
			filename: "/empty.go",
			loader: goPkgLoader{
				loadPackages: func(cfg *packages.Config, patterns ...string) ([]*packages.Package, error) {
					return []*packages.Package{}, nil
				},
			},
			wantErrContains: "loaded no packages",
		},
		{
			name:     "returns package successfully",
			filename: "/main.go",
			loader: goPkgLoader{
				openDocSources: map[string]string{
					"file:///main.go": "package main\nfunc main() {}",
				},
				loadPackages: func(cfg *packages.Config, patterns ...string) ([]*packages.Package, error) {
					assert.Equal(t, "file=/main.go", patterns[0])
					assert.NotNil(t, cfg.Overlay)
					content, ok := cfg.Overlay["/main.go"]
					assert.True(t, ok)
					assert.Equal(t, string(content), "package main\nfunc main() {}")
					return []*packages.Package{
						{
							Name:    "main",
							PkgPath: "example.com/main",
							GoFiles: []string{"/main.go"},
						},
					}, nil
				},
			},
			wantPkg: &packages.Package{
				Name:    "main",
				PkgPath: "example.com/main",
				GoFiles: []string{"/main.go"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPkg, err := tt.loader.load(tt.filename)

			if tt.wantErrContains != "" {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErrContains)
				assert.Nil(t, gotPkg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantPkg.Name, gotPkg.Name)
				assert.Equal(t, tt.wantPkg.PkgPath, gotPkg.PkgPath)
				assert.Equal(t, tt.wantPkg.GoFiles, gotPkg.GoFiles)
			}
		})
	}
}
