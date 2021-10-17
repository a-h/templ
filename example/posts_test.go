package main

import (
	"context"
	"io"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestPosts(t *testing.T) {
	testPosts := []Post{
		{Name: "Name1", Author: "Author1"},
		{Name: "Name2", Author: "Author2"},
	}
	// Pipe the rendered template into goquery.
	r, w := io.Pipe()
	go func() {
		posts(testPosts).Render(context.Background(), w)
		w.Close()
	}()
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		t.Fatalf("failed to read template: %v", err)
	}
	// Expect the page title to be set correctly.
	expectedTitle := "Posts"
	if actualTitle := doc.Find("title").Text(); actualTitle != expectedTitle {
		t.Errorf("expected title name %q, got %q", expectedTitle, actualTitle)
	}
	// Expect the page name to be set correctly.
	expectedPageName := "Posts"
	if actualPageName := doc.Find("h1").Text(); actualPageName != expectedPageName {
		t.Errorf("expected page name %q, got %q", expectedPageName, actualPageName)
	}
	// Expect the navigation to be rendered.
	if doc.Find(`[data-testid="navTemplate"]`).Length() == 0 {
		t.Error("expected nav to be rendered, but it wasn't")
	}
	// Expect both posts to be rendered.
	if actualPostCount := doc.Find(`[data-testid="post"]`).Length(); actualPostCount != len(testPosts) {
		t.Errorf("expected %d posts to be rendered, found %d", len(testPosts), actualPostCount)
	}
}
