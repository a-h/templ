package templ

// SourceExpressionTo is a record of an expression, along with its start and end positions.
type SourceExpressionTo struct {
	Source Expression
	Target Range
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
func (sm *SourceMap) Add(src Expression, tgt Range) (updatedFrom Position) {
	sm.Items = append(sm.Items, SourceExpressionTo{
		Source: src,
		Target: tgt,
	})
	return src.Range.From
}

// Lookup the target position by using the source position.
func (sm *SourceMap) TargetPositionFromSource(src Position) (tgt Position, mapping SourceExpressionTo, ok bool) {
	mapping, offset, ok := sm.lookupTargetBySourceLineCol(src.Line, src.Col)
	if ok {
		tgt = mapping.Target.From
		tgt.Col += offset
	}
	return
}

func (sm *SourceMap) lookupTargetBySourceLineCol(line, col int) (ir SourceExpressionTo, offset int, ok bool) {
	for _, cc := range sm.Items {
		if cc.Source.Range.From.Line == cc.Source.Range.To.Line && cc.Source.Range.To.Line == line && ((col >= cc.Source.Range.From.Col && col <= cc.Source.Range.To.Col) ||
			(col <= cc.Source.Range.From.Col && col >= cc.Source.Range.To.Col)) {
			ccOffset := col - cc.Source.Range.From.Col
			if isBestMatch := ccOffset < offset || !ok; isBestMatch {
				ok = true
				offset = ccOffset
				ir = cc
			}
		}
	}
	return
}
