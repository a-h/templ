package parser

const (
	errScriptStartTagInEscapedScriptData = "<script>: <script is not allowed after an unclosed <!--, because the browser does not end the script element at the next </script>"
	errScriptEndTagNotRecognized         = "<script>: the browser ends the script element at </script followed by whitespace, / or >, in any case, so end it with </script>, and write </script within a string as <\\/script"
	errGoExpressionCompletesTag          = "<script>: a Go expression can't directly follow <, <!, <!-, or part of a <script or </script tag, because its value could complete the tag, and change where the browser ends the script element"
)

// scriptDataState is a state of the HTML tokenizer of a browser within the contents
// of a script element. Each state is the state of the same name in
// https://html.spec.whatwg.org/multipage/parsing.html#script-data-state, except that
// the end tag name states are merged into the end tag open states.
type scriptDataState int

const (
	scriptData scriptDataState = iota
	scriptDataLessThanSign
	scriptDataEndTagOpen
	scriptDataEscapeStart
	scriptDataEscapeStartDash
	scriptDataEscaped
	scriptDataEscapedDash
	scriptDataEscapedDashDash
	scriptDataEscapedLessThanSign
	scriptDataEscapedEndTagOpen
	scriptDataDoubleEscapeStart
)

// scriptDataResult is the result of a byte of script content.
type scriptDataResult int

const (
	scriptDataContinues scriptDataResult = iota
	// scriptDataEnds means that the browser ends the script element.
	scriptDataEnds
	// scriptDataDoubleEscaped means that the browser does not end the script
	// element at the next </script> end tag.
	scriptDataDoubleEscaped
)

// GetErrorMessage returns the parse error for a result that is not
// scriptDataContinues.
func (r scriptDataResult) GetErrorMessage() string {
	if r == scriptDataEnds {
		return errScriptEndTagNotRecognized
	}
	return errScriptStartTagInEscapedScriptData
}

// scriptDataTracker tracks the state of the HTML tokenizer of a browser within the
// contents of a script element, so that the parser can return an error where the
// browser would end the element somewhere other than the </script> end tag that
// templ ends it at.
type scriptDataTracker struct {
	state scriptDataState
	// tagName is the temporary buffer of the spec, which holds the lowercased name
	// of the tag being read.
	tagName []byte
	// tagIndex is the index in the source of the '<' of the tag being read.
	tagIndex int
}

// RecordText records text that starts at index in the source. It returns the index
// of the '<' of a tag that ends the script element, or that starts double escaped
// script data, and the result.
func (t *scriptDataTracker) RecordText(index int, text string) (tagIndex int, result scriptDataResult) {
	for i := range len(text) {
		result = t.consume(index+i, text[i])
		if result != scriptDataContinues {
			return t.tagIndex, result
		}
	}
	return 0, scriptDataContinues
}

// RecordValue records the value of a Go expression. Each escaper encodes '<' and
// '>', so a value can't start a tag, or end an HTML comment, but it can contain the
// other characters of <!--, <script, and </script. It returns false if the value
// directly follows the start of one, because the value could complete it. In
// escaped script data, the value is assumed not to end it.
func (t *scriptDataTracker) RecordValue() (ok bool) {
	switch t.state {
	case scriptData:
		return true
	case scriptDataEscaped, scriptDataEscapedDash, scriptDataEscapedDashDash:
		t.state = scriptDataEscaped
		return true
	}
	return false
}

func (t *scriptDataTracker) consume(index int, c byte) (result scriptDataResult) {
	switch t.state {
	case scriptData:
		if c == '<' {
			t.state = scriptDataLessThanSign
			t.tagIndex = index
		}
	case scriptDataLessThanSign:
		switch c {
		case '/':
			t.state = scriptDataEndTagOpen
			t.tagName = t.tagName[:0]
		case '!':
			t.state = scriptDataEscapeStart
		default:
			t.state = scriptData
			return t.consume(index, c)
		}
	case scriptDataEndTagOpen, scriptDataEscapedEndTagOpen:
		if isASCIIAlpha(c) {
			t.tagName = append(t.tagName, toLowerASCIIByte(c))
			return scriptDataContinues
		}
		if isTagNameTerminator(c) && string(t.tagName) == "script" {
			return scriptDataEnds
		}
		t.state = t.leaveEndTagOpen()
		return t.consume(index, c)
	case scriptDataEscapeStart:
		if c != '-' {
			t.state = scriptData
			return t.consume(index, c)
		}
		t.state = scriptDataEscapeStartDash
	case scriptDataEscapeStartDash:
		if c != '-' {
			t.state = scriptData
			return t.consume(index, c)
		}
		t.state = scriptDataEscapedDashDash
	case scriptDataEscaped:
		switch c {
		case '-':
			t.state = scriptDataEscapedDash
		case '<':
			t.state = scriptDataEscapedLessThanSign
			t.tagIndex = index
		}
	case scriptDataEscapedDash:
		switch c {
		case '-':
			t.state = scriptDataEscapedDashDash
		case '<':
			t.state = scriptDataEscapedLessThanSign
			t.tagIndex = index
		default:
			t.state = scriptDataEscaped
		}
	case scriptDataEscapedDashDash:
		switch c {
		case '-':
		case '<':
			t.state = scriptDataEscapedLessThanSign
			t.tagIndex = index
		case '>':
			t.state = scriptData
		default:
			t.state = scriptDataEscaped
		}
	case scriptDataEscapedLessThanSign:
		switch {
		case c == '/':
			t.state = scriptDataEscapedEndTagOpen
			t.tagName = t.tagName[:0]
		case isASCIIAlpha(c):
			t.state = scriptDataDoubleEscapeStart
			t.tagName = append(t.tagName[:0], toLowerASCIIByte(c))
		default:
			t.state = scriptDataEscaped
			return t.consume(index, c)
		}
	case scriptDataDoubleEscapeStart:
		switch {
		case isTagNameTerminator(c):
			if string(t.tagName) == "script" {
				return scriptDataDoubleEscaped
			}
			t.state = scriptDataEscaped
		case isASCIIAlpha(c):
			t.tagName = append(t.tagName, toLowerASCIIByte(c))
		default:
			t.state = scriptDataEscaped
			return t.consume(index, c)
		}
	}
	return scriptDataContinues
}

// leaveEndTagOpen returns the state that an end tag that does not end the element
// returns to.
func (t *scriptDataTracker) leaveEndTagOpen() scriptDataState {
	if t.state == scriptDataEscapedEndTagOpen {
		return scriptDataEscaped
	}
	return scriptData
}

func isTagNameTerminator(c byte) bool {
	switch c {
	case '\t', '\n', '\f', '\r', ' ', '/', '>':
		return true
	}
	return false
}

func isASCIIAlpha(c byte) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

func toLowerASCIIByte(c byte) byte {
	if 'A' <= c && c <= 'Z' {
		return c + 'a' - 'A'
	}
	return c
}
