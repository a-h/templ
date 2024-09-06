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

func CodeAction(expected, actual []protocol.CodeAction) string {
	return cmp.Diff(expected, actual)
}

func CompletionList(expected, actual *protocol.CompletionList) string {
	return cmp.Diff(expected, actual,
		cmpopts.IgnoreFields(protocol.CompletionList{}, "IsIncomplete"),
	)
}

func References(expected, actual []protocol.Location) string {
	return cmp.Diff(expected, actual)
}

func CompletionListContainsText(cl *protocol.CompletionList, text string) bool {
	if cl == nil {
		return false
	}
	for _, item := range cl.Items {
		if item.Label == text {
			return true
		}
	}
	return false
}
