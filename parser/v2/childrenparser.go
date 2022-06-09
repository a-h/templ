package parser

import (
	"github.com/a-h/lexical/parse"
)

var childrenExpression = parse.All(func(i []interface{}) (result interface{}, ok bool) {
	return ChildrenExpression{}, true
},
	openBraceWithOptionalPadding,
	optionalWhitespaceParser,
	parse.String("children..."),
	optionalWhitespaceParser,
	closeBraceWithOptionalPadding,
)
