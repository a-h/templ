package generator

import (
	"go/types"
	"strings"

	"github.com/a-h/templ/parser/v2"
)

// TemplSignatureResolver extracts component signatures from templ template definitions
type TemplSignatureResolver struct {
	signatures map[string]*ComponentSignature
}

// NewTemplSignatureResolver creates a new templ signature resolver
func NewTemplSignatureResolver() *TemplSignatureResolver {
	return &TemplSignatureResolver{
		signatures: make(map[string]*ComponentSignature),
	}
}

// ExtractSignatures walks through a templ file and extracts all template signatures
func (tsr *TemplSignatureResolver) ExtractSignatures(tf *parser.TemplateFile) {
	for _, node := range tf.Nodes {
		switch n := node.(type) {
		case *parser.HTMLTemplate:
			sig := tsr.extractHTMLTemplateSignature(n)
			if sig != nil {
				tsr.signatures[sig.Name] = sig
			}
		}
	}
}

// GetSignature returns the signature for a template name
func (tsr *TemplSignatureResolver) GetSignature(name string) (*ComponentSignature, bool) {
	sig, ok := tsr.signatures[name]
	return sig, ok
}

// extractHTMLTemplateSignature extracts the signature from an HTML template
func (tsr *TemplSignatureResolver) extractHTMLTemplateSignature(tmpl *parser.HTMLTemplate) *ComponentSignature {
	// Parse the template declaration from Expression.Value
	// Format: "templateName(param1 type1, param2 type2)"
	exprValue := tmpl.Expression.Value
	if exprValue == "" {
		return nil
	}

	name, params := tsr.parseTemplateDeclaration(exprValue)
	if name == "" {
		return nil
	}

	return &ComponentSignature{
		PackagePath: "", // Local package
		Name:        name,
		Parameters:  params,
	}
}

// parseTemplateDeclaration parses a templ template declaration like "Button(title string)" 
func (tsr *TemplSignatureResolver) parseTemplateDeclaration(decl string) (name string, params []ParameterInfo) {
	decl = strings.TrimSpace(decl)
	
	// Find the opening parenthesis
	parenIdx := strings.Index(decl, "(")
	if parenIdx == -1 {
		// No parameters
		return strings.TrimSpace(decl), nil
	}
	
	name = strings.TrimSpace(decl[:parenIdx])
	
	// Find the closing parenthesis
	closeParenIdx := strings.LastIndex(decl, ")")
	if closeParenIdx == -1 || closeParenIdx <= parenIdx {
		// Malformed declaration
		return name, nil
	}
	
	paramStr := strings.TrimSpace(decl[parenIdx+1 : closeParenIdx])
	if paramStr == "" {
		return name, nil
	}
	
	params = tsr.parseTemplateParameters(paramStr)
	return name, params
}

// parseTemplateParameters parses templ template parameter strings like "term, detail string" or "title string, count int"
func (tsr *TemplSignatureResolver) parseTemplateParameters(paramStr string) []ParameterInfo {
	paramStr = strings.TrimSpace(paramStr)
	if paramStr == "" {
		return nil
	}

	params := make([]ParameterInfo, 0)
	
	// Handle Go-style parameter syntax: "name1, name2 type1, name3 type2"
	// Split by commas first to get all individual parts
	parts := strings.Split(paramStr, ",")
	
	i := 0
	for i < len(parts) {
		part := strings.TrimSpace(parts[i])
		if part == "" {
			i++
			continue
		}
		
		// Look ahead to see if this is a "name type" pattern or just a "name" that shares type with the next parts
		words := strings.Fields(part)
		if len(words) == 2 {
			// This is "name type" - standalone parameter
			params = append(params, ParameterInfo{
				Name: words[0],
				Type: types.Typ[types.String], // For now, assume string type
			})
			i++
		} else if len(words) == 1 {
			// This is just a name, need to collect names until we find a type
			names := []string{words[0]}
			i++
			
			// Collect additional names that share the same type
			for i < len(parts) {
				nextPart := strings.TrimSpace(parts[i])
				nextWords := strings.Fields(nextPart)
				if len(nextWords) == 2 {
					// Found "name type" - the type applies to all collected names
					names = append(names, nextWords[0])
					_ = nextWords[1] // typeName - we assume string for now
					
					// Create parameters for all names with this type
					for _, name := range names {
						params = append(params, ParameterInfo{
							Name: name,
							Type: types.Typ[types.String], // For now, assume string type regardless of typeName
						})
					}
					i++
					break
				} else if len(nextWords) == 1 {
					// Another name sharing the type
					names = append(names, nextWords[0])
					i++
				} else {
					// Malformed
					i++
					break
				}
			}
		} else {
			// Malformed parameter
			i++
		}
	}
	
	return params
}

// parseParameterGroup parses a parameter group like "name type" or handles complex cases
func (tsr *TemplSignatureResolver) parseParameterGroup(group string) []ParameterInfo {
	// For now, handle simple cases - this could be enhanced to handle more complex Go parameter syntax
	parts := strings.Fields(group)
	if len(parts) < 2 {
		return nil
	}
	
	// Check if this is a "name1, name2 type" format by looking for the last part as type
	nameCount := len(parts) - 1
	
	params := make([]ParameterInfo, 0, nameCount)
	for i := 0; i < nameCount; i++ {
		name := strings.TrimSuffix(parts[i], ",")
		if name != "" {
			params = append(params, ParameterInfo{
				Name: name,
				Type: types.Typ[types.String], // For now, assume string type - could be enhanced
			})
		}
	}
	
	return params
}