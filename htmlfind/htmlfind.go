package htmlfind

import (
	"io"

	"golang.org/x/net/html"
)

// AllReader returns all nodes that match the given function.
func AllReader(r io.Reader, f Matcher) (nodes []*html.Node, err error) {
	root, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	return All(root, f), nil
}

func All(n *html.Node, f Matcher) (nodes []*html.Node) {
	if f(n) {
		nodes = append(nodes, n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		nodes = append(nodes, All(c, f)...)
	}
	return nodes
}

type Matcher func(*html.Node) bool

type Attribute struct {
	Name, Value string
}

func Attr(name, value string) Attribute {
	return Attribute{name, value}
}

func Element(name string, attrs ...Attribute) Matcher {
	return func(n *html.Node) bool {
		if n.Type != html.ElementNode {
			return false
		}
		if n.Data != name {
			return false
		}
		for _, a := range attrs {
			if getAttributeValue(n, a.Name) != a.Value {
				return false
			}
		}
		return true
	}
}

func getAttributeValue(n *html.Node, name string) string {
	for _, a := range n.Attr {
		if a.Key == name {
			return a.Val
		}
	}
	return ""
}
