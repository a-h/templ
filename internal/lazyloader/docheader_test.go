package lazyloader

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoDocHeaderEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        *goDocHeader
		b        docHeader
		expected bool
	}{
		{
			name: "other header is not a goDocHeader",
			a: &goDocHeader{
				pkgName: "a",
				imports: map[string]struct{}{"fmt": {}},
			},
			b: mockDocHeader{},
		},
		{
			name: "nil other header",
			a: &goDocHeader{
				pkgName: "a",
				imports: map[string]struct{}{"fmt": {}},
			},
			b: nil,
		},
		{
			name: "different package names",
			a: &goDocHeader{
				pkgName: "a",
				imports: map[string]struct{}{"fmt": {}},
			},
			b: &goDocHeader{
				pkgName: "b",
				imports: map[string]struct{}{"fmt": {}},
			},
		},
		{
			name: "different number of imports",
			a: &goDocHeader{
				pkgName: "a",
				imports: map[string]struct{}{"fmt": {}, "strings": {}},
			},
			b: &goDocHeader{
				pkgName: "a",
				imports: map[string]struct{}{"fmt": {}},
			},
		},
		{
			name: "different import keys",
			a: &goDocHeader{
				pkgName: "a",
				imports: map[string]struct{}{"fmt": {}, "bytes": {}},
			},
			b: &goDocHeader{
				pkgName: "a",
				imports: map[string]struct{}{"fmt": {}, "strings": {}},
			},
		},
		{
			name: "equal headers with same pkg and imports",
			a: &goDocHeader{
				pkgName: "mypkg",
				imports: map[string]struct{}{
					"fmt":     {},
					"strings": {},
				},
			},
			b: &goDocHeader{
				pkgName: "mypkg",
				imports: map[string]struct{}{
					"strings": {},
					"fmt":     {},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.a.equal(tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

type mockDocHeader struct{}

func (m mockDocHeader) equal(_ docHeader) bool {
	return false
}
