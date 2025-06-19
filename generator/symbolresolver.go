package generator

import (
	"fmt"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

// ComponentSignature represents a templ component's function signature
type ComponentSignature struct {
	PackagePath string
	Name        string
	Parameters  []ParameterInfo
	IsStruct    bool
}

// ParameterInfo represents a function parameter
type ParameterInfo struct {
	Name string
	Type string
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

	var paramInfo []ParameterInfo

	// Check if it's a function
	fn, isFn := obj.(*types.Func)
	if isFn {
		sig := fn.Type().(*types.Signature)

		// Extract parameter information
		params := sig.Params()
		paramInfo = make([]ParameterInfo, 0, params.Len())

		for i := range params.Len() {
			param := params.At(i)
			paramInfo = append(paramInfo, ParameterInfo{
				Name: param.Name(),
				Type: param.Type().String(),
			})
		}
	} else {
		// Check if it's a type that implements templ.Component
		typeName, ok := obj.(*types.TypeName)
		if !ok {
			return nil, fmt.Errorf("%s is neither a function nor a type", componentName)
		}

		// Check if the type implements the templ.Component interface
		implements := sr.implementsComponent(typeName.Type(), pkg.Types)
		if !implements {
			return nil, fmt.Errorf("%s does not implement templ.Component interface", componentName)
		}

		// For types that implement Component, they don't have parameters
		paramInfo = make([]ParameterInfo, 0)
	}

	componentSig := &ComponentSignature{
		PackagePath: pkgPath,
		Name:        componentName,
		Parameters:  paramInfo,
		IsStruct:    !isFn,
	}

	sr.cache[key] = componentSig
	return componentSig, nil
}

// ResolveLocalComponent resolves a component in the current package
func (sr *SymbolResolver) ResolveLocalComponent(componentName string) (*ComponentSignature, error) {
	// Use packages.Load to get the correct package path in module mode
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedModule,
		Dir:  sr.workingDir,
	}

	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to load current package: %w", err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no package found in current directory")
	}

	pkgPath := pkgs[0].PkgPath

	return sr.ResolveComponent(pkgPath, componentName)
}

// implementsComponent checks if a type implements the templ.Component interface
func (sr *SymbolResolver) implementsComponent(t types.Type, pkg *types.Package) bool {
	// Define the templ.Component interface
	var componentInterface *types.Interface

	// Try to find the templ package
	for _, imp := range pkg.Imports() {
		if imp.Path() == "github.com/a-h/templ" {
			obj := imp.Scope().Lookup("Component")
			if obj != nil {
				if iface, ok := obj.Type().Underlying().(*types.Interface); ok {
					componentInterface = iface
					break
				}
			}
		}
	}

	// If we couldn't find the interface, check for the Render method manually
	if componentInterface == nil {
		// Check if the type has a Render method with the correct signature
		// Render(ctx context.Context, w io.Writer) error
		method, _, _ := types.LookupFieldOrMethod(t, true, pkg, "Render")
		if method == nil {
			return false
		}

		fn, ok := method.(*types.Func)
		if !ok {
			return false
		}

		sig := fn.Type().(*types.Signature)

		// Check parameters: (context.Context, io.Writer)
		params := sig.Params()
		if params.Len() != 2 {
			return false
		}

		// Check first parameter is context.Context
		param1Type := params.At(0).Type()
		if param1Type.String() != "context.Context" {
			return false
		}

		// Check second parameter is io.Writer
		param2Type := params.At(1).Type()
		if param2Type.String() != "io.Writer" {
			return false
		}

		// Check return type: error
		results := sig.Results()
		if results.Len() != 1 {
			return false
		}

		// Check if return type is error
		returnType := results.At(0).Type().String()
		if returnType != "error" {
			return false
		}

		return true
	}

	// Use the found interface to check implementation
	result := types.Implements(t, componentInterface)
	return result
}
