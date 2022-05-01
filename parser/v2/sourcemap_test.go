package parser

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

	// Test that out of bounds requests don't return results.
	actualTarget, _, ok := sm.TargetPositionFromSource(10, 10)
	if ok {
		t.Errorf("searching for a source position that's not in the map should not result in a target position, but got %v", actualTarget)
	}
	actualSource, _, ok := sm.SourcePositionFromTarget(10, 10)
	if ok {
		t.Errorf("searching for a target position that's not in the map should not result in a source position, but got %v", actualSource)
	}

	var tests = []struct {
		name   string
		source Position
		target Position
	}{
		{
			name:   "searching within the map returns a result",
			source: NewPositionFromValues(-1, 1, 1), // a
			target: NewPositionFromValues(-1, 5, 1),
		},
		{
			name:   "offsets within the match are handled",
			source: NewPositionFromValues(-1, 1, 2), // b
			target: NewPositionFromValues(-1, 5, 2),
		},
		{
			name:   "the match that starts closes to the source is returned",
			source: NewPositionFromValues(-1, 3, 4), // s
			target: NewPositionFromValues(-1, 8, 7),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actualTarget, _, ok := sm.TargetPositionFromSource(tt.source.Line, tt.source.Col)
			if !ok {
				t.Errorf("TargetPositionFromSource: expected result, got no results")
			}
			if diff := cmp.Diff(tt.target, actualTarget); diff != "" {
				t.Error("TargetPositionFromSource\n\n" + diff)
			}
			actualSource, _, ok := sm.SourcePositionFromTarget(actualTarget.Line, actualTarget.Col)
			if !ok {
				t.Errorf("SourcePositionFromTarget: expected result, got no results")
			}
			if diff := cmp.Diff(tt.source, actualSource); diff != "" {
				t.Error("SourcePositionFromTarget\n\n" + diff)
			}
		})
	}
}
