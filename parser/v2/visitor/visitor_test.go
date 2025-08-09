package visitor_test

import (
	"bytes"
	"testing"

	"github.com/a-h/templ/parser/v2"
	"github.com/a-h/templ/parser/v2/visitor"
	"github.com/google/go-cmp/cmp"
)

func assetRewriter(rewrite func(path string) string) parser.Visitor {
	ar := visitor.New()
	inSrcElement := false
	inHrefElement := false

	// Save the original Element visitor to allow chaining.
	visitElement := ar.Element
	ar.Element = func(e *parser.Element) error {
		switch e.Name {
		case "link":
			inHrefElement = true
		case "img":
			inSrcElement = true
		}
		// Visit child elements.
		if err := visitElement(e); err != nil {
			return err
		}
		inHrefElement = false
		inSrcElement = false
		return nil
	}

	// Save the original ScriptElement visitor to allow chaining.
	visitScriptElement := ar.ScriptElement
	ar.ScriptElement = func(e *parser.ScriptElement) error {
		inSrcElement = true
		// Visit child script elements.
		if err := visitScriptElement(e); err != nil {
			return err
		}
		inSrcElement = false
		return nil
	}

	// Save the original ConstantAttribute visitor to allow chaining.
	visitConstantAttribute := ar.ConstantAttribute
	ar.ConstantAttribute = func(n *parser.ConstantAttribute) error {
		if inSrcElement && n.Key.String() == "src" {
			n.Value = rewrite(n.Value)
		}
		if inHrefElement && n.Key.String() == "href" {
			n.Value = rewrite(n.Value)
		}
		return visitConstantAttribute(n)
	}

	return ar
}

func TestVisitorAssetRewriter(t *testing.T) {
	input := `package view

templ indexStyles() {
	<link rel="stylesheet" href="view/index.css"/>
}

templ indexScripts() {
	<script src="view/index.js" type="module" defer></script>
}

templ Index(title string, res *hn.SearchResponse) {
	@layout(&Layout{
		Styles:  indexStyles,
		Scripts: indexScripts,
	}) {
		@index(res)
	}
}

templ index(res *hn.SearchResponse) {
	<main class="index">
		@Header()
		for _, story := range res.Stories {
			@storyComponent(story)
		}
		@Pagination(res.Page, res.NumPages)
	</main>
}
`
	templateFile, err := parser.ParseString(input)
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}
	rewriter := assetRewriter(func(assetPath string) string {
		return "http://somecdn.com/" + assetPath
	})
	if err := templateFile.Visit(rewriter); err != nil {
		t.Fatalf("error visiting template file: %v", err)
	}

	var actual bytes.Buffer
	if err := templateFile.Write(&actual); err != nil {
		t.Fatalf("error writing template file: %v", err)
	}
	expected := `package view

templ indexStyles() {
	<link rel="stylesheet" href="http://somecdn.com/view/index.css"/>
}

templ indexScripts() {
	<script src="http://somecdn.com/view/index.js" type="module" defer></script>
}

templ Index(title string, res *hn.SearchResponse) {
	@layout(&Layout{
		Styles:  indexStyles,
		Scripts: indexScripts,
	}) {
		@index(res)
	}
}

templ index(res *hn.SearchResponse) {
	<main class="index">
		@Header()
		for _, story := range res.Stories {
			@storyComponent(story)
		}
		@Pagination(res.Page, res.NumPages)
	</main>
}
`
	if diff := cmp.Diff(expected, actual.String()); diff != "" {
		t.Fatalf("expected != actual:\n%s", diff)
	}
}
