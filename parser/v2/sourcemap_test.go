package parser

import (
	"sort"
	"testing"

	"github.com/a-h/parse"
	"github.com/google/go-cmp/cmp"
)

// Test data.
//
//	|  - | 0 1 2 3 4 5 6 7 8 9
//	|  - | - - - - - - - - -
//	|  0 |
//	|  1 |   a b c d e f g h i
//	|  2 |   j k l m n o
//	|  3 |   p q r s t u v
//	|  4 |
//	|  5 |   w x y
//	|  6 |   z
//	|  7 |   m u l t i
//	|  8 | l i n e
//	|  9 | m a t c h
//	| 10 | 生 日 快 乐
func pos(index, line, col int) parse.Position {
	return parse.Position{
		Index: index,
		Line:  line,
		Col:   col,
	}
}

func TestSourceMapPosition(t *testing.T) {
	sm := NewSourceMap()

	// Test that out of bounds requests don't return results.
	t.Run("out of bounds", func(t *testing.T) {
		actualTarget, ok := sm.TargetPositionFromSource(20, 10)
		if ok {
			t.Errorf("searching for a source position that's not in the map should not result in a target position, but got %v", actualTarget)
		}
		actualSource, ok := sm.SourcePositionFromTarget(20, 10)
		if ok {
			t.Errorf("searching for a target position that's not in the map should not result in a source position, but got %v", actualSource)
		}
	})

	var tests = []struct {
		name       string
		setup      func(sm *SourceMap)
		source     Position
		target     Position
		expectedOK bool
	}{
		{
			name: "searching within the map returns a result",
			setup: func(sm *SourceMap) {
				sm.Add(NewExpression("abc", pos(0, 1, 1), pos(2, 1, 3)),
					Range{From: NewPosition(0, 5, 1), To: NewPosition(0, 5, 3)})
			},
			source: NewPosition(0, 1, 1), // a
			target: NewPosition(0, 5, 1),
		},
		{
			name: "offsets within the match are handled",
			setup: func(sm *SourceMap) {
				sm.Add(NewExpression("abc", pos(0, 1, 1), pos(2, 1, 3)),
					Range{From: NewPosition(0, 5, 1), To: NewPosition(0, 5, 3)})
			},
			source: NewPosition(1, 1, 2), // b
			target: NewPosition(1, 5, 2),
		},
		{
			name: "the match that starts closest to the source is returned",
			setup: func(sm *SourceMap) {
				sm.Add(NewExpression("rst", pos(4, 3, 3), pos(7, 3, 6)),
					Range{From: NewPosition(0, 8, 6), To: NewPosition(0, 8, 9)})
				// s is inside rst.
				sm.Add(NewExpression("s", pos(-1, 3, 4), pos(-1, 3, 4)),
					Range{From: NewPosition(-1, 8, 7), To: NewPosition(-1, 8, 7)})
			},
			source: NewPosition(-1, 3, 4), // s
			target: NewPosition(-1, 8, 7),
		},
		{
			name: "the start line within a multiline match is detected",
			setup: func(sm *SourceMap) {
				// Multi-line match.
				sm.Add(NewExpression("multi\nline\nmatch", pos(0, 0, 0), pos(16, 2, 5)),
					Range{From: NewPosition(1, 1, 1), To: NewPosition(17, 3, 5)})
			},
			source: NewPosition(0, 0, 0), // m (ulti)
			target: NewPosition(1, 1, 1),
		},
		{
			name: "the middle line within a multiline match is detected",
			setup: func(sm *SourceMap) {
				// Multi-line match.
				sm.Add(NewExpression("multi\nline\nmatch", pos(0, 0, 0), pos(16, 2, 5)),
					Range{From: NewPosition(1, 1, 1), To: NewPosition(17, 3, 5)})
			},
			source: NewPosition(7, 1, 1), // (l) i (ne)
			target: NewPosition(8, 2, 1),
		},
		{
			name: "the final line within a multiline match is detected",
			setup: func(sm *SourceMap) {
				// Multi-line match.
				sm.Add(NewExpression("multi\nline\nmatch", pos(0, 0, 0), pos(16, 2, 5)),
					Range{From: NewPosition(1, 1, 1), To: NewPosition(17, 3, 5)})
			},
			source: NewPosition(11, 2, 0), // m (atch)
			target: NewPosition(12, 3, 0),
		},
		{
			name: "unicode characters are indexed correctly (sheng)",
			setup: func(sm *SourceMap) {
				sm.Add(NewExpression("生日快乐", pos(0, 10, 0), pos(12, 10, 4)),
					Range{From: NewPosition(1, 11, 1), To: NewPosition(13, 11, 5)})
			},
			source: NewPosition(0, 10, 0), // 生
			target: NewPosition(1, 11, 1),
		},
		{
			name: "unicode characters are indexed correctly (ri)",
			setup: func(sm *SourceMap) {
				sm.Add(NewExpression("生日快乐", pos(0, 10, 0), pos(12, 10, 4)),
					Range{From: NewPosition(1, 11, 1), To: NewPosition(13, 11, 5)})
			},
			source: NewPosition(3, 10, 3), // 日
			target: NewPosition(4, 11, 4),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSourceMap()
			tt.setup(sm)
			actualTarget, ok := sm.TargetPositionFromSource(uint32(tt.source.Line), uint32(tt.source.Col))
			if !ok {
				t.Errorf("TargetPositionFromSource: expected result from source %v, got no results", tt.source)
			}
			if diff := cmp.Diff(tt.target, actualTarget); diff != "" {
				lines := keys(sm.SourceLinesToTarget)
				sort.Slice(lines, func(i, j int) bool {
					return lines[i] < lines[j]
				})
				for _, lineIndex := range lines {
					cols := keys(sm.SourceLinesToTarget[lineIndex])
					sort.Slice(cols, func(i, j int) bool {
						return cols[i] < cols[j]
					})
					t.Error(lineIndex, cols)
				}
				t.Error("TargetPositionFromSource\n\n" + diff)
			}
			actualSource, ok := sm.SourcePositionFromTarget(actualTarget.Line, actualTarget.Col)
			if !ok {
				t.Fatalf("SourcePositionFromTarget: expected result, got no results")
			}
			if diff := cmp.Diff(tt.source, actualSource); diff != "" {
				t.Error("SourcePositionFromTarget\n\n" + diff)
			}
		})
	}
}

func keys[K comparable, V any](m map[K]V) (keys []K) {
	for k := range m {
		keys = append(keys, k)
	}
	return
}
