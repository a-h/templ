package parser

import (
	"strings"
	"unicode/utf8"
)

// NewSourceMap creates a new lookup to map templ source code to items in the
// parsed template.
func NewSourceMap() *SourceMap {
	return &SourceMap{
		SourceLinesToTarget:       make(map[uint32]map[uint32]Position),
		TargetLinesToSource:       make(map[uint32]map[uint32]Position),
		SourceSymbolRangeToTarget: make(map[uint32]map[uint32]Range),
		TargetSymbolRangeToSource: make(map[uint32]map[uint32]Range),
	}
}

type SourceMap struct {
	Expressions               []string
	SourceLinesToTarget       map[uint32]map[uint32]Position
	TargetLinesToSource       map[uint32]map[uint32]Position
	SourceSymbolRangeToTarget map[uint32]map[uint32]Range
	TargetSymbolRangeToSource map[uint32]map[uint32]Range
}

func (sm *SourceMap) AddSymbolRange(src Range, tgt Range) {
	sm.SourceSymbolRangeToTarget[src.From.Line] = make(map[uint32]Range)
	sm.SourceSymbolRangeToTarget[src.From.Line][src.From.Col] = tgt
	sm.TargetSymbolRangeToSource[tgt.From.Line] = make(map[uint32]Range)
	sm.TargetSymbolRangeToSource[tgt.From.Line][tgt.From.Col] = src
}

func (sm *SourceMap) SymbolTargetRangeFromSource(line, col uint32) (tgt Range, ok bool) {
	lm, ok := sm.SourceSymbolRangeToTarget[line]
	if !ok {
		return
	}
	tgt, ok = lm[col]
	return
}

func (sm *SourceMap) SymbolSourceRangeFromTarget(line, col uint32) (src Range, ok bool) {
	lm, ok := sm.TargetSymbolRangeToSource[line]
	if !ok {
		return
	}
	src, ok = lm[col]
	return
}

// Add an item to the lookup.
func (sm *SourceMap) Add(src Expression, tgt Range) (updatedFrom Position) {
	sm.Expressions = append(sm.Expressions, src.Value)
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
		for _, r := range line {
			if _, ok := sm.SourceLinesToTarget[srcLine]; !ok {
				sm.SourceLinesToTarget[srcLine] = make(map[uint32]Position)
			}
			sm.SourceLinesToTarget[srcLine][srcCol] = NewPosition(tgtIndex, tgtLine, tgtCol)

			if _, ok := sm.TargetLinesToSource[tgtLine]; !ok {
				sm.TargetLinesToSource[tgtLine] = make(map[uint32]Position)
			}
			sm.TargetLinesToSource[tgtLine][tgtCol] = NewPosition(srcIndex, srcLine, srcCol)

			// Ignore invalid runes.
			rlen := utf8.RuneLen(r)
			if rlen < 0 {
				rlen = 1
			}
			srcCol += uint32(rlen)
			tgtCol += uint32(rlen)
			srcIndex += int64(rlen)
			tgtIndex += int64(rlen)
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
