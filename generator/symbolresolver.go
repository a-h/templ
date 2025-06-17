package generator

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/types"
	"strings"
	"path/filepath"
	"os"

	"golang.org/x/tools/go/packages"
	"github.com/a-h/templ/parser/v2"
)

// ComponentSignature represents a templ component's function signature
type ComponentSignature struct {
	PackagePath string
	Name        string
	Parameters  []ParameterInfo
}

// ParameterInfo represents a function parameter
type ParameterInfo struct {
	Name string
	Type types.Type
}

// SymbolResolver resolves component symbols across packages
type SymbolResolver struct {
	cache          map[string]*ComponentSignature
	workingDir     string
	templCache     map[string]*TemplSignatureResolver // Cache for external templ packages
}

// NewSymbolResolver creates a new symbol resolver
func NewSymbolResolver(workingDir string) *SymbolResolver {
	return &SymbolResolver{
		cache:      make(map[string]*ComponentSignature),
		workingDir: workingDir,
	}
}

// ResolveComponent resolves a component's function signature
func (sr *SymbolResolver) ResolveComponent(pkgPath, componentName string) (*ComponentSignature, error) {
	key := fmt.Sprintf("%s.%s", pkgPath, componentName)
	if sig, ok := sr.cache[key]; ok {
		return sig, nil
	}

	// Load the package
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo,
		Dir: sr.workingDir,
	}

	pkgs, err := packages.Load(cfg, pkgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load package %s: %w", pkgPath, err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("package %s not found", pkgPath)
	}

	pkg := pkgs[0]
	// Allow packages with errors if they're compilation errors from generated files
	// We'll still try to find the function in the type information
	if pkg.Errors != nil && len(pkg.Errors) > 0 {
		// Check if errors are from _templ.go files - if so, we can ignore them
		hasNonTemplErrors := false
		for _, err := range pkg.Errors {
			errStr := err.Error()
			if !strings.Contains(errStr, "_templ.go") {
				hasNonTemplErrors = true
				break
			}
		}
		if hasNonTemplErrors {
			return nil, fmt.Errorf("package %s has non-generated file errors: %v", pkgPath, pkg.Errors)
		}
		// Continue with type checking even if there are _templ.go errors
	}

	// Look for the component function
	obj := pkg.Types.Scope().Lookup(componentName)
	if obj == nil {
		return nil, fmt.Errorf("component %s not found in package %s", componentName, pkgPath)
	}

	// Check if it's a function
	fn, ok := obj.(*types.Func)
	if !ok {
		return nil, fmt.Errorf("%s is not a function", componentName)
	}

	sig := fn.Type().(*types.Signature)
	
	// Extract parameter information
	params := sig.Params()
	paramInfo := make([]ParameterInfo, 0, params.Len())
	
	for i := 0; i < params.Len(); i++ {
		param := params.At(i)
		paramInfo = append(paramInfo, ParameterInfo{
			Name: param.Name(),
			Type: param.Type(),
		})
	}

	componentSig := &ComponentSignature{
		PackagePath: pkgPath,
		Name:        componentName,
		Parameters:  paramInfo,
	}

	sr.cache[key] = componentSig
	return componentSig, nil
}

// ResolveLocalComponent resolves a component in the current package
func (sr *SymbolResolver) ResolveLocalComponent(componentName string) (*ComponentSignature, error) {
	// Get the current package path
	pkg, err := build.ImportDir(sr.workingDir, build.FindOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to determine package path: %w", err)
	}

	return sr.ResolveComponent(pkg.ImportPath, componentName)
}

// ResolveExternalTemplComponent resolves a component from an external package
// following the strategy: same project -> use templ files, external project -> use Go symbols
func (sr *SymbolResolver) ResolveExternalTemplComponent(pkgPath, componentName string) (*ComponentSignature, error) {
	// Check if this is from the same project
	if sr.isSameProject(pkgPath) {
		// Same project - try to resolve from templ files first
		if sig, err := sr.resolveFromTemplFiles(pkgPath, componentName); err == nil {
			return sig, nil
		}
		// Fall back to Go symbols if templ resolution fails
	}
	
	// External project or fallback - only use Go symbols
	return sr.ResolveComponent(pkgPath, componentName)
}

