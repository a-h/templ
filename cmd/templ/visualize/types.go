package visualize

import (
	"context"
	"fmt"
	"html"
	"io"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/a-h/templ/parser/v2"
)

func HTML(templFileName string, templContents, goContents string, sourceMap *parser.SourceMap) templ.Component {
	tl := templLines{contents: string(templContents), sourceMap: sourceMap}
	gl := goLines{contents: string(goContents), sourceMap: sourceMap}
	return combine(templFileName, tl, gl)
}

type templLines struct {
	contents  string
	sourceMap *parser.SourceMap
}

func (tl templLines) Render(ctx context.Context, w io.Writer) error {
	templLines := strings.Split(tl.contents, "\n")
	for lineIndex, line := range templLines {
		w.Write([]byte("<span>" + strconv.Itoa(lineIndex) + "&nbsp;</span>\n"))
		for colIndex, c := range line {
			if _, m, ok := tl.sourceMap.TargetPositionFromSource(uint32(lineIndex), uint32(colIndex)); ok {
				sourceID := fmt.Sprintf("src_%d_%d_%d", m.Source.Range.From.Index, m.Source.Range.From.Line, m.Source.Range.From.Col)
				targetID := fmt.Sprintf("tgt_%d_%d_%d", m.Target.From.Index, m.Target.From.Index, m.Target.From.Col)
				if err := mappedCharacter(string(c), sourceID, targetID).Render(ctx, w); err != nil {
					return err
				}
			} else {
				s := html.EscapeString(string(c))
				s = strings.ReplaceAll(s, "\t", "&nbsp;")
				s = strings.ReplaceAll(s, " ", "&nbsp;")
				if _, err := w.Write([]byte(s)); err != nil {
					return err
				}
			}
		}
		w.Write([]byte("\n<br/>\n"))
	}
	return nil
}

type goLines struct {
	contents  string
	sourceMap *parser.SourceMap
}

func (gl goLines) Render(ctx context.Context, w io.Writer) error {
	templLines := strings.Split(gl.contents, "\n")
	for lineIndex, line := range templLines {
		w.Write([]byte("<span>" + strconv.Itoa(lineIndex) + "&nbsp;</span>\n"))
		for colIndex, c := range line {
			if _, m, ok := gl.sourceMap.SourcePositionFromTarget(uint32(lineIndex), uint32(colIndex)); ok {
				sourceID := fmt.Sprintf("src_%d_%d_%d", m.Source.Range.From.Index, m.Source.Range.From.Line, m.Source.Range.From.Col)
				targetID := fmt.Sprintf("tgt_%d_%d_%d", m.Target.From.Index, m.Target.From.Index, m.Target.From.Col)
				if err := mappedCharacter(string(c), sourceID, targetID).Render(ctx, w); err != nil {
					return err
				}
			} else {
				s := html.EscapeString(string(c))
				s = strings.ReplaceAll(s, "\t", "&nbsp;")
				s = strings.ReplaceAll(s, " ", "&nbsp;")
				if _, err := w.Write([]byte(s)); err != nil {
					return err
				}
			}
		}
		w.Write([]byte("\n<br/>\n"))
	}
	return nil
}
