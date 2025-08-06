package testhtml

import (
	"context"
	"html/template"
	"io"
	"strings"
	"testing"

	_ "embed"

	"github.com/a-h/templ/parser/v2"
)

func BenchmarkTemplRender(b *testing.B) {
	b.ReportAllocs()
	t := Render(Person{
		Name:  "Luiz Bonfa",
		Email: "luiz@example.com",
	})

	w := new(strings.Builder)
	for range b.N {
		err := t.Render(context.Background(), w)
		if err != nil {
			b.Errorf("failed to render: %v", err)
		}
		w.Reset()
	}
}

//go:embed template.templ
var parserBenchmarkTemplate string

func BenchmarkTemplParser(b *testing.B) {
	for range b.N {
		tf, err := parser.ParseString(parserBenchmarkTemplate)
		if err != nil {
			b.Fatal(err)
		}
		if tf.Package.Expression.Value == "" {
			b.Fatal("unexpected nil template")
		}
	}
}

var goTemplate = template.Must(template.New("example").Parse(`<div>
	<h1>{{.Name}}</h1>
	<div style="font-family: &#39;sans-serif&#39;" id="test" data-contents="something with &#34;quotes&#34; and a &lt;tag&gt;">
		<div>
			email:<a href="mailto: {{.Email}}">{{.Email}}</a></div>
		</div>
	</div>
	<hr noshade>
	<hr optionA optionB optionC="other">
	<hr noshade>
`))

func BenchmarkGoTemplateRender(b *testing.B) {
	w := new(strings.Builder)
	person := Person{
		Name:  "Luiz Bonfa",
		Email: "luiz@exapmle.com",
	}
	b.ReportAllocs()
	for range b.N {
		err := goTemplate.Execute(w, person)
		if err != nil {
			b.Errorf("failed to render: %v", err)
		}
		w.Reset()
	}
}

const html = `<div><h1>Luiz Bonfa</h1><div style="font-family: &#39;sans-serif&#39;" id="test" data-contents="something with &#34;quotes&#34; and a &lt;tag&gt;"><div>email:<a href="mailto: luiz@example.com">luiz@example.com</a></div></div></div><hr noshade><hr optionA optionB optionC="other"><hr noshade>`

func BenchmarkIOWriteString(b *testing.B) {
	b.ReportAllocs()
	w := new(strings.Builder)
	for range b.N {
		_, err := io.WriteString(w, html)
		if err != nil {
			b.Errorf("failed to render: %v", err)
		}
		w.Reset()
	}
}
