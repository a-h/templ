package format

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/a-h/templ/internal/prettier"
	parser "github.com/a-h/templ/parser/v2"
	"github.com/google/go-cmp/cmp"
	nethtml "golang.org/x/net/html"
)

// TestPrettierBatchStartsPrettierOncePerFile guards against performance regressions. Starting
// prettier takes over 100ms, so formatting must not start it once per template or element.
func TestPrettierBatchStartsPrettierOncePerFile(t *testing.T) {
	if !prettier.IsAvailable(prettier.DefaultCommand()) {
		// Skipping in CI would remove the guard without anyone noticing.
		if os.Getenv("CI") != "" {
			t.Fatal("Prettier is not available, but is required in CI")
		}
		t.Skip("Prettier is not available, skipping prettier invocation count tests")
	}
	tests := []struct {
		name     string
		src      string
		expected int
	}{
		{
			name:     "a file without attributes, scripts, or styles does not start prettier",
			src:      "package main\n\ntempl A() {\n\t<div></div>\n}\n",
			expected: 0,
		},
		{
			name:     "a file with one component starts prettier once",
			src:      createBenchmarkTemplate(1),
			expected: 1,
		},
		{
			name:     "a file with many components, a script, and a style starts prettier once",
			src:      createBenchmarkTemplate(50),
			expected: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			countFile := filepath.Join(t.TempDir(), "count.txt")
			config := Config{
				PrettierCommand:  "echo x >> " + countFile + "; " + prettier.DefaultCommand(),
				PrettierRequired: true,
			}
			if _, _, err := Templ([]byte(tt.src), "", config); err != nil {
				t.Fatalf("failed to format: %v", err)
			}
			count, err := os.ReadFile(countFile)
			if err != nil && !os.IsNotExist(err) {
				t.Fatalf("failed to read count file: %v", err)
			}
			if diff := cmp.Diff(tt.expected, strings.Count(string(count), "x")); diff != "" {
				t.Errorf("unexpected number of prettier invocations (-expected +got):\n%s", diff)
			}
		})
	}
}

func TestPrettierBatchFormatsAttributesScriptsAndStylesTogether(t *testing.T) {
	if !prettier.IsAvailable(prettier.DefaultCommand()) {
		t.Skip("Prettier is not available, skipping formatting tests")
	}
	src := "package main\n\n" +
		"templ A() {\n\t<div style=\"color:red\">\n\t\t<script>let  x=1</script>\n\t</div>\n}\n\n" +
		"templ B() {\n\t<style>p{margin:0}</style>\n\t<script>let  y=2</script>\n}\n"
	expected := "package main\n\n" +
		"templ A() {\n\t<div style=\"color: red\">\n\t\t<script>\n\t\t\tlet x = 1;\n\t\t</script>\n\t</div>\n}\n\n" +
		"templ B() {\n\t<style>\n\t\tp {\n\t\t\tmargin: 0;\n\t\t}\n\t</style>\n\t<script>\n\t\tlet y = 2;\n\t</script>\n}\n"

	actual, _, err := Templ([]byte(src), "", Config{PrettierRequired: true})
	if err != nil {
		t.Fatalf("failed to format: %v", err)
	}
	if diff := cmp.Diff(expected, string(actual)); diff != "" {
		t.Errorf("unexpected output (-expected +got):\n%s", diff)
	}
}

func TestPrettierBatchLeavesElementsUnchangedWhenPrettierOutputMatchesInput(t *testing.T) {
	content := "\n\tconsole.log(\"hello\");\n"
	tests := []struct {
		name  string
		depth int
	}{
		{name: "elements at depth 0 are unchanged", depth: 0},
		{name: "elements at depth 2 are unchanged", depth: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := &parser.ScriptElement{Contents: []parser.ScriptContents{parser.NewScriptContentsScriptCode(content)}}
			style := &parser.RawElement{Name: "style", Contents: content}
			batch := newPrettierBatch("cat")
			batch.AddScriptElement(script, tt.depth)
			batch.AddStyleElement(style, tt.depth)
			if err := batch.Run(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			expected := []parser.ScriptContents{parser.NewScriptContentsScriptCode(content)}
			if diff := cmp.Diff(expected, script.Contents); diff != "" {
				t.Errorf("unexpected script contents (-expected +got):\n%s", diff)
			}
			if diff := cmp.Diff(content, style.Contents); diff != "" {
				t.Errorf("unexpected style contents (-expected +got):\n%s", diff)
			}
		})
	}
}

func TestFindEntryNodes(t *testing.T) {
	tests := []struct {
		name          string
		html          string
		entryCount    int
		expectedIDs   []string
		expectedError string
	}{
		{
			name:        "elements are returned in entry order regardless of document order",
			html:        `<p data-templ-script-id="1" id="b"></p><p data-templ-script-id="0" id="a"></p>`,
			entryCount:  2,
			expectedIDs: []string{"a", "b"},
		},
		{
			name:        "elements without the id key are ignored",
			html:        `<p data-templ-style-id="0" id="x"></p><p data-templ-script-id="0" id="a"></p>`,
			entryCount:  1,
			expectedIDs: []string{"a"},
		},
		{
			name:          "a missing element is an error",
			html:          `<p data-templ-script-id="0" id="a"></p>`,
			entryCount:    2,
			expectedError: "prettier output is missing the element with data-templ-script-id 1",
		},
		{
			name:          "a duplicated element is an error",
			html:          `<p data-templ-script-id="0"></p><p data-templ-script-id="0"></p>`,
			entryCount:    1,
			expectedError: "prettier output contains data-templ-script-id 0 more than once",
		},
		{
			name:          "an out of range index is an error",
			html:          `<p data-templ-script-id="5"></p>`,
			entryCount:    1,
			expectedError: "data-templ-script-id 5 out of range [0, 1)",
		},
		{
			name:          "a non-integer index is an error",
			html:          `<p data-templ-script-id="x"></p>`,
			entryCount:    1,
			expectedError: `failed to parse data-templ-script-id: strconv.Atoi: parsing "x": invalid syntax`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := nethtml.Parse(strings.NewReader(tt.html))
			if err != nil {
				t.Fatalf("failed to parse HTML: %v", err)
			}
			entryNodes, err := findEntryNodes(doc, scriptIDKey, tt.entryCount)
			var actualError string
			if err != nil {
				actualError = err.Error()
			}
			if diff := cmp.Diff(tt.expectedError, actualError); diff != "" {
				t.Fatalf("unexpected error (-expected +got):\n%s", diff)
			}
			var actualIDs []string
			for _, n := range entryNodes {
				id, ok := getAttr(n, "id")
				if !ok {
					t.Fatalf("entry node has no id attribute")
				}
				actualIDs = append(actualIDs, id)
			}
			if diff := cmp.Diff(tt.expectedIDs, actualIDs); diff != "" {
				t.Errorf("unexpected entry nodes (-expected +got):\n%s", diff)
			}
		})
	}
}
