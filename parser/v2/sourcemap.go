package parser

import (
	"strings"
)

// NewSourceMap creates a new lookup to map templ source code to items in the
// parsed template.
func NewSourceMap() *SourceMap {
	return &SourceMap{
		SourceLinesToTarget: make(map[uint32]map[uint32]Position),
		TargetLinesToSource: make(map[uint32]map[uint32]Position),
	}
}

type SourceMap struct {
	SourceLinesToTarget map[uint32]map[uint32]Position
	TargetLinesToSource map[uint32]map[uint32]Position
}

// Add an item to the lookup.
func (sm *SourceMap) Add(src Expression, tgt Range) (updatedFrom Position) {
	srcIndex := src.Range.From.Index
	tgtIndex := tgt.From.Index

	lines := strings.Split(src.Value, "\n")
	for lineIndex, line := range lines {
		srcLine := src.Range.From.Line + uint32(lineIndex)
		tgtLine := tgt.From.Line + uint32(lineIndex)

		var srcCol, tgtCol uint32
		if lineIndex == 0 {
			// First line can have an offset.
			srcCol += src.Range.From.Col
			tgtCol += tgt.From.Col
		}

		// Process the cols.
		for colIndex := 0; colIndex < len(line); colIndex++ {
			if _, ok := sm.SourceLinesToTarget[srcLine]; !ok {
				sm.SourceLinesToTarget[srcLine] = make(map[uint32]Position)
			}
			sm.SourceLinesToTarget[srcLine][srcCol] = NewPosition(tgtIndex, tgtLine, tgtCol)

			if _, ok := sm.TargetLinesToSource[tgtLine]; !ok {
				sm.TargetLinesToSource[tgtLine] = make(map[uint32]Position)
			}
			sm.TargetLinesToSource[tgtLine][tgtCol] = NewPosition(srcIndex, srcLine, srcCol)

			srcCol++
			tgtCol++
			srcIndex++
			tgtIndex++
		}

		// LSPs include the newline char as a col.
		if _, ok := sm.SourceLinesToTarget[srcLine]; !ok {
			sm.SourceLinesToTarget[srcLine] = make(map[uint32]Position)
		}
		sm.SourceLinesToTarget[srcLine][srcCol] = NewPosition(tgtIndex, tgtLine, tgtCol)

		if _, ok := sm.TargetLinesToSource[tgtLine]; !ok {
			sm.TargetLinesToSource[tgtLine] = make(map[uint32]Position)
		}
		sm.TargetLinesToSource[tgtLine][tgtCol] = NewPosition(srcIndex, srcLine, srcCol)

		srcIndex++
		tgtIndex++
	}
	return src.Range.From
}

// TargetPositionFromSource looks up the target position using the source position.
func (sm *SourceMap) TargetPositionFromSource(line, col uint32) (tgt Position, ok bool) {
	lm, ok := sm.SourceLinesToTarget[line]
	if !ok {
		return
	}
	tgt, ok = lm[col]
	return
}

// SourcePositionFromTarget looks the source position using the target position.
// If a source exists on the line but not the col, the function will search backwards.
func (sm *SourceMap) SourcePositionFromTarget(line, col uint32) (src Position, ok bool) {
	lm, ok := sm.TargetLinesToSource[line]
	if !ok {
		return
	}
	for {
		src, ok = lm[col]
		if ok || col == 0 {
			return
		}
		col--
	}
}
