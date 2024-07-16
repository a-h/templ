package templ

// JsVar refers to a JS variable name.
// The string value of JsVar will be used as the variable argument in the function call
type JsVar string

const (
	// JsEvent represents the "event" variable in Javascript
	JsEvent JsVar = "event"
	// JsTargetElement represents the "this" variable in Javascript
	JsTargetElement JsVar = "this"
)
