// Code generated by templ@(devel) DO NOT EDIT.

package testswitch

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

func render(input string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		templBuffer, templIsBuffer := w.(*bytes.Buffer)
		if !templIsBuffer {
			templBuffer = new(bytes.Buffer)
		}
		ctx = templ.InitializeContext(ctx)
		var_1 := templ.GetChildren(ctx)
		if var_1 == nil {
			var_1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		// Switch
		switch input {
		case "a":
			// StringExpression
			_, err = templBuffer.WriteString(templ.EscapeString("it was 'a'"))
			if err != nil {
				return err
			}
		default:
			// StringExpression
			_, err = templBuffer.WriteString(templ.EscapeString("it was something else"))
			if err != nil {
				return err
			}
		}
		if !templIsBuffer {
			_, err = io.Copy(w, templBuffer)
		}
		return err
	})
}

