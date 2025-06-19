package parser

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestDiagnoseFiles tests diagnostics using external .templ files with embedded expectations
// This approach is better for complex scenarios and provides more realistic test cases
func TestDiagnoseFiles(t *testing.T) {
	testDataDir := "testdata/diagnostics"
	
	// Create test data directory if it doesn't exist
	if err := os.MkdirAll(testDataDir, 0755); err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}
	
	// Create test files
	testFiles := map[string]string{
		"valid_components.templ": `package test

templ Button() {
	<button>Click me</button>
}

templ Page() {
	<div>
		<Button />  // Valid component reference
		<p>Hello World</p>
	</div>
}`,

		"missing_components.templ": `package test

templ Page() {
	<div>
		<MissingButton />   // @diagnostic: Component MissingButton not found
		<AnotherMissing />  // @diagnostic: Component AnotherMissing not found
	</div>
}`,

		"mixed_components.templ": `package test

templ ValidButton() {
	<button>Valid</button>
}

templ Page() {
	<div>
		<ValidButton />     // Valid - should not error
		<MissingButton />   // @diagnostic: Component MissingButton not found
		<pkg.External />    // Package qualified - should be ignored
		<var.Method />      // Struct method - should be ignored
	</div>
}`,

		"complex_nesting.templ": `package test

templ Layout() {
	<html>
		<body>
			{ children... }
		</body>
	</html>
}

templ Card() {
	<div class="card">
		{ children... }
	</div>
}

templ Page() {
	@Layout() {
		@Card() {
			<MissingComponent />  // @diagnostic: Component MissingComponent not found
		}
		<AnotherMissing />       // @diagnostic: Component AnotherMissing not found
	}
}`,
	}
	
	// Write test files
	for filename, content := range testFiles {
		filepath := filepath.Join(testDataDir, filename)
		if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write test file %s: %v", filename, err)
		}
	}
	
	// Run tests on each file
	for filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			testDiagnosticFile(t, filepath.Join(testDataDir, filename))
		})
	}
}

// testDiagnosticFile tests a single .templ file with embedded diagnostic expectations
func testDiagnosticFile(t *testing.T, filePath string) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read test file %s: %v", filePath, err)
	}
	
	// Parse expected diagnostics from comments
	expectedDiagnostics := parseExpectedDiagnostics(string(content))
	
	// Parse the template
	tf, err := ParseString(string(content))
	if err != nil {
		t.Fatalf("Failed to parse template from %s: %v", filePath, err)
	}
	
	// Run diagnostics
	actualDiagnostics, err := Diagnose(tf)
	if err != nil {
		t.Fatalf("Diagnose() failed for %s: %v", filePath, err)
	}
	
	// Compare results
	if diff := cmp.Diff(expectedDiagnostics, actualDiagnostics, cmp.AllowUnexported(Position{})); diff != "" {
		t.Errorf("Diagnostics mismatch for %s (-expected +actual):\n%s", filePath, diff)
		
		// Additional debug info
		t.Logf("Expected %d diagnostics:", len(expectedDiagnostics))
		for i, d := range expectedDiagnostics {
			t.Logf("  %d: %s at %+v", i, d.Message, d.Range)
		}
		t.Logf("Actual %d diagnostics:", len(actualDiagnostics))
		for i, d := range actualDiagnostics {
			t.Logf("  %d: %s at %+v", i, d.Message, d.Range)
		}
	}
}

// parseExpectedDiagnostics extracts expected diagnostics from @diagnostic comments
// Format: // @diagnostic: <message>
func parseExpectedDiagnostics(content string) []Diagnostic {
	var diagnostics []Diagnostic
	lines := strings.Split(content, "\n")
	
	// Regex to match diagnostic comments
	diagnosticRegex := regexp.MustCompile(`//\s*@diagnostic:\s*(.+)`)
	
	for lineIndex, line := range lines {
		matches := diagnosticRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			expectedMessage := strings.TrimSpace(matches[1])
			
			// Find the component name that this diagnostic should apply to
			// Look for <ComponentName in the same line
			componentRegex := regexp.MustCompile(`<([A-Z][a-zA-Z0-9]*)\s*`)
			componentMatches := componentRegex.FindStringSubmatch(line)
			
			if len(componentMatches) == 2 {
				componentName := componentMatches[1]
				
				// Calculate position of the component name
				componentIndex := strings.Index(line, "<"+componentName)
				if componentIndex != -1 {
					nameStart := componentIndex + 1 // Skip '<'
					nameEnd := nameStart + len(componentName)
					
					// Calculate absolute positions
					absoluteStart := 0
					for i := 0; i < lineIndex; i++ {
						absoluteStart += len(lines[i]) + 1 // +1 for newline
					}
					absoluteStart += nameStart
					absoluteEnd := absoluteStart + len(componentName)
					
					diagnostics = append(diagnostics, Diagnostic{
						Message: expectedMessage,
						Range: Range{
							From: Position{
								Index: int64(absoluteStart),
								Line:  uint32(lineIndex),
								Col:   uint32(nameStart),
							},
							To: Position{
								Index: int64(absoluteEnd),
								Line:  uint32(lineIndex),
								Col:   uint32(nameEnd),
							},
						},
					})
				}
			}
		}
	}
	
	return diagnostics
}

// Cleanup function to remove test files after tests
func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()
	
	// Cleanup
	os.RemoveAll("testdata/diagnostics")
	
	os.Exit(code)
}