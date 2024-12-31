package templ

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"html/template"
	"io"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/a-h/templ/safehtml"
)

// Types exposed by all components.

// Component is the interface that all templates implement.
type Component interface {
	// Render the template.
	Render(ctx context.Context, w io.Writer) error
}

// ComponentFunc converts a function that matches the Component interface's
// Render method into a Component.
type ComponentFunc func(ctx context.Context, w io.Writer) error

// Render the template.
func (cf ComponentFunc) Render(ctx context.Context, w io.Writer) error {
	return cf(ctx, w)
}

// WithNonce sets a CSP nonce on the context and returns it.
func WithNonce(ctx context.Context, nonce string) context.Context {
	ctx, v := getContext(ctx)
	v.nonce = nonce
	return ctx
}

// GetNonce returns the CSP nonce value set with WithNonce, or an
// empty string if none has been set.
func GetNonce(ctx context.Context) (nonce string) {
	if ctx == nil {
		return ""
	}
	_, v := getContext(ctx)
	return v.nonce
}

func WithChildren(ctx context.Context, children Component) context.Context {
	ctx, v := getContext(ctx)
	v.children = &children
	return ctx
}

func ClearChildren(ctx context.Context) context.Context {
	_, v := getContext(ctx)
	v.children = nil
	return ctx
}

// NopComponent is a component that doesn't render anything.
var NopComponent = ComponentFunc(func(ctx context.Context, w io.Writer) error { return nil })

// GetChildren from the context.
func GetChildren(ctx context.Context) Component {
	_, v := getContext(ctx)
	if v.children == nil {
		return NopComponent
	}
	return *v.children
}

// EscapeString escapes HTML text within templates.
func EscapeString(s string) string {
	return html.EscapeString(s)
}

// Bool attribute value.
func Bool(value bool) bool {
	return value
}

// Classes for CSS.
// Supported types are string, ConstantCSSClass, ComponentCSSClass, map[string]bool.
func Classes(classes ...any) CSSClasses {
	return CSSClasses(classes)
}

// CSSClasses is a slice of CSS classes.
type CSSClasses []any

// String returns the names of all CSS classes.
func (classes CSSClasses) String() string {
	if len(classes) == 0 {
		return ""
	}
	cp := newCSSProcessor()
	for _, v := range classes {
		cp.Add(v)
	}
	return cp.String()
}

func newCSSProcessor() *cssProcessor {
	return &cssProcessor{
		classNameToEnabled: make(map[string]bool),
	}
}

type cssProcessor struct {
	classNameToEnabled map[string]bool
	orderedNames       []string
}

func (cp *cssProcessor) Add(item any) {
	switch c := item.(type) {
	case []string:
		for _, className := range c {
			cp.AddClassName(className, true)
		}
	case string:
		cp.AddClassName(c, true)
	case ConstantCSSClass:
		cp.AddClassName(c.ClassName(), true)
	case ComponentCSSClass:
		cp.AddClassName(c.ClassName(), true)
	case map[string]bool:
		// In Go, map keys are iterated in a randomized order.
		// So the keys in the map must be sorted to produce consistent output.
		keys := make([]string, len(c))
		var i int
		for key := range c {
			keys[i] = key
			i++
		}
		sort.Strings(keys)
		for _, className := range keys {
			cp.AddClassName(className, c[className])
		}
	case []KeyValue[string, bool]:
		for _, kv := range c {
			cp.AddClassName(kv.Key, kv.Value)
		}
	case KeyValue[string, bool]:
		cp.AddClassName(c.Key, c.Value)
	case []KeyValue[CSSClass, bool]:
		for _, kv := range c {
			cp.AddClassName(kv.Key.ClassName(), kv.Value)
		}
	case KeyValue[CSSClass, bool]:
		cp.AddClassName(c.Key.ClassName(), c.Value)
	case CSSClasses:
		for _, item := range c {
			cp.Add(item)
		}
	case []CSSClass:
		for _, item := range c {
			cp.Add(item)
		}
	case func() CSSClass:
		cp.AddClassName(c().ClassName(), true)
	default:
		cp.AddClassName(unknownTypeClassName, true)
	}
}

func (cp *cssProcessor) AddClassName(className string, enabled bool) {
	cp.classNameToEnabled[className] = enabled
	cp.orderedNames = append(cp.orderedNames, className)
}

