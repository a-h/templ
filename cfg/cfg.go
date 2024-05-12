// This package is inspired by the GOEXPERIMENT approach of allowing feature flags for experimenting with breaking changes.
package cfg

import (
	"os"
	"strings"
)

type Flags struct {
	// RawGo will enable the support of arbibrary Go code in templates.
	RawGo bool
}

var Experiment = parse()

func parse() *Flags {
	m := map[string]bool{}
	for _, f := range strings.Split(os.Getenv("TEMPL_EXPERIMENT"), ",") {
		m[strings.ToLower(f)] = true
	}

	return &Flags{
		RawGo: m["rawgo"],
	}
}
