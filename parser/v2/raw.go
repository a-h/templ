package parser

import (
	"fmt"

	"github.com/a-h/parse"
)

var styleElement = rawElementParser{
	name: "style",
}

type rawElementParser struct {
	name string
}

func (p rawElementParser) Parse(pi *parse.Input) (n Node, ok bool, err error) {
	start := pi.Index()

	// <
	if _, ok, err = lt.Parse(pi); err != nil || !ok {
		return
	}

	// Element name.
	var e RawElement
	if e.Name, ok, err = elementNameParser.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	if e.Name != p.name {
		pi.Seek(start)
		ok = false
		return
	}

	if e.Attributes, ok, err = (attributesParser{}).Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	// Optional whitespace.
	if _, _, err = parse.OptionalWhitespace.Parse(pi); err != nil {
		pi.Seek(start)
		return
	}

	// >
	if _, ok, err = gt.Parse(pi); err != nil || !ok {
		pi.Seek(start)
		return
	}

	// Once we've got an open tag, parse anything until the end tag as the tag contents.
	// It's going to be rendered out raw.
	end := parse.All(parse.String("</"), parse.String(p.name), parse.String(">"))
	if e.Contents, ok, err = parse.StringUntil(end).Parse(pi); err != nil || !ok {
		err = parse.Error(fmt.Sprintf("<%s>: expected end tag not present", e.Name), pi.Position())
		return
	}
	// Cut the end element.
	_, _, _ = end.Parse(pi)

	return e, true, nil
}
