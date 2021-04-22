package templ

// SourceAndTargetExpression is a record of a template Item, along with its start and end positions.
type SourceAndTargetExpression struct {
	Source, Target Expression
}

// NewSourceMap creates a new lookup to map templ source code to items in the
// parsed template.
func NewSourceMap() *SourceMap {
	return &SourceMap{
		Items: make([]SourceAndTargetExpression, 0),
	}
}

type SourceMap struct {
	Items []SourceAndTargetExpression
}

// Add an item to the lookup.
func (sm *SourceMap) Add(src, tgt Expression) (updatedFrom Position) {
	sm.Items = append(sm.Items, SourceAndTargetExpression{
		Source: src,
		Target: tgt,
	})
	return src.Range.From
}

func (sm *SourceMap) LookupByIndex(index int64) (ir SourceAndTargetExpression, ok bool) {
	//TODO: Update the design so it's not looping through all the items!
	//TODO: Also find the smallest match.
	for _, cc := range sm.Items {
		if index >= cc.Source.Range.From.Index && index < cc.Source.Range.To.Index {
			return cc, true
		}
	}
	return SourceAndTargetExpression{}, false
}

func (sm *SourceMap) LookupByLineCol(line, col int) (ir SourceAndTargetExpression, ok bool) {
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
	return SourceAndTargetExpression{}, false
}
