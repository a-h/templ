// This package is inspired by the GOEXPERIMENT approach of allowing feature flags for experimenting with breaking changes.
package cfg

import (
	"os"
	"strings"
)

type Flags struct{}

var Experiment = parse()

func parse() *Flags {
	m := map[string]bool{}
	for f := range strings.SplitSeq(os.Getenv("TEMPL_EXPERIMENT"), ",") {
		m[strings.ToLower(f)] = true
	}

	return &Flags{}
}