func (cp *cssProcessor) String() string {
	// Order the outputs according to how they were input, and remove disabled names.
	rendered := make(map[string]any, len(cp.classNameToEnabled))
	var names []string
	for _, name := range cp.orderedNames {
		if enabled := cp.classNameToEnabled[name]; !enabled {
			continue
		}
		if _, hasBeenRendered := rendered[name]; hasBeenRendered {
			continue
		}
		names = append(names, name)
		rendered[name] = struct{}{}
	}

	return strings.Join(names, " ")
}

// KeyValue is a key and value pair.
type KeyValue[TKey comparable, TValue any] struct {
	Key   TKey   `json:"name"`
	Value TValue `json:"value"`
}

// KV creates a new key/value pair from the input key and value.
func KV[TKey comparable, TValue any](key TKey, value TValue) KeyValue[TKey, TValue] {
	return KeyValue[TKey, TValue]{
		Key:   key,
		Value: value,
	}
}

const unknownTypeClassName = "--templ-css-class-unknown-type"

// Class returns a CSS class name.
// Deprecated: use a string instead.
func Class(name string) CSSClass {
	return SafeClass(name)
}

// SafeClass bypasses CSS class name validation.
// Deprecated: use a string instead.
func SafeClass(name string) CSSClass {
	return ConstantCSSClass(name)
}

// CSSClass provides a class name.
type CSSClass interface {
	ClassName() string
}

// ConstantCSSClass is a string constant of a CSS class name.
// Deprecated: use a string instead.
type ConstantCSSClass string

// ClassName of the CSS class.
func (css ConstantCSSClass) ClassName() string {
	return string(css)
}

// ComponentCSSClass is a templ.CSS
type ComponentCSSClass struct {
	// ID of the class, will be autogenerated.
	ID string
	// Definition of the CSS.
	Class SafeCSS
}

// ClassName of the CSS class.
func (css ComponentCSSClass) ClassName() string {
	return css.ID
}

// CSSID calculates an ID.
func CSSID(name string, css string) string {
	sum := sha256.Sum256([]byte(css))
	hs := hex.EncodeToString(sum[:])[0:8] // NOTE: See issue #978. Minimum recommended hs length is 6.
	// Benchmarking showed this was fastest, and with fewest allocations (1).
	// Using strings.Builder (2 allocs).
	// Using fmt.Sprintf (3 allocs).
	return name + "_" + hs
}

// NewCSSMiddleware creates HTTP middleware that renders a global stylesheet of ComponentCSSClass
// CSS if the request path matches, or updates the HTTP context to ensure that any handlers that
// use templ.Components skip rendering <style> elements for classes that are included in the global
// stylesheet. By default, the stylesheet path is /styles/templ.css
func NewCSSMiddleware(next http.Handler, classes ...CSSClass) CSSMiddleware {
	return CSSMiddleware{
		Path:       "/styles/templ.css",
		CSSHandler: NewCSSHandler(classes...),
		Next:       next,
	}
}

// CSSMiddleware renders a global stylesheet.
type CSSMiddleware struct {
	Path       string
	CSSHandler CSSHandler
	Next       http.Handler
}

func (cssm CSSMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == cssm.Path {
		cssm.CSSHandler.ServeHTTP(w, r)
		return
	}
	// Add registered classes to the context.
	ctx, v := getContext(r.Context())
	for _, c := range cssm.CSSHandler.Classes {
		v.addClass(c.ID)
	}
	// Serve the request. Templ components will use the updated context
	// to know to skip rendering <style> elements for any component CSS
	// classes that have been included in the global stylesheet.
	cssm.Next.ServeHTTP(w, r.WithContext(ctx))
}

// NewCSSHandler creates a handler that serves a stylesheet containing the CSS of the
// classes passed in. This is used by the CSSMiddleware to provide global stylesheets
// for templ components.
func NewCSSHandler(classes ...CSSClass) CSSHandler {
	ccssc := make([]ComponentCSSClass, 0, len(classes))
	for _, c := range classes {
		ccss, ok := c.(ComponentCSSClass)
		if !ok {
			continue
		}
		ccssc = append(ccssc, ccss)
	}
	return CSSHandler{
		Classes: ccssc,
	}
}

// CSSHandler is a HTTP handler that serves CSS.
type CSSHandler struct {
	Logger  func(err error)
	Classes []ComponentCSSClass
}

