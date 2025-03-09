module github.com/a-h/templ/runtime/fuzzing

go 1.23.3

require (
	github.com/a-h/templ v0.3.833
	golang.org/x/net v0.34.0
)

require rogchap.com/v8go v0.9.0

require github.com/brianvoe/gofakeit/v7 v7.2.1 // indirect

replace github.com/a-h/templ => ../..
