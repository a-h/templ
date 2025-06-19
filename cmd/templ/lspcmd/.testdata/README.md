# LSP Test Data Files

This directory contains test data files for LSP (Language Server Protocol) features.

## Annotated Test Files

Files prefixed with `annotated_` use a special format to specify expected LSP behavior:

### Format

```
// @expect-diagnostic: Expected diagnostic message
<ComponentWithIssue />
```

### Supported Annotations

- `@expect-diagnostic: message` - Expects a diagnostic with the specified message

### Example

```templ
package test

templ ValidComponent() {
    <div>Valid</div>
}

templ TestFile() {
    <ValidComponent />  // No diagnostic expected
    
    // @expect-diagnostic: Component MissingComponent not found
    <MissingComponent />
}
```

### Test Execution

Run tests with:
```bash
go test -v -run TestAnnotatedMessages
```

The test will automatically discover all `.templ` files in the `.testdata` directory and run tests on files that contain `@expect-diagnostic` annotations.

### Creating New Test Files

1. Create any `.templ` file in the `.testdata` directory
2. Add your templ code
3. Add `// @expect-diagnostic: message` comments on the line before any component that should generate a diagnostic
4. The test framework will automatically:
   - Parse your annotations
   - Run diagnostics on the template
   - Compare expected vs actual diagnostic messages

### Features Tested

- **Missing Components**: Components that are referenced but not defined
- **Package Qualified Components**: Components like `pkg.Component` (ignored)
- **Struct Method Components**: Components like `variable.Method` (ignored)
- **Nested Components**: Components inside HTML elements or other templates

### File Types

- `annotated_simple.templ` - Basic test cases for missing components
- `annotated_comprehensive.templ` - Complex scenarios with nesting and various component types
- `annotated_diagnostics.templ` - Tests basic missing component scenarios
- Other `.templ` files without annotations will be skipped by the test runner

## Future Extensions

The annotation format can be extended to support other LSP features:

- `@expect-hover: hover_text` - Expected hover information
- `@expect-completion: item1,item2` - Expected completion items
- `@expect-definition: file:line` - Expected go-to-definition target