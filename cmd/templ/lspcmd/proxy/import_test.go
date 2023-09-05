package proxy

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFindLastImport(t *testing.T) {
	tests := []struct {
		name          string
		templContents string
		packageName   string
		expected      string
	}{
		{
			name: "if there are no imports, add a single line import",
			templContents: `package main

templ example() {
}
`,
			packageName: "strings",
			expected: `package main

import "strings"

templ example() {
}
`,
		},
		{
			name: "if there is an existing single-line imports, add one at the end",
			templContents: `package main

import "strings"

templ example() {
}
`,
			packageName: "fmt",
			expected: `package main

import "strings"
import "fmt"

templ example() {
}
`,
		},
		{
			name: "if there are multiple existing single-line imports, add one at the end",
			templContents: `package main

import "strings"
import "fmt"

templ example() {
}
`,
			packageName: "time",
			expected: `package main

import "strings"
import "fmt"
import "time"

templ example() {
}
`,
		},
		{
			name: "if there are existing multi-line imports, add one at the end",
			templContents: `package main

import (
	"strings"
)

templ example() {
}
`,
			packageName: "fmt",
			expected: `package main

import (
	"strings"
	"fmt"
)

templ example() {
}
`,
		},
		{
			name: "ignore imports that happen after templates",
			templContents: `package main

import "strings"

templ example() {
}

import "other"
`,
			packageName: "fmt",
			expected: `package main

import "strings"
import "fmt"

templ example() {
}

import "other"
`,
		},
		{
			name: "ignore imports that happen after funcs in the file",
			templContents: `package main

import "strings"

func example() {
}

import "other"
`,
			packageName: "fmt",
			expected: `package main

import "strings"
import "fmt"

func example() {
}

import "other"
`,
		},
		{
			name: "ignore imports that happen after css expressions in the file",
			templContents: `package main

import "strings"

css example() {
}

import "other"
`,
			packageName: "fmt",
			expected: `package main

import "strings"
import "fmt"

css example() {
}

import "other"
`,
		},
		{
			name: "ignore imports that happen after script expressions in the file",
			templContents: `package main

import "strings"

script example() {
}

import "other"
`,
			packageName: "fmt",
			expected: `package main

import "strings"
import "fmt"

script example() {
}

import "other"
`,
		},
		{
			name: "ignore imports that happen after var expressions in the file",
			templContents: `package main

import "strings"

var s string

import "other"
`,
			packageName: "fmt",
			expected: `package main

import "strings"
import "fmt"

var s string

import "other"
`,
		},
		{
			name: "ignore imports that happen after const expressions in the file",
			templContents: `package main

import "strings"

const s = "test"

import "other"
`,
			packageName: "fmt",
			expected: `package main

import "strings"
import "fmt"

const s = "test"

import "other"
`,
		},
		{
			name: "ignore imports that happen after type expressions in the file",
			templContents: `package main

import "strings"

type Value int

import "other"
`,
			packageName: "fmt",
			expected: `package main

import "strings"
import "fmt"

type Value int

import "other"
`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lines := strings.Split(test.templContents, "\n")
			imp := addImport(lines, fmt.Sprintf("%q", test.packageName))
			textWithoutNewline := strings.TrimSuffix(imp.Text, "\n")
			actualLines := append(lines[:imp.LineIndex], append([]string{textWithoutNewline}, lines[imp.LineIndex:]...)...)
			actual := strings.Join(actualLines, "\n")
			if diff := cmp.Diff(test.expected, actual); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestGetPackageFromItemDetail(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    `"fmt"`,
			expected: `"fmt"`,
		},
		{
			input:    `func(state fmt.State, verb rune) string (from "fmt")`,
			expected: `"fmt"`,
		},
		{
			input:    `non matching`,
			expected: `non matching`,
		},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			actual := getPackageFromItemDetail(test.input)
			if test.expected != actual {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}
