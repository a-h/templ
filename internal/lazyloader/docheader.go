package lazyloader

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

	if len(h.imports) != len(o.imports) {
		return false
	}

	for imp := range h.imports {
		if _, ok := o.imports[imp]; !ok {
			return false
		}
	}

	return true
}
