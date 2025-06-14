package lazyloader

import (
	"fmt"
	"unsafe"

	"github.com/a-h/templ/lsp/uri"
	"golang.org/x/tools/go/packages"
)

type pkgLoader interface {
	load(file string) (*packages.Package, error)
}

type goPkgLoader struct {
	openDocSources map[string]string
	loadPackages   func(cfg *packages.Config, patterns ...string) ([]*packages.Package, error)
}

func (l *goPkgLoader) load(file string) (*packages.Package, error) {
	pkgs, err := l.loadPackages(
		&packages.Config{
			Mode:    packages.NeedName | packages.NeedFiles | packages.NeedImports | packages.NeedDeps,
			Overlay: l.prepareOverlay(),
		},
		"file="+file,
	)

	if err != nil {
		return nil, err
	}

	if len(pkgs) != 1 {
		return nil, fmt.Errorf("expected 1 package, got %d packages", len(pkgs))
	}

	return pkgs[0], nil
}

func (l *goPkgLoader) prepareOverlay() map[string][]byte {
	overlay := make(map[string][]byte, len(l.openDocSources))
	for fileURI, source := range l.openDocSources {
		overlay[uri.New(fileURI).Filename()] = unsafe.Slice(unsafe.StringData(source), len(source))
	}
	return overlay
}
