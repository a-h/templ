package lspcmd

import (
	"strings"
	"testing"

	"github.com/a-h/templ/generator"
	"github.com/a-h/templ/parser/v2"
)

func TestJSXAttributeNameToParameterMapping(t *testing.T) {
	// Test that JSX attribute names are properly mapped to function parameters in source maps
	templContent := `package main

templ Button(text string, onClick string, disabled bool) {
	<button onclick={ onClick } disabled?={ disabled }>{ text }</button>
}

templ TestTemplate() {
	<div>
		<Button text="Click me" onClick={ fmt.Sprintf("handle()") } disabled={ false } />
	</div>
}`

	// Parse the template
	template, err := parser.ParseString(templContent)
	if err != nil {
		t.Fatal(err)
	}

	// Generate with working directory to enable component signature resolution
	var output strings.Builder
	generatorOutput, err := generator.Generate(template, &output, generator.WithWorkingDir("."))
	if err != nil {
		t.Fatalf("Generator failed: %v", err)
	}

	sourceMap := generatorOutput.SourceMap
	if sourceMap == nil {
		t.Fatal("Source map should not be nil")
	}

	t.Logf("Source map contains %d expressions", len(sourceMap.Expressions))

	// Test mapping for "text" attribute name (first parameter)
	textAttrLine := uint32(9)
	textAttrChar := uint32(12) // Position within "text" attribute name
	
	textTargetPos, textFound := sourceMap.TargetPositionFromSource(textAttrLine, textAttrChar)
	if textFound {
		t.Logf("Attribute name 'text' maps to target position: line %d, col %d", textTargetPos.Line, textTargetPos.Col)
	} else {
		t.Errorf("Attribute name 'text' position not found in source map at line %d, char %d", textAttrLine, textAttrChar)
	}

	// Test mapping for "onClick" attribute name (second parameter)
	onClickAttrLine := uint32(9)
	onClickAttrChar := uint32(28) // Position within "onClick" attribute name
	
	onClickTargetPos, onClickFound := sourceMap.TargetPositionFromSource(onClickAttrLine, onClickAttrChar)
	if onClickFound {
		t.Logf("Attribute name 'onClick' maps to target position: line %d, col %d", onClickTargetPos.Line, onClickTargetPos.Col)
	} else {
		t.Errorf("Attribute name 'onClick' position not found in source map at line %d, char %d", onClickAttrLine, onClickAttrChar)
	}

	// Test mapping for "disabled" attribute name (third parameter)
	disabledAttrLine := uint32(9)
	disabledAttrChar := uint32(70) // Position within "disabled" attribute name
	
	disabledTargetPos, disabledFound := sourceMap.TargetPositionFromSource(disabledAttrLine, disabledAttrChar)
	if disabledFound {
		t.Logf("Attribute name 'disabled' maps to target position: line %d, col %d", disabledTargetPos.Line, disabledTargetPos.Col)
	} else {
		t.Errorf("Attribute name 'disabled' position not found in source map at line %d, char %d", disabledAttrLine, disabledAttrChar)
	}

	// Verify that we have source map entries for all attribute names
	hasTextMapping := textFound
	hasOnClickMapping := onClickFound
	hasDisabledMapping := disabledFound

	if hasTextMapping && hasOnClickMapping && hasDisabledMapping {
		t.Log("SUCCESS: All JSX attribute names are properly mapped to function parameters in the source map")
	} else {
		t.Errorf("INCOMPLETE: Not all attribute names are mapped. text: %v, onClick: %v, disabled: %v", 
			hasTextMapping, hasOnClickMapping, hasDisabledMapping)
	}

	// Verify the generated code is correct
	generatedCode := output.String()
	if !strings.Contains(generatedCode, `Button("Click me", fmt.Sprintf("handle()"), false)`) {
		t.Error("Generated code should contain the correct function call with parameters in order")
	}

	t.Log("JSX attribute name to parameter mapping test completed")
}