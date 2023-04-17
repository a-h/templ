package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func TestHeader(t *testing.T) {
	// Pipe the rendered template into goquery.
	r, w := io.Pipe()
	go func() {
		_ = headerTemplate("Posts").Render(context.Background(), w)
		_ = w.Close()
	}()
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		t.Fatalf("failed to read template: %v", err)
	}
	// Expect the component to include a testid.
	if doc.Find(`[data-testid="headerTemplate"]`).Length() == 0 {
		t.Error("expected data-testid attribute to be rendered, but it wasn't")
	}
	// Expect the page name to be set correctly.
	expectedPageName := "Posts"
	if actualPageName := doc.Find("h1").Text(); actualPageName != expectedPageName {
		t.Errorf("expected page name %q, got %q", expectedPageName, actualPageName)
	}
}

func TestFooter(t *testing.T) {
	// Pipe the rendered template into goquery.
	r, w := io.Pipe()
	go func() {
		_ = footerTemplate().Render(context.Background(), w)
		_ = w.Close()
	}()
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		t.Fatalf("failed to read template: %v", err)
	}
	// Expect the component to include a testid.
	if doc.Find(`[data-testid="footerTemplate"]`).Length() == 0 {
		t.Error("expected data-testid attribute to be rendered, but it wasn't")
	}
	// Expect the copyright notice to include the current year.
	expectedCopyrightNotice := fmt.Sprintf("© %d", time.Now().Year())
	if actualCopyrightNotice := doc.Find("div").Text(); actualCopyrightNotice != expectedCopyrightNotice {
		t.Errorf("expected copyright notice %q, got %q", expectedCopyrightNotice, actualCopyrightNotice)
	}
}

func TestNav(t *testing.T) {
	r, w := io.Pipe()
	go func() {
		_ = navTemplate().Render(context.Background(), w)
		_ = w.Close()
	}()
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		t.Fatalf("failed to read template: %v", err)
	}
	// Expect the component to include a testid.
	if doc.Find(`[data-testid="navTemplate"]`).Length() == 0 {
		t.Error("expected data-testid attribute to be rendered, but it wasn't")
	}
}

func TestHome(t *testing.T) {
	r, w := io.Pipe()
	go func() {
		_ = home().Render(context.Background(), w)
		_ = w.Close()
	}()
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		t.Fatalf("failed to read template: %v", err)
	}
	// Expect the page title to be set correctly.
	expectedTitle := "Home"
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
	// Expect the home template be rendered.
	if doc.Find(`[data-testid="homeTemplate"]`).Length() == 0 {
		t.Error("expected homeTemplate to be rendered, but it wasn't")
	}
}

func TestPosts(t *testing.T) {
	testPosts := []Post{
		{Name: "Name1", Author: "Author1"},
		{Name: "Name2", Author: "Author2"},
	}
	r, w := io.Pipe()
	go func() {
		_ = posts(testPosts).Render(context.Background(), w)
		_ = w.Close()
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

func TestPostsHandler(t *testing.T) {
	tests := []struct {
		name           string
		postGetter     func() (posts []Post, err error)
		expectedStatus int
		assert         func(doc *goquery.Document)
	}{
		{
			name: "database errors result in a 500 error",
			postGetter: func() (posts []Post, err error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			assert: func(doc *goquery.Document) {
				expected := "failed to retrieve posts\n"
				if actual := doc.Text(); actual != expected {
					t.Errorf("expected error message %q, got %q", expected, actual)
				}
			},
		},
		{
			name: "database success renders the posts",
			postGetter: func() (posts []Post, err error) {
				return []Post{
					{Name: "Name1", Author: "Author1"},
					{Name: "Name2", Author: "Author2"},
				}, nil
			},
			expectedStatus: http.StatusInternalServerError,
			assert: func(doc *goquery.Document) {
				if doc.Find(`[data-testid="postsTemplate"]`).Length() == 0 {
					t.Error("expected posts to be rendered, but it wasn't")
				}
			},
		},
	}
	for _, test := range tests {
		// Arrange.
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/posts", nil)

		ph := NewPostsHandler()
		ph.Log = log.New(io.Discard, "", 0) // Suppress logging.
		ph.GetPosts = test.postGetter

		// Act.
		ph.ServeHTTP(w, r)
		doc, err := goquery.NewDocumentFromReader(w.Result().Body)
		if err != nil {
			t.Fatalf("failed to read template: %v", err)
		}

		// Assert.
		test.assert(doc)
	}
}
