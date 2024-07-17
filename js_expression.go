package templ

// JsExpression represents a JavaScript expression intended for use as an argument for script blocks.
// The string value of JsExpression will be inserted directly as JavaScript code in function call arguments.
type JsExpression string

const (
	// JsEvent corresponds to the "event" variable in JavaScript, typically representing the event object
	// that is passed to event handler functions.
	JsEvent JsExpression = "event"

	// JsTargetElement corresponds to the "this" keyword in JavaScript, usually representing the element
	// that the event handler is bound to.
	JsTargetElement JsExpression = "this"
)
