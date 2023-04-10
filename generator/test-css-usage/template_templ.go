// Code generated by templ@(devel) DO NOT EDIT.

package testcssusage

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"
import "strings"

func green() templ.CSSClass {
	var templCSSBuilder strings.Builder
	templCSSBuilder.WriteString(`color:#00ff00;`)
	templCSSID := templ.CSSID(`green`, templCSSBuilder.String())
	return templ.ComponentCSSClass{
		ID: templCSSID,
		Class: templ.SafeCSS(`.` + templCSSID + `{` + templCSSBuilder.String() + `}`),
	}
}

func className() templ.CSSClass {
	var templCSSBuilder strings.Builder
	templCSSBuilder.WriteString(`background-color:#ffffff;`)
	templCSSBuilder.WriteString(string(templ.SanitizeCSS(`color`, red)))
	templCSSID := templ.CSSID(`className`, templCSSBuilder.String())
	return templ.ComponentCSSClass{
		ID: templCSSID,
		Class: templ.SafeCSS(`.` + templCSSID + `{` + templCSSBuilder.String() + `}`),
	}
}

func Button(text string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		templBuffer, templIsBuffer := w.(*bytes.Buffer)
		if !templIsBuffer {
			templBuffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templBuffer)
		}
		ctx = templ.InitializeContext(ctx)
		var_1 := templ.GetChildren(ctx)
		if var_1 == nil {
			var_1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		// Element (standard)
		// Element CSS
		var var_2 = []any{className(), templ.Class("&&&unsafe"), "safe", templ.SafeClass("safe2")}
		err = templ.RenderCSSItems(ctx, templBuffer, var_2...)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("<button")
		if err != nil {
			return err
		}
		// Element Attributes
		_, err = templBuffer.WriteString(" class=")
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("\"")
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString(templ.EscapeString(templ.CSSClasses(var_2).String()))
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("\"")
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString(" type=\"button\"")
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString(">")
		if err != nil {
			return err
		}
		// StringExpression
		var var_3 string = text
		_, err = templBuffer.WriteString(templ.EscapeString(var_3))
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</button>")
		if err != nil {
			return err
		}
		if !templIsBuffer {
			_, err = io.Copy(w, templBuffer)
		}
		return err
	})
}

func LegacySupport() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		templBuffer, templIsBuffer := w.(*bytes.Buffer)
		if !templIsBuffer {
			templBuffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templBuffer)
		}
		ctx = templ.InitializeContext(ctx)
		var_4 := templ.GetChildren(ctx)
		if var_4 == nil {
			var_4 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		// Element (standard)
		// Element CSS
		var var_5 = []any{templ.Classes(templ.Class("test"), "a")}
		err = templ.RenderCSSItems(ctx, templBuffer, var_5...)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("<div")
		if err != nil {
			return err
		}
		// Element Attributes
		_, err = templBuffer.WriteString(" class=")
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("\"")
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString(templ.EscapeString(templ.CSSClasses(var_5).String()))
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("\"")
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString(">")
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</div>")
		if err != nil {
			return err
		}
		if !templIsBuffer {
			_, err = io.Copy(w, templBuffer)
		}
		return err
	})
}

func MapCSSExample() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		templBuffer, templIsBuffer := w.(*bytes.Buffer)
		if !templIsBuffer {
			templBuffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templBuffer)
		}
		ctx = templ.InitializeContext(ctx)
		var_6 := templ.GetChildren(ctx)
		if var_6 == nil {
			var_6 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		// Element (standard)
		// Element CSS
		var var_7 = []any{map[string]bool{ "a": true, "b": false, "c": true }}
		err = templ.RenderCSSItems(ctx, templBuffer, var_7...)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("<div")
		if err != nil {
			return err
		}
		// Element Attributes
		_, err = templBuffer.WriteString(" class=")
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("\"")
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString(templ.EscapeString(templ.CSSClasses(var_7).String()))
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("\"")
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString(">")
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</div>")
		if err != nil {
			return err
		}
		if !templIsBuffer {
			_, err = io.Copy(w, templBuffer)
		}
		return err
	})
}

func KVExample() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		templBuffer, templIsBuffer := w.(*bytes.Buffer)
		if !templIsBuffer {
			templBuffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templBuffer)
		}
		ctx = templ.InitializeContext(ctx)
		var_8 := templ.GetChildren(ctx)
		if var_8 == nil {
			var_8 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		// Element (standard)
		// Element CSS
		var var_9 = []any{"a", templ.KV("b", false)}
		err = templ.RenderCSSItems(ctx, templBuffer, var_9...)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("<div")
		if err != nil {
			return err
		}
		// Element Attributes
		_, err = templBuffer.WriteString(" class=")
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("\"")
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString(templ.EscapeString(templ.CSSClasses(var_9).String()))
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("\"")
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString(">")
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</div>")
		if err != nil {
			return err
		}
		if !templIsBuffer {
			_, err = io.Copy(w, templBuffer)
		}
		return err
	})
}

func ThreeButtons() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		templBuffer, templIsBuffer := w.(*bytes.Buffer)
		if !templIsBuffer {
			templBuffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templBuffer)
		}
		ctx = templ.InitializeContext(ctx)
		var_10 := templ.GetChildren(ctx)
		if var_10 == nil {
			var_10 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		// CallTemplate
		err = Button("A").Render(ctx, templBuffer)
		if err != nil {
			return err
		}
		// CallTemplate
		err = Button("B").Render(ctx, templBuffer)
		if err != nil {
			return err
		}
		// Element (standard)
		// Element CSS
		var var_11 = []any{templ.Classes(green)}
		err = templ.RenderCSSItems(ctx, templBuffer, var_11...)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("<button")
		if err != nil {
			return err
		}
		// Element Attributes
		_, err = templBuffer.WriteString(" class=")
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("\"")
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString(templ.EscapeString(templ.CSSClasses(var_11).String()))
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("\"")
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString(" type=\"button\"")
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString(">")
		if err != nil {
			return err
		}
		// StringExpression
		var var_12 string = "Green"
		_, err = templBuffer.WriteString(templ.EscapeString(var_12))
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</button>")
		if err != nil {
			return err
		}
		// CallTemplate
		err = MapCSSExample().Render(ctx, templBuffer)
		if err != nil {
			return err
		}
		// CallTemplate
		err = KVExample().Render(ctx, templBuffer)
		if err != nil {
			return err
		}
		if !templIsBuffer {
			_, err = io.Copy(w, templBuffer)
		}
		return err
	})
}

