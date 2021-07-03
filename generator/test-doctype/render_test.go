package testdoctype

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const expected = `<!doctype html>` +
	`<html lang="en">` +
	`<head>` +
	`<meta charset="UTF-8">` +
	`<meta http-equiv="X-UA-Compatible" content="IE=edge">` +
	`<meta name="viewport" content="width=device-width, initial-scale=1.0">` +
	`<title>title</title>` +
	`</head>` +
	`<body>content</body>` +
	`</html>`

func TestHTML(t *testing.T) {
	w := new(strings.Builder)
	err := Layout("title", "content").Render(context.Background(), w)
	if err != nil {
		t.Errorf("failed to render: %v", err)
	}
	if diff := cmp.Diff(expected, w.String()); diff != "" {
		t.Error(diff)
	}
}
