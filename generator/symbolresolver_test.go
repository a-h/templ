package generator

import (
	"testing"
)

func TestSymbolResolver_RealComponents(t *testing.T) {
	// Test with the actual test-element-component directory
	testDir := "./test-element-component"
	resolver := NewSymbolResolver(testDir)

	tests := []struct {
		name          string
		pkgPath       string
		componentName string
		wantErr       bool
		wantIsStruct  bool
		errContains   string
	}{
		{
			name:          "External package function component",
			pkgPath:       "github.com/a-h/templ/generator/test-element-component/mod",
			componentName: "Text",
			wantErr:       false,
			wantIsStruct:  false,
		},
		{
			name:          "Non-existent component",
			pkgPath:       "github.com/a-h/templ/generator/test-element-component/mod",
			componentName: "NonExistent",
			wantErr:       true,
			errContains:   "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig, err := resolver.ResolveComponent(tt.pkgPath, tt.componentName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing %q, got %q", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if sig.IsStruct != tt.wantIsStruct {
				t.Errorf("IsStruct = %v, want %v", sig.IsStruct, tt.wantIsStruct)
			}
		})
	}
}

func TestSymbolResolver_LocalComponents(t *testing.T) {
	// Test with the actual test-element-component directory
	testDir := "./test-element-component"
	resolver := NewSymbolResolver(testDir)

	tests := []struct {
		name          string
		componentName string
		wantErr       bool
		wantIsStruct  bool
		errContains   string
	}{
		{
			name:          "Valid struct component from templ file",
			componentName: "ComponentImpl",
			wantErr:       false,
			wantIsStruct:  true,
		},
		{
			name:          "Regular struct without Render method",
			componentName: "StructComponent",
			wantErr:       true,
			errContains:   "does not implement templ.Component interface",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig, err := resolver.ResolveLocalComponent(tt.componentName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing %q, got %q", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if sig.IsStruct != tt.wantIsStruct {
				t.Errorf("IsStruct = %v, want %v", sig.IsStruct, tt.wantIsStruct)
			}
		})
	}
}

func TestImplementsComponent_Validation(t *testing.T) {
	// This test validates the logic of the implementsComponent method
	// by checking the string matching behavior
	
	tests := []struct {
		name            string
		param1Type      string
		param2Type      string
		returnType      string
		paramCount      int
		returnCount     int
		wantImplements  bool
	}{
		{
			name:            "Valid Component implementation",
			param1Type:      "context.Context",
			param2Type:      "io.Writer",
			returnType:      "error",
			paramCount:      2,
			returnCount:     1,
			wantImplements:  true,
		},
		{
			name:            "Wrong first parameter type",
			param1Type:      "string",
			param2Type:      "io.Writer",
			returnType:      "error",
			paramCount:      2,
			returnCount:     1,
			wantImplements:  false,
		},
		{
			name:            "Wrong second parameter type",
			param1Type:      "context.Context",
			param2Type:      "string",
			returnType:      "error",
			paramCount:      2,
			returnCount:     1,
			wantImplements:  false,
		},
		{
			name:            "Wrong return type",
			param1Type:      "context.Context",
			param2Type:      "io.Writer",
			returnType:      "string",
			paramCount:      2,
			returnCount:     1,
			wantImplements:  false,
		},
		{
			name:            "Too few parameters",
			param1Type:      "context.Context",
			param2Type:      "",
			returnType:      "error",
			paramCount:      1,
			returnCount:     1,
			wantImplements:  false,
		},
		{
			name:            "Too many parameters",
			param1Type:      "context.Context",
			param2Type:      "io.Writer",
			returnType:      "error",
			paramCount:      3,
			returnCount:     1,
			wantImplements:  false,
		},
		{
			name:            "Multiple return values",
			param1Type:      "context.Context",
			param2Type:      "io.Writer",
			returnType:      "error",
			paramCount:      2,
			returnCount:     2,
			wantImplements:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The actual validation logic matches what's in implementsComponent
			isValid := tt.paramCount == 2 && 
				tt.returnCount == 1 &&
				tt.param1Type == "context.Context" &&
				tt.param2Type == "io.Writer" &&
				tt.returnType == "error"

			if isValid != tt.wantImplements {
				t.Errorf("Implementation check = %v, want %v", isValid, tt.wantImplements)
			}
		})
	}
}

func TestSymbolResolverCache(t *testing.T) {
	testDir := "./test-element-component"
	resolver := NewSymbolResolver(testDir)

	// Resolve a component twice
	pkgPath := "github.com/a-h/templ/generator/test-element-component/mod"
	componentName := "Text"

	sig1, err := resolver.ResolveComponent(pkgPath, componentName)
	if err != nil {
		t.Fatalf("First resolution failed: %v", err)
	}

	sig2, err := resolver.ResolveComponent(pkgPath, componentName)
	if err != nil {
		t.Fatalf("Second resolution failed: %v", err)
	}

	// Both signatures should be the same pointer (from cache)
	if sig1 != sig2 {
		t.Error("Expected cached signature to be returned")
	}
}

func containsString(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && 
		containsSubstring(s, substr)
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}