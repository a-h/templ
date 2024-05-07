// This package is inspired by the GOEXPERIMENT approach of allowing feature flags for experimenting with breaking changes.
package cfg

import (
	"os"
	"strings"
)

type Flags struct {
	// RawGo will enable the support of arbibrary go code in templates.
	RawGo bool
}

var Experiment = parseTEMPLEXPERIMENT()

func parseTEMPLEXPERIMENT() *Flags {
	m := map[string]bool{}
	for _, f := range strings.Split(os.Getenv("TEMPLEXPERIMENT"), ",") {
		m[strings.ToLower(f)] = true
	}

	return &Flags{
		RawGo: m["rawgo"],
	}
}
