package format

import (
	"bytes"
	"regexp"
)

// cleanupBlankLines removes excessive blank lines in specific contexts.
// - Between consecutive @component/@layout blocks
// - Immediately after opening braces in component blocks
// - Between control structures and component calls
func cleanupBlankLines(src []byte) []byte {
	lines := bytes.Split(src, []byte("\n"))
	var result [][]byte

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := bytes.TrimSpace(line)

		// Skip blank line if it's between consecutive component calls
		if len(trimmed) == 0 && i > 0 && i < len(lines)-1 {
			prevLine := bytes.TrimSpace(lines[i-1])
			nextLine := bytes.TrimSpace(lines[i+1])

			// Check if prev and next are both component calls
			if isComponentLine(prevLine) && isComponentLine(nextLine) {
				continue
			}

			// Check if this blank line is immediately after {
			if bytes.HasSuffix(prevLine, []byte("{")) {
				continue
			}

			// Check if this blank line is immediately before }
			if bytes.HasPrefix(nextLine, []byte("}")) || 
			   bytes.HasPrefix(nextLine, []byte(")")) {
				continue
			}

			// Check if between control structure and component
			if isControlStructure(prevLine) && isComponentLine(nextLine) {
				continue
			}
		}

		result = append(result, line)
	}

	return bytes.Join(result, []byte("\n"))
}

func isComponentLine(line []byte) bool {
	return (bytes.Contains(line, []byte("@")) && bytes.Contains(line, []byte("component."))) ||
		(bytes.Contains(line, []byte("@")) && bytes.Contains(line, []byte("layout.")))
}

func isControlStructure(line []byte) bool {
	trimmed := bytes.TrimSpace(line)
	controlKeywords := []string{"if ", "for ", "switch ", "case ", "default:"}
	for _, kw := range controlKeywords {
		if bytes.HasPrefix(trimmed, []byte(kw)) {
			return true
		}
	}
	return false
}

// ensureConsistentBlankLines enforces a single blank line between logical sections.
func ensureConsistentBlankLines(src []byte) []byte {
	// Collapse multiple blank lines to single
	multiBlank := regexp.MustCompile(`\n\n\n+`)
	result := multiBlank.ReplaceAll(src, []byte("\n\n"))
	return result
}
