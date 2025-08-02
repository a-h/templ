package htmlfind_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/a-h/templ/internal/htmlfind"
	"golang.org/x/net/html"
)

func TestFind(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		find   htmlfind.Matcher
		assert func(t *testing.T, nodes []*html.Node)
	}{
		{
			name:  "find a paragraph",
			input: "<div><p>hello</p></div>",
			find:  htmlfind.Element("p"),
			assert: func(t *testing.T, nodes []*html.Node) {
				if len(nodes) != 1 {
					t.Fatalf("expected 1 node, got %d", len(nodes))
				}
				n := nodes[0]
				if n.Data != "p" {
					t.Errorf("expected p, got %s", n.Data)
				}
				if n.FirstChild.Data != "hello" {
					t.Errorf("expected hello, got %s", n.FirstChild.Data)
				}
			},
		},
		{
			name: "find a div with a specific attribute",
			input: `<div class="a">
			  <div class="b">
				  <div class="c"></div>
					</div>
				</div>`,
			find: htmlfind.Element("div", htmlfind.Attr("class", "b")),
			assert: func(t *testing.T, nodes []*html.Node) {
				if len(nodes) != 1 {
					t.Fatalf("expected 1 node, got %d", len(nodes))
				}
				n := nodes[0]
				if n.Data != "div" {
					t.Errorf("expected div, got %s", n.Data)
				}
				if n.Attr[0].Val != "b" {
					t.Errorf("expected b, got %s", n.Attr[0].Val)
				}
			},
		},
		{
			name: "find multiple divs with a specific attribute",
			input: `<div class="a">
			  <div class="b">
				  Content A
				</div>
				<div noclass></div>
				<div class="b">
				  Content B
				</div>
				<div class="c">
				</div>
			</div>`,
			find: htmlfind.Element("div", htmlfind.Attr("class", "b")),
			assert: func(t *testing.T, nodes []*html.Node) {
				if len(nodes) != 2 {
					t.Fatalf("expected 2 nodes, got %d", len(nodes))
				}
				for _, n := range nodes {
					if n.Data != "div" {
						t.Errorf("expected div, got %s", n.Data)
					}
					if n.Attr[0].Val != "b" {
						t.Errorf("expected b, got %s", n.Attr[0].Val)
					}
					if strings.TrimSpace(n.FirstChild.Data) != "Content A" && strings.TrimSpace(n.FirstChild.Data) != "Content B" {
						t.Errorf("expected Content A or Content B, got %s", n.FirstChild.Data)
					}
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := strings.NewReader(tt.input)
			results, err := htmlfind.AllReader(r, tt.find)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.assert == nil {
				t.Fatalf("no assertion provided")
			}
			tt.assert(t, results)
		})
	}

	t.Run("invalid HTML returns an error", func(t *testing.T) {
		var r errorReader
		finder := func(n *html.Node) bool {
			return n.Data == "p"
		}
		_, err := htmlfind.AllReader(r, finder)
		if err != errFailedToRead {
			t.Fatalf("expected an error, got %v", err)
		}
	})
}

var errFailedToRead = errors.New("failed to read")

type errorReader struct{}

func (errorReader) Read(p []byte) (n int, err error) {
	return 0, errFailedToRead
}
