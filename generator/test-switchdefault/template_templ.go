// Code generated by templ@(devel) DO NOT EDIT.

package testswitchdefault

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

func template(input string) templ.Component {
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
		// Switch
		switch input {
		case "a":			// StringExpression
			var var_2 string = "it was 'a'"
			_, err = templBuffer.WriteString(templ.EscapeString(var_2))
			if err != nil {
				return err
			}
		default:			// StringExpression
			var var_3 string = "it was something else"
			_, err = templBuffer.WriteString(templ.EscapeString(var_3))
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

