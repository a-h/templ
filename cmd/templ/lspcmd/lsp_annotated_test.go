package lspcmd

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/a-h/templ/generator"
	"github.com/a-h/templ/parser/v2"
)

// TestLSPAnnotatedTempl tests diagnostics using simple message-based annotation format
// This test focuses on the diagnostic messages rather than exact positions
func TestLSPAnnotatedTempl(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping annotated file tests in short mode")
	}
	testDataDir := ".testdata"

	// Find all templ test files
	files, err := filepath.Glob(filepath.Join(testDataDir, "*.templ"))
	if err != nil {
		t.Fatalf("Failed to find test files: %v", err)
	}

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			testAnnotatedMessageFile(t, file)
		})
	}
}

// testAnnotatedMessageFile tests a single annotated file by comparing messages
func testAnnotatedMessageFile(t *testing.T, filePath string) {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read test file %s: %v", filePath, err)
	}

	// Parse expected diagnostics
	expectedMessages := parseExpectedDiagnosticMessages(string(content))

	// Skip files without annotations
	if len(expectedMessages) == 0 {
		t.Skipf("No diagnostic annotations found in %s", filePath)
		return
	}

	// Parse template
	tf, err := parser.ParseString(string(content))
	if err != nil {
		t.Fatalf("Failed to parse template from %s: %v", filePath, err)
	}

	// Run enhanced diagnostics that can find Go types implementing templ.Component
	workingDir := ".testdata" // Use testdata directory as working directory
	actualDiagnostics, err := generator.DiagnoseWithSymbolResolution(tf, workingDir)
	if err != nil {
		t.Fatalf("DiagnoseWithSymbolResolution() failed for %s: %v", filePath, err)
	}

	// Extract actual messages
	actualMessages := make([]string, len(actualDiagnostics))
	for i, d := range actualDiagnostics {
		actualMessages[i] = d.Message
	}

	// Compare counts
	if len(expectedMessages) != len(actualMessages) {
		t.Errorf("Diagnostic count mismatch for %s: expected %d, got %d",
			filePath, len(expectedMessages), len(actualMessages))
		t.Logf("Expected: %v", expectedMessages)
		t.Logf("Actual: %v", actualMessages)
		return
	}

	// Check that all expected messages are present
	for _, expected := range expectedMessages {
		found := slices.Contains(actualMessages, expected)
		if !found {
			t.Errorf("Expected diagnostic message not found in %s: %s", filePath, expected)
			t.Logf("Expected: %v", expectedMessages)
			t.Logf("Actual: %v", actualMessages)
		}
	}
}

// parseExpectedDiagnosticMessages extracts expected diagnostic messages from comments
// Format: // @expect-diagnostic: <message>
func parseExpectedDiagnosticMessages(content string) []string {
	var messages []string
	lines := strings.Split(content, "\n")

	// Regex to match expectation comments
	expectationRegex := regexp.MustCompile(`//\s*@expect-diagnostic:\s*(.+)`)

	for _, line := range lines {
		matches := expectationRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			message := strings.TrimSpace(matches[1])
			messages = append(messages, message)
		}
	}

	return messages
}
