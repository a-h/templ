package templ

import (
	"fmt"

	"github.com/a-h/lexical/parse"
)

// Source mapping to map from the source code of the template to the
// in-memory representation.
type Position struct {
	Index int64
	Line  int
	Col   int
}

func (p Position) String() string {
	return fmt.Sprintf("line %d, col %d (index %d)", p.Line, p.Col, p.Index)
}

// NewPositionFromInput creates a position from a parse input.
func NewPositionFromInput(pi parse.Input) Position {
	l, c := pi.Position()
	return Position{
		Index: pi.Index(),
		Line:  l,
		Col:   c,
	}
}

// ItemRange is a record of a template Item, along with its start and end positions.
type ItemRange struct {
	Item Item
	From Position
	To   Position
}

type SourceRangeToItemLookup []ItemRange

// Add an item to the lookup.
func (ir SourceRangeToItemLookup) Add(item Item, from, to Position) (updatedFrom Position) {
	ir = append(ir, ItemRange{
		Item: item,
		From: from,
		To:   to,
	})
	return from
}

func (sril SourceRangeToItemLookup) LookupByIndex(index int64) (ir ItemRange, ok bool) {
	//TODO: Update the design so it's not looping through all the items!
	for _, cc := range sril {
		if cc.From.Index >= index && cc.To.Index <= index {
			return cc, true
		}
	}
	return ItemRange{}, false
}

func (sril SourceRangeToItemLookup) LookupByLineCol(line, col int) (ir ItemRange, ok bool) {
	for _, cc := range sril {
		// Single line.
		if cc.From.Line == cc.To.Line && cc.To.Line == line && ((col >= cc.From.Col && col <= cc.To.Col) ||
			(col <= cc.From.Col && col >= cc.To.Col)) {
			return cc, true
		}
		// Upwards multiline.
		if cc.From.Line > line && (cc.To.Line < line || cc.To.Line == line && col >= cc.To.Col) {
			return cc, true
		}
		// Downwards multiline.
		if cc.From.Line < line && (cc.To.Line > line || cc.To.Line == line && col <= cc.To.Col) {
			return cc, true
		}
	}
	return ItemRange{}, false
}
