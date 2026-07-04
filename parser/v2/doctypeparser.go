package parser

import (
	"github.com/a-h/parse"
)

var doctypeStartParser = parse.StringInsensitive("<!doctype ")

var untilLtOrGt = parse.Or(lt, gt)
var stringUntilLtOrGt = parse.StringUntil(untilLtOrGt)

var docType = parse.Func(func(pi *parse.Input) (n Node, ok bool, err error) {
	start := pi.Position()
	if _, ok, err = doctypeStartParser.Parse(pi); err != nil || !ok {
		return
	}
	// The parser consumes the required separator space, but OpenRange covers only the opener.
	openEnd := pi.PositionAt(pi.Index() - 1)

	r := &DocType{
		OpenRange: NewRange(start, openEnd),
	}

	valueStart := pi.Position()
	if r.Value, ok, err = stringUntilLtOrGt.Parse(pi); err != nil || !ok {
		err = parse.Error("unclosed DOCTYPE", start)
		return
	}
	valueEnd := pi.Position()
	r.ValueRange = NewRange(valueStart, valueEnd)

	closeStart := pi.Position()
	if _, ok, err = gt.Parse(pi); err != nil || !ok {
		err = parse.Error("unclosed DOCTYPE", start)
		return
	}
	closeEnd := pi.Position()
	r.CloseRange = NewRange(closeStart, closeEnd)

	r.Range = NewRange(start, closeEnd)

	return r, true, nil
})
