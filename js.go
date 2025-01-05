package templ

import (
	"crypto/sha256"
	"encoding/hex"
	"html"
)

// JSUnsafeFuncCall calls arbitrary JavaScript in the js parameter.
//
// Use of this function presents a security risk - the JavaScript must come
// from a trusted source, because it will be included as-is in the output.
func JSUnsafeFuncCall[T ~string](js T) ComponentScript {
	sum := sha256.Sum256([]byte(js))
	return ComponentScript{
		Name: "jsUnsafeFuncCall_" + hex.EncodeToString(sum[:]),
		// Function is empty because the body of the function is defined elsewhere,
		// e.g. in a <script> tag within a templ.Once block.
		Function:   "",
		Call:       html.EscapeString(string(js)),
		CallInline: string(js),
	}
}

// JSFuncCall calls a JavaScript function with the given arguments.
//
// It can be used in event handlers, e.g. onclick, onhover, etc. or
// directly in HTML.
func JSFuncCall[T ~string](functionName T, args ...any) ComponentScript {
	call := SafeScript(string(functionName), args...)
	sum := sha256.Sum256([]byte(call))
	return ComponentScript{
		Name: "jsFuncCall_" + hex.EncodeToString(sum[:]),
		// Function is empty because the body of the function is defined elsewhere,
		// e.g. in a <script> tag within a templ.Once block.
		Function:   "",
		Call:       call,
		CallInline: SafeScriptInline(string(functionName), args...),
	}
}
