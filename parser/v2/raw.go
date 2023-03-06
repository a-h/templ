package parser

import (
	"fmt"

	"github.com/a-h/lexical/parse"
)

var styleElement = rawElementParser{
	name: "style",
}

var scriptElement = rawElementParser{
	name: "script",
}

type rawElementParser struct {
	name string
}

func (p rawElementParser) Parse(pi parse.Input) parse.Result {
	var r RawElement

	// Check the named open tag.
	otr := parse.All(asElementOpenTag,
		parse.Rune('<'),
		parse.String(p.name),
		newAttributesParser().Parse,
		parse.Optional(parse.WithStringConcatCombiner, whitespaceParser),
		parse.Rune('>'),
	)(pi)
	if otr.Error != nil || !otr.Success {
		return otr
	}
	ot := otr.Item.(elementOpenTag)
	r.Name = ot.Name
	r.Attributes = ot.Attributes

	// Once we've got an open tag, parse anything until the end tag as the tag contents.
	// It's going to be rendered out raw.
	from := NewPositionFromInput(pi)
	end := parse.String("</" + p.name + ">")
	ectpr := parse.StringUntil(end)(pi)
	if !ectpr.Success {
		return parse.Failure(p.name, newParseError(fmt.Sprintf("<%s>: expected end tag not present", r.Name), from, NewPositionFromInput(pi)))
	}
	r.Contents = ectpr.Item.(string)
	end(pi)

	return parse.Success(p.name, r, nil)
}