func (cssh CSSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	for _, c := range cssh.Classes {
		_, err := w.Write([]byte(c.Class))
		if err != nil && cssh.Logger != nil {
			cssh.Logger(err)
		}
	}
}

// RenderCSSItems renders the CSS to the writer, if the items haven't already been rendered.
func RenderCSSItems(ctx context.Context, w io.Writer, classes ...any) (err error) {
	if len(classes) == 0 {
		return nil
	}
	_, v := getContext(ctx)
	sb := new(strings.Builder)
	renderCSSItemsToBuilder(sb, v, classes...)
	if sb.Len() > 0 {
		if _, err = io.WriteString(w, `<style type="text/css">`); err != nil {
			return err
		}
		if _, err = io.WriteString(w, sb.String()); err != nil {
			return err
		}
		if _, err = io.WriteString(w, `</style>`); err != nil {
			return err
		}
	}
	return nil
}

func renderCSSItemsToBuilder(sb *strings.Builder, v *contextValue, classes ...any) {
	for _, c := range classes {
		switch ccc := c.(type) {
		case ComponentCSSClass:
			if !v.hasClassBeenRendered(ccc.ID) {
				sb.WriteString(string(ccc.Class))
				v.addClass(ccc.ID)
			}
		case KeyValue[ComponentCSSClass, bool]:
			if !ccc.Value {
				continue
			}
			renderCSSItemsToBuilder(sb, v, ccc.Key)
		case KeyValue[CSSClass, bool]:
			if !ccc.Value {
				continue
			}
			renderCSSItemsToBuilder(sb, v, ccc.Key)
		case CSSClasses:
			renderCSSItemsToBuilder(sb, v, ccc...)
		case []CSSClass:
			for _, item := range ccc {
				renderCSSItemsToBuilder(sb, v, item)
			}
		case func() CSSClass:
			renderCSSItemsToBuilder(sb, v, ccc())
		case []string:
			// Skip. These are class names, not CSS classes.
		case string:
			// Skip. This is a class name, not a CSS class.
		case ConstantCSSClass:
			// Skip. This is a class name, not a CSS class.
		case CSSClass:
			// Skip. This is a class name, not a CSS class.
		case map[string]bool:
			// Skip. These are class names, not CSS classes.
		case KeyValue[string, bool]:
			// Skip. These are class names, not CSS classes.
		case []KeyValue[string, bool]:
			// Skip. These are class names, not CSS classes.
		case KeyValue[ConstantCSSClass, bool]:
			// Skip. These are class names, not CSS classes.
		case []KeyValue[ConstantCSSClass, bool]:
			// Skip. These are class names, not CSS classes.
		}
	}
}

// SafeCSS is CSS that has been sanitized.
type SafeCSS string

type SafeCSSProperty string

var safeCSSPropertyType = reflect.TypeOf(SafeCSSProperty(""))

// SanitizeCSS sanitizes CSS properties to ensure that they are safe.
func SanitizeCSS[T ~string](property string, value T) SafeCSS {
	if reflect.TypeOf(value) == safeCSSPropertyType {
		return SafeCSS(safehtml.SanitizeCSSProperty(property) + ":" + string(value) + ";")
	}
	p, v := safehtml.SanitizeCSS(property, string(value))
	return SafeCSS(p + ":" + v + ";")
}

// Attributes is an alias to map[string]any made for spread attributes.
type Attributes map[string]any

