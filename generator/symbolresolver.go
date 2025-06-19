package generator

import (
	"fmt"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

// ComponentSignature represents a templ component's function signature
type ComponentSignature struct {
	PackagePath   string
	Name          string
	Parameters    []ParameterInfo
	IsStruct      bool
	IsPointerRecv bool
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
	var isStruct, isPointerRecv bool

	// The component can be either a function or a type that implements templ.Component
	if fn, ok := obj.(*types.Func); ok {
		sig := fn.Type().(*types.Signature)

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
		typeName, ok := obj.(*types.TypeName)
		if !ok {
			return nil, fmt.Errorf("%s is neither a function nor a type", componentName)
		}
		isStruct, isPointerRecv = sr.implementsComponent(typeName.Type(), pkg.Types)
		if !isStruct {
			return nil, fmt.Errorf("%s does not implement templ.Component interface", componentName)
		}
		// TODO: Handle parameters for struct components
		paramInfo = make([]ParameterInfo, 0)
	}

	componentSig := &ComponentSignature{
		PackagePath:   pkgPath,
		Name:          componentName,
		Parameters:    paramInfo,
		IsStruct:      isStruct,
		IsPointerRecv: isPointerRecv,
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

// implementsComponent checks if a type implements the templ.Component interface: Render(ctx context.Context, w io.Writer) error
// Returns (implements, isPointerReceiver)
func (sr *SymbolResolver) implementsComponent(t types.Type, pkg *types.Package) (bool, bool) {
	method, _, _ := types.LookupFieldOrMethod(t, true, pkg, "Render")
	if method == nil {
		return false, false
	}

	fn, ok := method.(*types.Func)
	if !ok {
		return false, false
	}

	sig := fn.Type().(*types.Signature)

	// Check parameters: (context.Context, io.Writer)
	params := sig.Params()
	if params.Len() != 2 {
		return false, false
	}
	if params.At(0).Type().String() != "context.Context" {
		return false, false
	}
	if params.At(1).Type().String() != "io.Writer" {
		return false, false
	}

	// Check return type: error
	results := sig.Results()
	if results.Len() != 1 {
		return false, false
	}
	returnType := results.At(0).Type().String()
	if returnType != "error" {
		return false, false
	}

	// Check if the receiver is a pointer by examining the method signature
	isPointerReceiver := false
	if sig.Recv() == nil {
		// We should never reach here since the Render should always have a receiver
		panic("Method signature has no receiver")
	}
	recvType := sig.Recv().Type()
	_, isPointerReceiver = recvType.(*types.Pointer)

	return true, isPointerReceiver
}
