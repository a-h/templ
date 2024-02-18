package lspdiff

import (
	"github.com/a-h/protocol"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// This package provides a way to compare LSP	protocol messages, ignoring irrelevant fields.

func Hover(expected, actual protocol.Hover) string {
	return cmp.Diff(expected, actual,
		cmpopts.IgnoreFields(protocol.Hover{}, "Range"),
		cmpopts.IgnoreFields(protocol.MarkupContent{}, "Kind"),
	)
}
