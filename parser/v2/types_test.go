package parser

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFormatting(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "void elements are converted to self-closing elements",
			input: ` // first line removed to make indentation clear in Go code
package test

templ input(value, validation string) {
	<area></area>
	<base></base>
	<br></br>
	<col></col>
	<command></command>
	<embed></embed>
	<hr></hr>
	<img></img>
	<input></input>
	<keygen></keygen>
	<link></link>
	<meta></meta>
	<param></param>
	<source></source>
	<track></track>
	<wbr></wbr>
}

`,
			expected: `// first line removed to make indentation clear in Go code
package test

templ input(value, validation string) {
	<area/>
	<base/>
	<br/>
	<col/>
	<command/>
	<embed/>
	<hr/>
	<img/>
	<input/>
	<keygen/>
	<link/>
	<meta/>
	<param/>
	<source/>
	<track/>
	<wbr/>
}

`,
		},
		{
			name: "script tags are not converted to self-closing elements",
			input: ` // first line removed to make indentation clear in Go code
package test

templ input(value, validation string) {
	<script src="https://example.com/myscript.js"></script>
}

`,
			expected: `// first line removed to make indentation clear in Go code
package test

templ input(value, validation string) {
	<script src="https://example.com/myscript.js"></script>
}

`,
		},
		{
			name: "empty elements stay on the same line",
			input: ` // first line removed to make indentation clear in Go code
package test

templ input(value, validation string) {
	<div>
		<p>
		</p>
	</div>
}

`,
			expected: `// first line removed to make indentation clear in Go code
package test

templ input(value, validation string) {
	<div>
		<p></p>
	</div>
}

`,
		},
		{
			name: "if the element only contains inline elements, they end up on the same line",
			input: ` // first line removed to make indentation clear in Go code
package test

templ input(value, validation string) {
	<div><p>{ "the" }<a href="http://example.com">{ "data" }</a></p></div>
}
`,
			expected: `// first line removed to make indentation clear in Go code
package test

templ input(value, validation string) {
	<div>
		<p>{ "the" }<a href="http://example.com">{ "data" }</a></p>
	</div>
}

`,
		},
		{
			name: "if an element contains any block elements, all of the child elements are split onto new lines",
			input: ` // first line removed to make indentation clear in Go code
package test

templ nested() {
	<div>{ "the" }<div>{ "other" }</div></div>
}

`,
			expected: `// first line removed to make indentation clear in Go code
package test

templ nested() {
	<div>
		{ "the" }
		<div>{ "other" }</div>
	</div>
}

`,
		},
		{
			name: "for loops are placed on a new line",
			input: ` // first line removed to make indentation clear in Go code
package test

templ input(items []string) {
<div>{ "the" }<div>{ "other" }</div>for _, item := range items {
<div>{ item }</div>
}</div>
}
`,
			expected: `// first line removed to make indentation clear in Go code
package test

templ input(items []string) {
	<div>
		{ "the" }
		<div>{ "other" }</div>
		for _, item := range items {
			<div>{ item }</div>
		}
	</div>
}

`,
		},
		{
			name: "if statements are placed on a new line",
			input: ` // first line removed to make indentation clear in Go code
package test

templ input(items []string) {
<div>{ "the" }<div>{ "other" }</div>if items != nil {
<div>{ items[0] }</div>
		} else {
			<div>{ items[1] }</div>
		}
</div>
}
`,
			expected: `// first line removed to make indentation clear in Go code
package test

templ input(items []string) {
	<div>
		{ "the" }
		<div>{ "other" }</div>
		if items != nil {
			<div>{ items[0] }</div>
		} else {
			<div>{ items[1] }</div>
		}
	</div>
}

`,
		},
		{
			name: "switch statements are placed on a new line",
			input: ` // first line removed to make indentation clear in Go code
package test

templ input(items []string) {
<div>{ "the" }<div>{ "other" }</div>switch items[0] {
	case "a":
<div>{ items[0] }</div>
	case "b":
<div>{ items[1] }</div>
}</div>
}
`,
			expected: `// first line removed to make indentation clear in Go code
package test

templ input(items []string) {
	<div>
		{ "the" }
		<div>{ "other" }</div>
		switch items[0] {
			case "a":
				<div>{ items[0] }</div>
			case "b":
				<div>{ items[1] }</div>
		}
	</div>
}

`,
		},
		{
			name: "css is indented by one level",
			input: ` // first line removed to make indentation clear in Go code
package test

css ClassName() {
background-color: #ffffff;
color: { constants.White };
}
`,
			expected: `// first line removed to make indentation clear in Go code
package test

css ClassName() {
	background-color: #ffffff;
	color: { constants.White };
}

`,
		},
		{
			name: "css whitespace is tidied",
			input: ` // first line removed to make indentation clear in Go code
package test

css ClassName() {
background-color    :   #ffffff	;
  color	:  { constants.White };
  }
		`,
			expected: `// first line removed to make indentation clear in Go code
package test

css ClassName() {
	background-color: #ffffff;
	color: { constants.White };
}

`,
		},
		{
			name: "tables are formatted well",
			input: ` // first line removed to make indentation clear
package test

templ table(accountNumber string, registration string) {
	<table>
	       <tr>
		       <th width="20%">Your account number</th>
		       <td width="80%">{ accountNumber }</td>
	       </tr>
	       <tr>
		       <td>Registration</td>
		       <td>{ strings.ToUpper(registration) }</td>
	       </tr>
	</table>
}
`,
			expected: ` // first line removed to make indentation clear
package test

templ table(accountNumber string, registration string) {
	<table>
		<tr>
			<th width="20%">Your account number</th>
			<td width="80%">{ accountNumber }</td>
		</tr>
		<tr>
			<td>Registration</td>
			<td>{ strings.ToUpper(registration) }</td>
		</tr>
	</table>
}

`,
		},
		{
			name: "conditional expressions are placed on their own lines",
			input: ` // first line removed to make indentation clear
package test

templ conditionalAttributes(addClass bool) {
	<div id="conditional" if addClass {
		class="itWasTrue"
	}
	width="300">Content</div>
}
`,
			expected: ` // first line removed to make indentation clear
package test

templ conditionalAttributes(addClass bool) {
	<div id="conditional"
		if addClass {
			class="itWasTrue"
		}
		width="300">Content</div>
}

`,
		},
		{
			name: "conditional expressions are indented on their own lines",
			input: ` // first line removed to make indentation clear
package test

templ conditionalAttributes(addClass bool) {
	<div id="conditional"
if addClass {
class="itWasTrue"
}
width="300">Content</div>
}
`,
			expected: ` // first line removed to make indentation clear
package test

templ conditionalAttributes(addClass bool) {
	<div id="conditional"
		if addClass {
			class="itWasTrue"
		}
		width="300">Content</div>
}

`,
		},
		{
			name: "conditional expressions have their end > indented",
			input: ` // first line removed to make indentation clear
package test

templ conditionalAttributes(addClass bool) {
	<div id="conditional"
if addClass {
class="itWasTrue"
}
>Content</div>
}
`,
			expected: ` // first line removed to make indentation clear
package test

templ conditionalAttributes(addClass bool) {
	<div id="conditional"
		if addClass {
			class="itWasTrue"
		}
		>Content</div>
}

`,
		},
		{
			name: "conditional expressions with else blocks are also formatted",
			input: ` // first line removed to make indentation clear
package test

templ conditionalAttributes(addClass bool) {
	<div id="conditional"
if addClass {
class="itWasTrue"
} else {
	class="itWasNotTrue"
}
width="300">Content</div>
}
`,
			expected: ` // first line removed to make indentation clear
package test

templ conditionalAttributes(addClass bool) {
	<div id="conditional"
		if addClass {
			class="itWasTrue"
		} else {
			class="itWasNotTrue"
		}
		width="300">Content</div>
}

`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Remove the first line of the test data.
			input := strings.SplitN(tt.input, "\n", 2)[1]
			expected := strings.SplitN(tt.expected, "\n", 2)[1]

			// Execute the test.
			template, err := ParseString(input)
			if err != nil {
				t.Fatalf("failed to parse template: %v", err)
			}
			w := new(strings.Builder)
			err = template.Write(w)
			if err != nil {
				t.Fatalf("failed to write template: %v", err)
			}
			if diff := cmp.Diff(expected, w.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}
