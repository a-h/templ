package format

import (
	"bytes"
	"regexp"
)

// normalizeComponentSpacing enforces consistent blank line placement around
// @component and @layout blocks.
func normalizeComponentSpacing(src []byte) []byte {
	lines := bytes.Split(src, []byte("\n"))
	var result [][]byte

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := bytes.TrimSpace(line)

		// Check if this is a component or layout call
		isComponentCall := bytes.Contains(trimmed, []byte("@")) && 
			(bytes.Contains(trimmed, []byte("component.")) || bytes.Contains(trimmed, []byte("layout.")))

		// Rule 1: Add blank line before component/layout (unless preceded by {)
		if isComponentCall && i > 0 {
			prevLine := bytes.TrimSpace(lines[i-1])
			if len(prevLine) > 0 && !bytes.HasSuffix(prevLine, []byte("{")) && !bytes.HasSuffix(prevLine, []byte("(")) {
				// Check if previous line is not already blank
				if len(result) > 0 && len(bytes.TrimSpace(result[len(result)-1])) > 0 {
					result = append(result, []byte(""))
				}
			}
		}

		result = append(result, line)

		// Rule 2: Add blank line after multi-line component blocks
		// Look ahead for closing brace that ends a component block
		if isComponentCall && i < len(lines)-1 {
			// Simple heuristic: if next non-empty line is @component or starts with @, add blank line
			for j := i + 1; j < len(lines); j++ {
				next := bytes.TrimSpace(lines[j])
				if len(next) > 0 {
					if (bytes.Contains(next, []byte("@")) && bytes.Contains(next, []byte("component."))) ||
						bytes.HasPrefix(next, []byte("}")) {
						break
					}
					// If we hit another statement, might need spacing
					break
				}
			}
		}
	}

	// Clean up multiple consecutive blank lines
	result = collapseMultipleBlankLines(result)

	// Remove blank lines after opening braces
	result = removeBlankLinesAfterOpeningBrace(result)

	// Remove blank lines before closing braces
	result = removeBlankLinesBeforeClosingBrace(result)

	return bytes.Join(result, []byte("\n"))
}

func collapseMultipleBlankLines(lines [][]byte) [][]byte {
	var result [][]byte
	blankCount := 0

	for _, line := range lines {
		if len(bytes.TrimSpace(line)) == 0 {
			blankCount++
			if blankCount == 1 {
				result = append(result, line)
			}
		} else {
			blankCount = 0
			result = append(result, line)
		}
	}

	return result
}

func removeBlankLinesAfterOpeningBrace(lines [][]byte) [][]byte {
	var result [][]byte

	for i, line := range lines {
		trimmed := bytes.TrimSpace(line)
		result = append(result, line)

		// If this line ends with {, skip next blank lines
		if bytes.HasSuffix(trimmed, []byte("{")) && i < len(lines)-1 {
			for j := i + 1; j < len(lines); j++ {
				next := bytes.TrimSpace(lines[j])
				if len(next) == 0 {
					// Skip this blank line, don't add to result
					continue
				}
				break
			}
		}
	}

	return result
}

func removeBlankLinesBeforeClosingBrace(lines [][]byte) [][]byte {
	var result [][]byte

	for i, line := range lines {
		trimmed := bytes.TrimSpace(line)

		// Check if next non-blank line is a closing brace
		if len(trimmed) == 0 && i < len(lines)-1 {
			for j := i + 1; j < len(lines); j++ {
				next := bytes.TrimSpace(lines[j])
				if len(next) > 0 {
					if bytes.HasPrefix(next, []byte("}")) || 
					   bytes.HasPrefix(next, []byte(")")) {
						// Skip this blank line
						continue
					}
					break
				}
			}
		}
		result = append(result, line)
	}

	return result
}

// normalizeOperatorSpacing standardizes spacing around + operators in
// component method calls and string concatenations.
func normalizeOperatorSpacing(src []byte) []byte {
	// Pattern: + with any amount of whitespace before/after
	re := regexp.MustCompile(`\s*\+\s*`)
	result := re.ReplaceAll(src, []byte(" + "))
	return result
}

// normalizeHTMLTags converts self-closing XML-style tags to HTML5 style.
func normalizeHTMLTags(src []byte) []byte {
	voidElements := []string{"input", "br", "hr", "img", "meta", "link", "col", "area", "base", "embed", "source", "track", "wbr"}

	result := src
	for _, element := range voidElements {
		// Match <element ... /> and replace with <element ... >
		pattern := regexp.MustCompile(`<` + element + `([^>]*?)\s*/\s*>`)
		result = pattern.ReplaceAll(result, []byte("<"+element+"$1>"))
	}

	return result
}
