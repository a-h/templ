package generator

import (
	"fmt"
	"go/types"
	"strings"

	"github.com/a-h/templ/parser/v2"
	"golang.org/x/tools/go/packages"
)

// ComponentSignature represents a templ component's function signature or struct fields
type ComponentSignature struct {
	PackagePath   string
	Name          string
	QualifiedName string          // For functions: pkgPath.Name, For structs: pkgPath.TypeName
	Parameters    []ParameterInfo // For functions: parameters, For structs: exported fields
	IsStruct      bool
	IsPointerRecv bool
}

// ParameterInfo represents a function parameter or struct field
type ParameterInfo struct {
	Name string
	Type string
}

// ComponentSignatures represents a collection of component signatures
type ComponentSignatures map[string]ComponentSignature

// Get returns the signature for a component name
func (cs ComponentSignatures) Get(qname string) (ComponentSignature, bool) {
	sig, ok := cs[qname]
	return sig, ok
}

// Set adds a signature to the collection
func (cs ComponentSignatures) Add(sig ComponentSignature) {
	cs[sig.QualifiedName] = sig
}

// ComponentResolutionError represents an error during component resolution with position information
type ComponentResolutionError struct {
	Err      error
	Position parser.Position
	FileName string
}

func (e ComponentResolutionError) Error() string {
	if e.FileName == "" {
		return e.Err.Error()
	}
	return fmt.Sprintf("%s:%d:%d: %v", e.FileName, e.Position.Line, e.Position.Col, e.Err)
}

// SymbolResolver resolves component symbols across packages
type SymbolResolver struct {
	cache      map[string]ComponentSignature
	workingDir string
}

// NewSymbolResolver creates a new symbol resolver
func NewSymbolResolver(workingDir string) *SymbolResolver {
	return &SymbolResolver{
		cache:      make(map[string]ComponentSignature),
		workingDir: workingDir,
	}
}

// ResolveComponent resolves a component's function signature
func (sr *SymbolResolver) ResolveComponent(pkgPath, componentName string) (ComponentSignature, error) {
	return sr.ResolveComponentWithPosition(pkgPath, componentName, parser.Position{}, "")
}

// ResolveComponentWithPosition resolves a component's function signature with position information for error reporting
func (sr *SymbolResolver) ResolveComponentWithPosition(pkgPath, componentName string, pos parser.Position, fileName string) (ComponentSignature, error) {
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
		baseErr := fmt.Errorf("failed to load package %s: %w", pkgPath, err)
		if fileName != "" {
			return ComponentSignature{}, ComponentResolutionError{Err: baseErr, Position: pos, FileName: fileName}
		}
		return ComponentSignature{}, baseErr
	}

	if len(pkgs) == 0 {
		baseErr := fmt.Errorf("package %s not found", pkgPath)
		if fileName != "" {
			return ComponentSignature{}, ComponentResolutionError{Err: baseErr, Position: pos, FileName: fileName}
		}
		return ComponentSignature{}, baseErr
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
			baseErr := fmt.Errorf("package %s has non-generated file errors: %v", pkgPath, pkg.Errors)
			if fileName != "" {
				return ComponentSignature{}, ComponentResolutionError{Err: baseErr, Position: pos, FileName: fileName}
			}
			return ComponentSignature{}, baseErr
		}
		// Continue with type checking even if there are _templ.go errors
	}

	// Look for the component function
	obj := pkg.Types.Scope().Lookup(componentName)
	if obj == nil {
		baseErr := fmt.Errorf("component %s not found in package %s", componentName, pkgPath)
		if fileName != "" {
			return ComponentSignature{}, ComponentResolutionError{Err: baseErr, Position: pos, FileName: fileName}
		}
		return ComponentSignature{}, baseErr
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
			baseErr := fmt.Errorf("%s is neither a function nor a type", componentName)
			if fileName != "" {
				return ComponentSignature{}, ComponentResolutionError{Err: baseErr, Position: pos, FileName: fileName}
			}
			return ComponentSignature{}, baseErr
		}
		isStruct, isPointerRecv = sr.implementsComponent(typeName.Type(), pkg.Types)
		if !isStruct {
			baseErr := fmt.Errorf("%s does not implement templ.Component interface", componentName)
			if fileName != "" {
				return ComponentSignature{}, ComponentResolutionError{Err: baseErr, Position: pos, FileName: fileName}
			}
			return ComponentSignature{}, baseErr
		}
		// Extract struct fields for struct components
		paramInfo = sr.extractStructFields(typeName.Type())
	}

	componentSig := ComponentSignature{
		PackagePath:   pkgPath,
		Name:          componentName,
		Parameters:    paramInfo,
		IsStruct:      isStruct,
		IsPointerRecv: isPointerRecv,
	}

	sr.cache[key] = componentSig
	return componentSig, nil
}

// ResolveLocalComponent resolves a component in the current package with position information
func (sr *SymbolResolver) ResolveLocalComponent(componentName string, pos parser.Position, fileName string) (ComponentSignature, error) {
	// Use packages.Load to get the correct package path in module mode
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedModule,
		Dir:  sr.workingDir,
	}

	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		baseErr := fmt.Errorf("failed to load current package: %w", err)
		if fileName != "" {
			return ComponentSignature{}, ComponentResolutionError{Err: baseErr, Position: pos, FileName: fileName}
		}
		return ComponentSignature{}, baseErr
	}

	if len(pkgs) == 0 {
		baseErr := fmt.Errorf("no package found in current directory")
		if fileName != "" {
			return ComponentSignature{}, ComponentResolutionError{Err: baseErr, Position: pos, FileName: fileName}
		}
		return ComponentSignature{}, baseErr
	}

	pkgPath := pkgs[0].PkgPath

	return sr.ResolveComponentWithPosition(pkgPath, componentName, pos, fileName)
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

func (sr *SymbolResolver) extractStructFields(t types.Type) []ParameterInfo {
	var structType *types.Struct
	switch underlying := t.Underlying().(type) {
	case *types.Struct:
		structType = underlying
	default:
		return nil
	}

	var fields []ParameterInfo
	for i := range structType.NumFields() {
		field := structType.Field(i)
		fields = append(fields, ParameterInfo{
			Name: field.Name(),
			Type: field.Type().String(),
		})
	}

	return fields
}
