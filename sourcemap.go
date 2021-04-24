package templ

// SourceExpressionTo is a record of a template Item, along with its start and end positions.
type SourceExpressionTo struct {
	Source Expression
	To     Range
}

// NewSourceMap creates a new lookup to map templ source code to items in the
// parsed template.
func NewSourceMap() *SourceMap {
	return &SourceMap{
		Items: make([]SourceExpressionTo, 0),
	}
}

type SourceMap struct {
	Items []SourceExpressionTo
}

// Add an item to the lookup.
func (sm *SourceMap) Add(src Expression, to Range) (updatedFrom Position) {
	sm.Items = append(sm.Items, SourceExpressionTo{
		Source: src,
		To:     to,
	})
	return src.Range.From
}

func (sm *SourceMap) LookupByIndex(index int64) (ir SourceExpressionTo, ok bool) {
	//TODO: Update the design so it's not looping through all the items!
	//TODO: Also find the smallest match.
	for _, cc := range sm.Items {
		if index >= cc.Source.Range.From.Index && index < cc.Source.Range.To.Index {
			return cc, true
		}
	}
	return SourceExpressionTo{}, false
}

func (sm *SourceMap) LookupByLineCol(line, col int) (ir SourceExpressionTo, ok bool) {
	for _, cc := range sm.Items {
		// Single line.
		if cc.Source.Range.From.Line == cc.Source.Range.To.Line && cc.Source.Range.To.Line == line && ((col >= cc.Source.Range.From.Col && col <= cc.Source.Range.To.Col) ||
			(col <= cc.Source.Range.From.Col && col >= cc.Source.Range.To.Col)) {
			return cc, true
		}
		// Upwards multiline.
		if cc.Source.Range.From.Line > line && (cc.Source.Range.To.Line < line || cc.Source.Range.To.Line == line && col >= cc.Source.Range.To.Col) {
			return cc, true
		}
		// Downwards multiline.
		if cc.Source.Range.From.Line < line && (cc.Source.Range.To.Line > line || cc.Source.Range.To.Line == line && col <= cc.Source.Range.To.Col) {
			return cc, true
		}
	}
	return SourceExpressionTo{}, false
}
