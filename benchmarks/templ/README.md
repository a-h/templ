# templ benchmark

Used to test code generation strategies for improvements to render time.

## Tasks

### run

```
go test -bench .
```

## Results as of 2023-08-17

```
go test -bench .
goos: darwin
goarch: arm64
pkg: github.com/a-h/templ/benchmarks/templ
BenchmarkTempl-10                3291883               369.1 ns/op           536 B/op          6 allocs/op
BenchmarkGoTemplate-10            481052              2475 ns/op            1400 B/op         38 allocs/op
BenchmarkIOWriteString-10       20353198                56.64 ns/op          320 B/op          1 allocs/op
PASS
ok      github.com/a-h/templ/benchmarks/templ   4.650s
```

React comes in at 1,000,000,000ns / 114,131 ops/s = 8,757.5 ns per operation.
