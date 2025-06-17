package generator

import (
	"fmt"
	"go/build"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

// ComponentSignature represents a templ component's function signature
type ComponentSignature struct {
	PackagePath string
	Name        string
	Parameters  []ParameterInfo
}

// ParameterInfo represents a function parameter
type ParameterInfo struct {
	Name     string
	Type     types.Type
	Position int // Parameter position in function signature (0-based)
}

// SymbolResolver resolves component symbols across packages
type SymbolResolver struct {
	cache      map[string]*ComponentSignature
	workingDir string
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
	if len(pkg.Errors) > 0 {
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

	for i := range params.Len() {
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
