# templ benchmark

Used to test code generation strategies for improvements to render time.

## Tasks

### run

```
go test -bench .
```

## Results

Currently getting the following results which show that using an internal `bytes.Buffer` within a template could save 25% of time.

To put this in perspective, the React benchmark is hitting 156,735 operations per second.

There are 1,000,000,000 nanoseconds in a second, so this is 6,380 ns per operation, which is 6 times slower than templ.


```
go test -bench .
goos: darwin
goarch: arm64
pkg: github.com/a-h/templ/benchmarks/templ
BenchmarkCurrent-10              1029445              1153 ns/op            1088 B/op         21 allocs/op
BenchmarkCandidate-10            1419076               845.7 ns/op          1464 B/op         20 allocs/op
BenchmarkIOWriteString-10       14667363                82.41 ns/op          352 B/op          2 allocs/op
PASS
ok      github.com/a-h/templ/benchmarks/templ   5.448s
````
