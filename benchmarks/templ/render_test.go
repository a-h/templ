package testhtml

import (
	"context"
	"io"
	"strings"
	"testing"
)

func BenchmarkCurrent(b *testing.B) {
	b.ReportAllocs()
	t := Render(Person{
		Name:  "Luiz Bonfa",
		Email: "luiz@example.com",
	})

	w := new(strings.Builder)
	for i := 0; i < b.N; i++ {
		err := t.Render(context.Background(), w)
		if err != nil {
			b.Errorf("failed to render: %v", err)
		}
		w.Reset()
	}
}

var html = `<div><h1>Luiz Bonfa</h1><div style="font-family: &#39;sans-serif&#39;" id="test" data-contents="something with &#34;quotes&#34; and a &lt;tag&gt;"><div>email:<a href="mailto: luiz@example.com">luiz@example.com</a></div></div></div><hr noshade><hr optionA optionB optionC="other"><hr noshade>`

func BenchmarkIOWriteString(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := new(strings.Builder)
		_, err := io.WriteString(w, html)
		if err != nil {
			b.Errorf("failed to render: %v", err)
		}
	}
}
