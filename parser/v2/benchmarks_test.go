package parser

import (
	_ "embed"
	"strings"
	"testing"
)

//go:embed benchmarktestdata/benchmark.txt
var benchmarkTemplate string

func BenchmarkParse(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		if _, err := ParseString(benchmarkTemplate); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFormat(b *testing.B) {
	b.ReportAllocs()
	sb := new(strings.Builder)
	for range b.N {
		tf, err := ParseString(benchmarkTemplate)
		if err != nil {
			b.Fatal(err)
		}
		if err = tf.Write(sb); err != nil {
			b.Fatal(err)
		}
		sb.Reset()
	}
}
