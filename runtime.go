package templ

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
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

// ComponentHandler is a http.Handler that renders components.
type ComponentHandler struct {
	Component    Component
	Status       int
	ContentType  string
	ErrorHandler func(r *http.Request, err error) http.Handler
}

const componentHandlerErrorMessage = "templ: failed to render template"

// ServeHTTP implements the http.Handler interface.
func (ch ComponentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if ch.Status != 0 {
		w.WriteHeader(ch.Status)
	}
	w.Header().Add("Content-Type", ch.ContentType)
	err := ch.Component.Render(r.Context(), w)
	if err != nil {
		if ch.ErrorHandler != nil {
			ch.ErrorHandler(r, err).ServeHTTP(w, r)
			return
		}
		http.Error(w, componentHandlerErrorMessage, http.StatusInternalServerError)
	}
}

// Handler creates a http.Handler that renders the template.
func Handler(c Component, options ...func(*ComponentHandler)) *ComponentHandler {
	ch := &ComponentHandler{
		Component:   c,
		ContentType: "text/html",
	}
	for _, o := range options {
		o(ch)
	}
	return ch
}

// WithStatus sets the HTTP status code returned by the ComponentHandler.
func WithStatus(status int) func(*ComponentHandler) {
	return func(ch *ComponentHandler) {
		ch.Status = status
	}
}

// WithConentType sets the Content-Type header returned by the ComponentHandler.
func WithContentType(contentType string) func(*ComponentHandler) {
	return func(ch *ComponentHandler) {
		ch.ContentType = contentType
	}
}

// WithErrorHandler sets the error handler used if rendering fails.
func WithErrorHandler(eh func(r *http.Request, err error) http.Handler) func(*ComponentHandler) {
	return func(ch *ComponentHandler) {
		ch.ErrorHandler = eh
	}
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
	hp := hex.EncodeToString(sum[:])[0:4]
	return fmt.Sprintf("%s_%s", name, hp)
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

// SanitizeCSS sanitizes CSS properties to ensure that they are safe.
func SanitizeCSS(property, value string) SafeCSS {
	p, v := safehtml.SanitizeCSS(property, value)
	return SafeCSS(p + ":" + v + ";")
}

// Hyperlink sanitization.

// FailedSanitizationURL is returned if a URL fails sanitization checks.
const FailedSanitizationURL = SafeURL("about:invalid#TemplFailedSanitizationURL")

// URL sanitizes the input string s and returns a SafeURL.
func URL(s string) SafeURL {
	if i := strings.IndexRune(s, ':'); i >= 0 && !strings.ContainsRune(s[:i], '/') {
		protocol := s[:i]
		if !strings.EqualFold(protocol, "http") && !strings.EqualFold(protocol, "https") && !strings.EqualFold(protocol, "mailto") {
			return FailedSanitizationURL
		}
	}
	return SafeURL(s)
}

// SafeURL is a URL that has been sanitized.
type SafeURL string

// Script handling.

func safeEncodeScriptParams(escapeHTML bool, params []any) []string {
	encodedParams := make([]string, len(params))
	for i := 0; i < len(encodedParams); i++ {
		enc, _ := json.Marshal(params[i])
		if !escapeHTML {
			encodedParams[i] = string(enc)
			continue
		}
		encodedParams[i] = EscapeString(string(enc))
	}
	return encodedParams
}

// SafeScript encodes unknown parameters for safety for inside HTML attributes.
func SafeScript(functionName string, params ...any) string {
	encodedParams := safeEncodeScriptParams(true, params)
	sb := new(strings.Builder)
	sb.WriteString(functionName)
	sb.WriteRune('(')
	sb.WriteString(strings.Join(encodedParams, ","))
	sb.WriteRune(')')
	return sb.String()
}

// SafeScript encodes unknown parameters for safety for inline scripts.
func SafeScriptInline(functionName string, params ...any) string {
	encodedParams := safeEncodeScriptParams(false, params)
	sb := new(strings.Builder)
	sb.WriteString(functionName)
	sb.WriteRune('(')
	sb.WriteString(strings.Join(encodedParams, ","))
	sb.WriteRune(')')
	return sb.String()
}

type contextKeyType int

const contextKey = contextKeyType(0)

type contextValue struct {
	ss       map[string]struct{}
	children *Component
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

// ComponentScript is a templ Script template.
type ComponentScript struct {
	// Name of the script, e.g. print.
	Name string
	// Function to render.
	Function string
	// Call of the function in JavaScript syntax, including parameters, and
	// ensures parameters are HTML escaped; useful for injecting into HTML
	// attributes like onclick, onhover, etc.
	//
	// Given:
	//    functionName("some string",12345)
	// It would render:
	//    __templ_functionName_sha(&#34;some string&#34;,12345))
	//
	// This is can be injected into HTML attributes:
	//    <button onClick="__templ_functionName_sha(&#34;some string&#34;,12345))">Click Me</button>
	Call string
	// Call of the function in JavaScript syntax, including parameters. It
	// does not HTML escape parameters; useful for directly calling in script
	// elements.
	//
	// Given:
	//    functionName("some string",12345)
	// It would render:
	//    __templ_functionName_sha("some string",12345))
	//
	// This is can be used to call the function inside a script tag:
	//    <script>__templ_functionName_sha("some string",12345))</script>
	CallInline string
}

var _ Component = ComponentScript{}

func (c ComponentScript) Render(ctx context.Context, w io.Writer) error {
	err := RenderScriptItems(ctx, w, c)
	if err != nil {
		return err
	}
	if len(c.Call) > 0 {
		if _, err = io.WriteString(w, `<script type="text/javascript">`); err != nil {
			return err
		}
		if _, err = io.WriteString(w, c.CallInline); err != nil {
			return err
		}
		if _, err = io.WriteString(w, `</script>`); err != nil {
			return err
		}
	}
	return nil
}

// RenderScriptItems renders a <script> element, if the script has not already been rendered.
func RenderScriptItems(ctx context.Context, w io.Writer, scripts ...ComponentScript) (err error) {
	if len(scripts) == 0 {
		return nil
	}
	_, v := getContext(ctx)
	sb := new(strings.Builder)
	for _, s := range scripts {
		if !v.hasScriptBeenRendered(s.Name) {
			sb.WriteString(s.Function)
			v.addScript(s.Name)
		}
	}
	if sb.Len() > 0 {
		if _, err = io.WriteString(w, `<script type="text/javascript">`); err != nil {
			return err
		}
		if _, err = io.WriteString(w, sb.String()); err != nil {
			return err
		}
		if _, err = io.WriteString(w, `</script>`); err != nil {
			return err
		}
	}
	return nil
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
