package lazyloader

import (
	"maps"
)

type docHeader interface {
	equal(other docHeader) bool
}

type goDocHeader struct {
	pkgName string
	imports map[string]struct{}
}

func (h *goDocHeader) equal(other docHeader) bool {
	o, ok := other.(*goDocHeader)
	if !ok || o == nil {
		return false
	}

	if h.pkgName != o.pkgName {
		return false
	}

	return maps.Equal(h.imports, o.imports)
}
