package parser

import (
	"github.com/a-h/parse"
)

var doctypeStartParser = parse.StringInsensitive("<!doctype ")

var untilLtOrGt = parse.Or(lt, gt)
var stringUntilLtOrGt = parse.StringUntil(untilLtOrGt)

var docType = parse.Func(func(pi *parse.Input) (n Node, ok bool, err error) {
	var r DocType
	if _, ok, err = doctypeStartParser.Parse(pi); err != nil || !ok {
		return
	}

	// Once a doctype has started, take everything until the end.
	if r.Value, ok, err = stringUntilLtOrGt.Parse(pi); err != nil || !ok {
		err = parse.Error("unclosed DOCTYPE", pi.Position())
		return
	}

	// Clear the final '>'.
	if _, ok, err = gt.Parse(pi); err != nil || !ok {
		err = parse.Error("unclosed DOCTYPE", pi.Position())
		return
	}

	return r, true, nil
})