// isSameProject checks if the given package path belongs to the same project
func (sr *SymbolResolver) isSameProject(pkgPath string) bool {
	// This project is github.com/a-h/templ
	return strings.HasPrefix(pkgPath, "github.com/a-h/templ")
}

// resolveFromTemplFiles attempts to resolve component from templ files in the same project
func (sr *SymbolResolver) resolveFromTemplFiles(pkgPath, componentName string) (*ComponentSignature, error) {
	// Get or create cached resolver for this package
	if sr.templCache == nil {
		sr.templCache = make(map[string]*TemplSignatureResolver)
	}
	
	templResolver, exists := sr.templCache[pkgPath]
	if !exists {
		// Create new resolver for this package
		templResolver = NewTemplSignatureResolver()
		
		// Find and parse templ files in the package directory
		if err := sr.loadTemplFilesForPackage(pkgPath, templResolver); err != nil {
			return nil, err
		}
		
		sr.templCache[pkgPath] = templResolver
	}
	
	// Try to get signature from templ resolver
	if sig, ok := templResolver.GetSignature(componentName); ok {
		return sig, nil
	}
	
	return nil, fmt.Errorf("component %s not found in templ files for package %s", componentName, pkgPath)
}

// loadTemplFilesForPackage loads and parses all templ files in the given package
func (sr *SymbolResolver) loadTemplFilesForPackage(pkgPath string, templResolver *TemplSignatureResolver) error {
	// Convert package path to file system path
	pkgDir, err := sr.packagePathToDir(pkgPath)
	if err != nil {
		return err
	}
	
	// Find all .templ files in the directory
	templFiles, err := filepath.Glob(filepath.Join(pkgDir, "*.templ"))
	if err != nil {
		return fmt.Errorf("failed to find templ files in %s: %w", pkgDir, err)
	}
	
	// Parse each templ file and extract signatures
	for _, templFile := range templFiles {
		if err := sr.parseTemplFile(templFile, templResolver); err != nil {
			// Log error but continue with other files
			continue
		}
	}
	
	return nil
}

// packagePathToDir converts a Go package path to a file system directory path
func (sr *SymbolResolver) packagePathToDir(pkgPath string) (string, error) {
	// For same-project packages, we can derive the path from the working directory
	// github.com/a-h/templ/generator/test-jsx/jsxmod -> generator/test-jsx/jsxmod
	
	if !strings.HasPrefix(pkgPath, "github.com/a-h/templ") {
		return "", fmt.Errorf("package %s is not in the same project", pkgPath)
	}
	
	// Remove the project prefix
	relativePath := strings.TrimPrefix(pkgPath, "github.com/a-h/templ")
	relativePath = strings.TrimPrefix(relativePath, "/")
	
	if relativePath == "" {
		// Root package
		return sr.workingDir, nil
	}
	
	return filepath.Join(sr.workingDir, relativePath), nil
}

// parseTemplFile parses a single templ file and extracts component signatures
func (sr *SymbolResolver) parseTemplFile(templFile string, templResolver *TemplSignatureResolver) error {
	// Read the file content
	content, err := os.ReadFile(templFile)
	if err != nil {
		return err
	}
	
	// Parse the templ file
	templateFile, err := parser.Parse(string(content))
	if err != nil {
		return err
	}
	
	// Extract signatures from the parsed file
	templResolver.ExtractSignatures(templateFile)
	
	return nil
}

// CollectJSXComponents walks the AST and collects all JSX component references
func CollectJSXComponents(node ast.Node) []ComponentReference {
	var refs []ComponentReference
	
	// This would walk the templ AST, not Go AST
	// We'll need to implement this based on the templ parser AST
	
	return refs
}