package parser

import (
	"github.com/a-h/parse"
)

var doctypeStartParser = parse.StringInsensitive("<!doctype ")

var docType = parse.Func(func(pi *parse.Input) (r DocType, ok bool, err error) {
	if _, ok, err = doctypeStartParser.Parse(pi); err != nil || !ok {
		return
	}

	// Once a doctype has started, take everything until the end.
	if r.Value, ok, err = Must(parse.StringUntil(parse.Or(lt, gt)), "unclosed DOCTYPE").Parse(pi); err != nil || !ok {
		return
	}

	// Clear the final '>'.
	if _, ok, err = Must(gt, "unclosed DOCTYPE").Parse(pi); err != nil || !ok {
		return
	}

	return r, true, nil
})