// sortedKeys returns the keys of a map in sorted order.
func sortedKeys(m map[string]any) (keys []string) {
	keys = make([]string, len(m))
	var i int
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

func writeStrings(w io.Writer, ss ...string) (err error) {
	for _, s := range ss {
		if _, err = io.WriteString(w, s); err != nil {
			return err
		}
	}
	return nil
}

func RenderAttributes(ctx context.Context, w io.Writer, attributes Attributes) (err error) {
	for _, key := range sortedKeys(attributes) {
		value := attributes[key]
		switch value := value.(type) {
		case string:
			if err = writeStrings(w, ` `, EscapeString(key), `="`, EscapeString(value), `"`); err != nil {
				return err
			}
		case *string:
			if value != nil {
				if err = writeStrings(w, ` `, EscapeString(key), `="`, EscapeString(*value), `"`); err != nil {
					return err
				}
			}
		case bool:
			if value {
				if err = writeStrings(w, ` `, EscapeString(key)); err != nil {
					return err
				}
			}
		case *bool:
			if value != nil && *value {
				if err = writeStrings(w, ` `, EscapeString(key)); err != nil {
					return err
				}
			}
		case KeyValue[string, bool]:
			if value.Value {
				if err = writeStrings(w, ` `, EscapeString(key), `="`, EscapeString(value.Key), `"`); err != nil {
					return err
				}
			}
		case KeyValue[bool, bool]:
			if value.Value && value.Key {
				if err = writeStrings(w, ` `, EscapeString(key)); err != nil {
					return err
				}
			}
		case func() bool:
			if value() {
				if err = writeStrings(w, ` `, EscapeString(key)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Context.

type contextKeyType int

const contextKey = contextKeyType(0)

type contextValue struct {
	ss          map[string]struct{}
	onceHandles map[*OnceHandle]struct{}
	children    *Component
	nonce       string
}

func (v *contextValue) setHasBeenRendered(h *OnceHandle) {
	if v.onceHandles == nil {
		v.onceHandles = map[*OnceHandle]struct{}{}
	}
	v.onceHandles[h] = struct{}{}
}

func (v *contextValue) getHasBeenRendered(h *OnceHandle) (ok bool) {
	if v.onceHandles == nil {
		v.onceHandles = map[*OnceHandle]struct{}{}
	}
	_, ok = v.onceHandles[h]
	return
}

func (v *contextValue) addScript(s string) {
	if v.ss == nil {
		v.ss = map[string]struct{}{}
	}
	v.ss["script_"+s] = struct{}{}
}

func (v *contextValue) hasScriptBeenRendered(s string) (ok bool) {
	if v.ss == nil {
		v.ss = map[string]struct{}{}
	}
	_, ok = v.ss["script_"+s]
	return
}

func (v *contextValue) addClass(s string) {
	if v.ss == nil {
		v.ss = map[string]struct{}{}
	}
	v.ss["class_"+s] = struct{}{}
}

func (v *contextValue) hasClassBeenRendered(s string) (ok bool) {
	if v.ss == nil {
		v.ss = map[string]struct{}{}
	}
	_, ok = v.ss["class_"+s]
	return
}

// InitializeContext initializes context used to store internal state used during rendering.
func InitializeContext(ctx context.Context) context.Context {
	if _, ok := ctx.Value(contextKey).(*contextValue); ok {
		return ctx
	}
	v := &contextValue{}
	ctx = context.WithValue(ctx, contextKey, v)
	return ctx
}

func getContext(ctx context.Context) (context.Context, *contextValue) {
	v, ok := ctx.Value(contextKey).(*contextValue)
	if !ok {
		ctx = InitializeContext(ctx)
		v = ctx.Value(contextKey).(*contextValue)
	}
	return ctx, v
}

var bufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func GetBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

func ReleaseBuffer(b *bytes.Buffer) {
	b.Reset()
	bufferPool.Put(b)
}

// JoinStringErrs joins an optional list of errors.
func JoinStringErrs(s string, errs ...error) (string, error) {
	return s, errors.Join(errs...)
}

// Error returned during template rendering.
type Error struct {
	Err error
	// FileName of the template file.
	FileName string
	// Line index of the error.
	Line int
	// Col index of the error.
	Col int
}

func (e Error) Error() string {
	if e.FileName == "" {
		e.FileName = "templ"
	}
	return fmt.Sprintf("%s: error at line %d, col %d: %v", e.FileName, e.Line, e.Col, e.Err)
}

func (e Error) Unwrap() error {
	return e.Err
}

// Raw renders the input HTML to the output without applying HTML escaping.
//
// Use of this component presents a security risk - the HTML should come from
// a trusted source, because it will be included as-is in the output.
func Raw[T ~string](html T, errs ...error) Component {
	return ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		if err = errors.Join(errs...); err != nil {
			return err
		}
		_, err = io.WriteString(w, string(html))
		return err
	})
}

// FromGoHTML creates a templ Component from a Go html/template template.
func FromGoHTML(t *template.Template, data any) Component {
	return ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		return t.Execute(w, data)
	})
}

// ToGoHTML renders the component to a Go html/template template.HTML string.
func ToGoHTML(ctx context.Context, c Component) (s template.HTML, err error) {
	b := GetBuffer()
	defer ReleaseBuffer(b)
	if err = c.Render(ctx, b); err != nil {
		return
	}
	s = template.HTML(b.String())
	return
}
