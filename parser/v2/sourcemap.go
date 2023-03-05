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

	for lineIndex, line := range strings.Split(src.Value, "\n") {
		srcLine := src.Range.From.Line + uint32(lineIndex)
		tgtLine := tgt.From.Line + uint32(lineIndex)

		for colIndex := 0; colIndex < len(line); colIndex++ {
			var srcCol, tgtCol uint32 = uint32(colIndex), uint32(colIndex)
			if lineIndex == 0 {
				// First line can have an offset.
				srcCol += src.Range.From.Col
				tgtCol += tgt.From.Col
			}

			if _, ok := sm.SourceLinesToTarget[srcLine]; !ok {
				sm.SourceLinesToTarget[srcLine] = make(map[uint32]Position)
			}
			sm.SourceLinesToTarget[srcLine][srcCol] = NewPositionFromValues(tgtIndex, tgtLine, tgtCol)

			if _, ok := sm.TargetLinesToSource[tgtLine]; !ok {
				sm.TargetLinesToSource[tgtLine] = make(map[uint32]Position)
			}
			sm.TargetLinesToSource[tgtLine][tgtCol] = NewPositionFromValues(srcIndex, srcLine, srcCol)

			srcIndex++
			tgtIndex++
		}
		// Increment the index for the newline char.
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
func (sm *SourceMap) SourcePositionFromTarget(line, col uint32) (src Position, ok bool) {
	lm, ok := sm.TargetLinesToSource[line]
	if !ok {
		return
	}
	src, ok = lm[col]
	return
}
