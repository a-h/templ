package runtimebench

import (
	"bufio"
	"context"
	"io"
	"strconv"
	"testing"

	"github.com/a-h/templ"
)

func newPage(rows int) Page {
	p := Page{Title: "Report & <summary>", Rows: make([]Row, rows)}
	tags := []string{"go", "html & css", "<templ>"}
	for i := range p.Rows {
		id := strconv.Itoa(i)
		cells := make([]Cell, 8)
		for j := range cells {
			v := i*j - 50
			cells[j] = Cell{Text: strconv.Itoa(v), Neg: v < 0}
		}
		name := "User " + id
		if i%2 == 0 {
			name = `User "` + id + `" <admin> & co`
		}
		note := ""
		if i%4 == 0 {
			note = "Note for <row> " + id + " & more"
		}
		p.Rows[i] = Row{
			ID:      "row-" + id,
			Name:    name,
			URL:     "/rows/" + id + "?tab=details&sort=asc",
			Mail:    "mailto:user" + id + "@example.com",
			Class:   "cell",
			Count:   i,
			Active:  i%2 == 0,
			Done:    i%3 == 0,
			Locked:  i%5 == 0,
			Odd:     i%2 == 1,
			Note:    note,
			Tags:    tags[:i%4],
			Cells:   cells,
			Attrs:   templ.Attributes{"data-role": "row", "aria-selected": "false"},
			Ordered: templ.OrderedAttributes{{Key: "data-id", Value: id}, {Key: "title", Value: name}},
		}
	}
	return p
}

func BenchmarkRender(b *testing.B) {
	p := newPage(100)
	ctx := context.Background()
	w := bufio.NewWriter(io.Discard)
	b.ReportAllocs()
	for b.Loop() {
		if err := Render(p).Render(ctx, w); err != nil {
			b.Fatal(err)
		}
		if err := w.Flush(); err != nil {
			b.Fatal(err)
		}
	}
}
