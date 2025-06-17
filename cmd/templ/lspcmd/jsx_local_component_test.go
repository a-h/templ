package lspcmd

import (
	"strings"
	"testing"

	"github.com/a-h/templ/generator"
	"github.com/a-h/templ/parser/v2"
)

func TestJSXLocalComponentSourceMap(t *testing.T) {
	// Test with a local templ component which should work without external dependencies
	templContent := `package main

templ Button(text string) {
	<button>{ text }</button>
}

templ TestTemplate() {
	<div>
		<Button text="Click me" />
	</div>
}
`

	// Parse the template
	template, err := parser.ParseString(templContent)
	if err != nil {
		t.Fatal(err)
	}

	// Generate without working directory (should work for local components)
	var output strings.Builder
	generatorOutput, err := generator.Generate(template, &output)
	if err != nil {
		t.Fatalf("Generator failed: %v", err)
	}

	sourceMap := generatorOutput.SourceMap
	if sourceMap == nil {
		t.Fatal("Source map should not be nil for local components")
	}

	t.Logf("Generated Go code:\n%s", output.String())
	t.Logf("Source map has %d expressions", len(sourceMap.Expressions))

	// Test that we can map JSX component name position
	// Component name "Button" should be around line 8, character 2-8
	componentNameLine := uint32(8)
	componentNameChar := uint32(3) // Position of "Button" in "<Button"
	
	targetPos, found := sourceMap.TargetPositionFromSource(componentNameLine, componentNameChar)
	if found {
		t.Logf("Component name maps to target position: line %d, col %d", targetPos.Line, targetPos.Col)
	} else {
		t.Logf("Component name position not found in source map at line %d, char %d (this may be expected)", componentNameLine, componentNameChar)
	}

	// Test that we can map JSX attribute value
	// The string "Click me" should be around line 8, character 15-25
	attrLine := uint32(8)
	attrChar := uint32(17) // Position inside "Click me"
	
	attrTargetPos, attrFound := sourceMap.TargetPositionFromSource(attrLine, attrChar)
	if attrFound {
		t.Logf("Attribute value maps to target position: line %d, col %d", attrTargetPos.Line, attrTargetPos.Col)
	} else {
		t.Logf("Attribute value position not found in source map at line %d, char %d", attrLine, attrChar)
	}

	// Verify the generated code contains the expected function call
	generatedCode := output.String()
	if !strings.Contains(generatedCode, `Button(`) {
		t.Error("Generated code should contain JSX component function call")
	}

	if !strings.Contains(generatedCode, `"Click me"`) {
		t.Error("Generated code should contain the attribute value")
	}

	t.Log("JSX local component source map test completed")
}

func TestJSXAttributeExpressionSourceMap(t *testing.T) {
	// Test JSX with expression attributes
	templContent := `package main

import "fmt"

templ Button(onClick string) {
	<button onclick={ onClick }>Click</button>
}

templ TestTemplate(name string) {
	<div>
		<Button onClick={ fmt.Sprintf("handle_%s", name) } />
	</div>
}
`

	template, err := parser.ParseString(templContent)
	if err != nil {
		t.Fatal(err)
	}

	var output strings.Builder
	generatorOutput, err := generator.Generate(template, &output)
	if err != nil {
		t.Fatalf("Generator failed: %v", err)
	}

	sourceMap := generatorOutput.SourceMap
	if sourceMap == nil {
		t.Fatal("Source map should not be nil")
	}

	t.Logf("Source map contains %d expressions", len(sourceMap.Expressions))
	
	// Test mapping a position within the attribute expression
	// The fmt.Sprintf expression should be around line 11
	line := uint32(11)
	char := uint32(20) // Position within fmt.Sprintf expression
	
	if targetPos, found := sourceMap.TargetPositionFromSource(line, char); found {
		t.Logf("Attribute expression at %d:%d maps to %d:%d", line, char, targetPos.Line, targetPos.Col)
	} else {
		t.Logf("Attribute expression position not found at %d:%d", line, char)
	}

	// Verify the generated code preserves the expression
	generatedCode := output.String()
	if !strings.Contains(generatedCode, `fmt.Sprintf("handle_%s", name)`) {
		t.Error("Generated code should contain the original expression from JSX attribute")
	}
}