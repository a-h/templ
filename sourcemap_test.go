package templ

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Test data.
// 0 | 1 2 3 4 5 6 7 8 9
// - - - - - - - - - - -
// 1 | a b c d e f g h i
// 2 | j k l m n o
// 3 | p q r s t u v
// 4 |
// 5 | w x y
// 6 | z

func TestSourceMapPosition(t *testing.T) {
	sm := NewSourceMap()
	// Set all of the indices to -1, they're not used in the lookup.
	sm.Add(NewExpression("abc", NewPositionFromValues(-1, 1, 1), NewPositionFromValues(-1, 1, 3)),
		NewRange(NewPositionFromValues(-1, 5, 1), NewPositionFromValues(-1, 5, 3)))
	sm.Add(NewExpression("rst", NewPositionFromValues(-1, 3, 3), NewPositionFromValues(-1, 3, 6)),
		NewRange(NewPositionFromValues(-1, 8, 6), NewPositionFromValues(-1, 8, 9)))
	// s is inside rst.
	sm.Add(NewExpression("s", NewPositionFromValues(-1, 3, 4), NewPositionFromValues(-1, 3, 4)),
		NewRange(NewPositionFromValues(-1, 8, 7), NewPositionFromValues(-1, 8, 7)))

	var tests = []struct {
		name       string
		data       *SourceMap
		source     Position
		expectedOK bool
		target     Position
	}{
		{
			name:       "searching for a position that's not in the map results in no position being returned",
			data:       sm,
			source:     NewPositionFromValues(100, 10, 10),
			expectedOK: false,
		},
		{
			name:       "searching within the map returns a result",
			data:       sm,
			source:     NewPositionFromValues(-1, 1, 1), // a
			expectedOK: true,
			target:     NewPositionFromValues(-1, 5, 1),
		},
		{
			name:       "offsets within the match are handled",
			data:       sm,
			source:     NewPositionFromValues(-1, 1, 2), // b
			expectedOK: true,
			target:     NewPositionFromValues(-1, 5, 2),
		},
		{
			name:       "the match that starts closes to the source is returned",
			data:       sm,
			source:     NewPositionFromValues(-1, 3, 4), // s
			expectedOK: true,
			target:     NewPositionFromValues(-1, 8, 7),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actualTarget, _, ok := tt.data.TargetPositionFromSource(tt.source)
			if tt.expectedOK != ok {
				t.Errorf("expected ok %v, but got %v", tt.expectedOK, ok)
				return
			}
			if diff := cmp.Diff(tt.target, actualTarget); diff != "" {
				t.Error(diff)
			}
		})
	}
}
