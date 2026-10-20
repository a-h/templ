package testattributeevaluation

import (
	"fmt"

	"github.com/a-h/templ"
)

// recorder records each evaluation of an attribute expression. Each value that
// it creates has a unique name, so that the rendered output shows which
// evaluation each part of the output came from.
type recorder struct {
	Calls []string
}

func (r *recorder) recordCall(name string) (id string) {
	r.Calls = append(r.Calls, name)
	return fmt.Sprintf("%s_%d", name, len(r.Calls))
}

// CreateScript returns a script, and records the call.
func (r *recorder) CreateScript(name string) templ.ComponentScript {
	id := r.recordCall(name)
	return templ.ComponentScript{
		Name:     id,
		Function: "function " + id + "(){}",
		Call:     id + "()",
	}
}

// CreateScriptOrError returns a script, or createErr if it is not nil, and
// records the call.
func (r *recorder) CreateScriptOrError(name string, createErr error) (script templ.ComponentScript, err error) {
	return r.CreateScript(name), createErr
}

// CreateAttributes returns spread attributes with an onclick script, and
// records the call.
func (r *recorder) CreateAttributes(name string) templ.Attributes {
	return templ.Attributes{"onclick": r.CreateScript(name)}
}

// CreateClass returns a CSS class, and records the call.
func (r *recorder) CreateClass(name string) templ.ComponentCSSClass {
	id := r.recordCall(name)
	return templ.ComponentCSSClass{
		ID:    id,
		Class: templ.SafeCSS("." + id + "{}"),
	}
}
