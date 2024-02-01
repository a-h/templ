# Testing

To test that data is rendered as expected, there are two main ways to do it:

* Expectation testing - testing that specific expectations are met by the output.
* Snapshot testing - testing that outputs match a pre-written output.

## Expectation testing

Expectation validates that data appears in the output in the right format, and position.

The example at https://github.com/a-h/templ/blob/main/examples/blog/posts_test.go uses the `goquery` library to make assertions on the HTML.

```go
func TestPosts(t *testing.T) {
	testPosts := []Post{
		{Name: "Name1", Author: "Author1"},
		{Name: "Name2", Author: "Author2"},
	}
	r, w := io.Pipe()
	go func() {
		posts(testPosts).Render(context.Background(), w)
		w.Close()
	}()
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		t.Fatalf("failed to read template: %v", err)
	}
	// Assert.
	// Expect the page title to be set correctly.
	expectedTitle := "Posts"
	if actualTitle := doc.Find("title").Text(); actualTitle != expectedTitle {
		t.Errorf("expected title name %q, got %q", expectedTitle, actualTitle)
	}
	// Expect the header to be rendered.
	if doc.Find(`[data-testid="headerTemplate"]`).Length() == 0 {
		t.Error("expected data-testid attribute to be rendered, but it wasn't")
	}
	// Expect the navigation to be rendered.
	if doc.Find(`[data-testid="navTemplate"]`).Length() == 0 {
		t.Error("expected nav to be rendered, but it wasn't")
	}
	// Expect the posts to be rendered.
	if doc.Find(`[data-testid="postsTemplate"]`).Length() == 0 {
		t.Error("expected posts to be rendered, but it wasn't")
	}
	// Expect both posts to be rendered.
	if actualPostCount := doc.Find(`[data-testid="postsTemplatePost"]`).Length(); actualPostCount != len(testPosts) {
		t.Fatalf("expected %d posts to be rendered, found %d", len(testPosts), actualPostCount)
	}
	// Expect the posts to contain the author name.
	doc.Find(`[data-testid="postsTemplatePost"]`).Each(func(index int, sel *goquery.Selection) {
		expectedName := testPosts[index].Name
		if actualName := sel.Find(`[data-testid="postsTemplatePostName"]`).Text(); actualName != expectedName {
			t.Errorf("expected name %q, got %q", actualName, expectedName)
		}
		expectedAuthor := testPosts[index].Author
		if actualAuthor := sel.Find(`[data-testid="postsTemplatePostAuthor"]`).Text(); actualAuthor != expectedAuthor {
			t.Errorf("expected author %q, got %q", actualAuthor, expectedAuthor)
		}
	})
}
```

## Snapshot testing

Snapshot testing is a more broad check. It simply checks that the output hasn't changed since the last time you took a copy of the output.

It relies on manually checking the output to make sure it's correct, and then "locking it in" by using the snapshot.

templ uses this strategy to check for regressions in behaviour between releases, as per https://github.com/a-h/templ/blob/main/generator/test-html-comment/render_test.go

To make it easier to compare the output against the expected HTML, templ uses a HTML formatting library before executing the diff.

```go
package testcomment

import (
	_ "embed"
	"testing"

	"github.com/a-h/templ/generator/htmldiff"
)

//go:embed expected.html
var expected string

func Test(t *testing.T) {
	component := render("sample content")

	diff, err := htmldiff.Diff(component, expected)
	if err != nil {
		t.Fatal(err)
	}
	if diff != "" {
		t.Error(diff)
	}
}
```
