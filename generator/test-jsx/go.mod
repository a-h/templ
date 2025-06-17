module github.com/a-h/templ/generator/test-jsx

go 1.24.3

require github.com/a-h/templ v0.3.898

require (
	github.com/a-h/htmlformat v0.0.0-20250209131833-673be874c677 // indirect
	github.com/a-h/templ/generator/test-jsx/externjsxmod v0.0.0-00010101000000-000000000000
	github.com/google/go-cmp v0.6.0 // indirect
	golang.org/x/net v0.39.0 // indirect
)

replace github.com/a-h/templ => ../../

replace github.com/a-h/templ/generator/test-jsx/externjsxmod => ./externjsxmod/
