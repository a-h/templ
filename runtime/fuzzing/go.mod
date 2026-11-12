module github.com/a-h/templ/runtime/fuzzing

go 1.26.0

require (
	github.com/a-h/templ v0.3.833
	golang.org/x/net v0.62.0
)

require (
	github.com/google/go-cmp v0.7.0
	rogchap.com/v8go v0.9.0
)

require github.com/tdewolff/parse/v2 v2.8.20

replace github.com/a-h/templ => ../..
