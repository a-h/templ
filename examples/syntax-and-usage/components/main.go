package main

import (
	"context"
	"fmt"
	"html"
	"io"
	"log"
	"os"

	"github.com/a-h/templ"
)

func main() {
	ctx := context.Background()
	if err := list([]string{"a", "b", "c"}).Render(ctx, os.Stdout); err != nil {
		log.Fatalf("failed to render list: %v", err)
	}
	if err := codeList([]string{"A", "B", "C"}).Render(ctx, os.Stdout); err != nil {
		log.Fatalf("failed to render code list: %v", err)
	}
}

func codeList(items []string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		if _, err = io.WriteString(w, "<ol>"); err != nil {
			return
		}
		for _, item := range items {
			if _, err = io.WriteString(w, fmt.Sprintf("<li>%s</li>", html.EscapeString(item))); err != nil {
				return
			}
		}
		if _, err = io.WriteString(w, "</ol>"); err != nil {
			return
		}
		return nil
	})
}
