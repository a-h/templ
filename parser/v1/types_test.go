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
			name: "there are two line breaks after the package statement",
			input: ` // first line removed to make indentation clear in Go code
{% package test %}
{% import "strings" %}`,
			expected: `// first line removed to make indentation clear in Go code
{% package test %}

{% import "strings" %}

`,
		},
		{
			name: "import statements are all placed next to each other, and there are two lines after",
			input: ` // first line removed to make indentation clear in Go code
{% package test %}
{% import "strings" %}

{% import "net/url" %}


{% import "rand" %}
`,
			expected: `// first line removed to make indentation clear in Go code
{% package test %}

{% import "strings" %}
{% import "net/url" %}
{% import "rand" %}

`,
		},
		{
			name: "import statements don't have whitespace before or after them",
			input: ` // first line removed to make indentation clear in Go code
{% package test %}

  {% import "net/url" %}  
`,
			expected: `// first line removed to make indentation clear in Go code
{% package test %}

{% import "net/url" %}

`,
		},
		{
			name: "void elements are converted to self-closing elements",
			input: ` // first line removed to make indentation clear in Go code
{% package test %}

{% templ input(value, validation string) %}
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

{% endtempl %}
`,
			expected: `// first line removed to make indentation clear in Go code
{% package test %}

{% templ input(value, validation string) %}
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
{% endtempl %}

`,
		},
		{
			name: "empty elements stay on the same line",
			input: ` // first line removed to make indentation clear in Go code
{% package test %}

{% templ input(value, validation string) %}
<div>
<p>
</p>
</div>
{% endtempl %}
`,
			expected: `// first line removed to make indentation clear in Go code
{% package test %}

{% templ input(value, validation string) %}
	<div>
		<p></p>
	</div>
{% endtempl %}

`,
		},
		{
			name: "if the element only contains inline elements, they end up on the same line",
			input: ` // first line removed to make indentation clear in Go code
{% package test %}

{% templ input(value, validation string) %}
<div><p>{%= "the" %}<a href="http://example.com">{%= "data" %}</a></p></div>
{% endtempl %}
`,
			expected: `// first line removed to make indentation clear in Go code
{% package test %}

{% templ input(value, validation string) %}
	<div>
		<p>{%= "the" %}<a href="http://example.com">{%= "data" %}</a></p>
	</div>
{% endtempl %}

`,
		},
		{
			name: "if an element contains any block elements, all of the child elements are split onto new lines",
			input: ` // first line removed to make indentation clear in Go code
{% package test %}

{% templ nested() %}
<div>{%= "the" %}<div>{%= "other" %}</div></div>
{% endtempl %}
`,
			expected: `// first line removed to make indentation clear in Go code
{% package test %}

{% templ nested() %}
	<div>
		{%= "the" %}
		<div>{%= "other" %}</div>
	</div>
{% endtempl %}

`,
		},
		{
			name: "for loops are placed on a new line",
			input: ` // first line removed to make indentation clear in Go code
{% package test %}

{% templ input(items []string) %}
<div>{%= "the" %}<div>{%= "other" %}</div>{% for _, item := range items %}
		<div>{%= item %}</div>
	{% endfor %}</div>
{% endtempl %}
`,
			expected: `// first line removed to make indentation clear in Go code
{% package test %}

{% templ input(items []string) %}
	<div>
		{%= "the" %}
		<div>{%= "other" %}</div>
		{% for _, item := range items %}
			<div>{%= item %}</div>
		{% endfor %}
	</div>
{% endtempl %}

`,
		},
		{
			name: "css is indented by one level",
			input: ` // first line removed to make indentation clear in Go code
{% package test %}

{% css ClassName() %}
background-color: #ffffff;
color: {%= constants.White %};
{% endcss %}
`,
			expected: `// first line removed to make indentation clear in Go code
{% package test %}

{% css ClassName() %}
	background-color: #ffffff;
	color: {%= constants.White %};
{% endcss %}

`,
		},
		{
			name: "css whitespace is tidied",
			input: ` // first line removed to make indentation clear in Go code
{% package test %}

{% css ClassName() %}
  background-color    :   #ffffff	;
	color	:  {%= constants.White %};
{% endcss %}
`,
			expected: `// first line removed to make indentation clear in Go code
{% package test %}

{% css ClassName() %}
	background-color: #ffffff;
	color: {%= constants.White %};
{% endcss %}

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
