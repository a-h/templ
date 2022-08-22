package turbo

import (
	"bytes"
	"context"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/a-h/templ"
)

var contentTemplate = templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
	_, err := io.WriteString(w, "content")
	return err
})

func TestStreamReplace(t *testing.T) {
	// Arrange.
	w := httptest.NewRecorder()

	if err := Replace(w, "replaceTarget", contentTemplate); err != nil {
		t.Fatalf("replace failed: %v", err)
	}

	// Act.
	bdy := w.Body.Bytes()
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(bdy))
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	// Assert.
	selection := doc.Find(`turbo-stream[action="replace"][target="replaceTarget"]`)
	if selection.Length() != 1 {
		t.Error("expected to find a replace action, but didn't")
	}
	text := selection.First().Text()
	if text != "content" {
		t.Errorf("expected 'content' to be the text, but got %q\n%s", text, bdy)
	}
}

func TestStream(t *testing.T) {
	w := httptest.NewRecorder()

	// Append.
	expectedAppends := 2
	for i := 0; i < expectedAppends; i++ {
		if err := Append(w, "appendTarget", contentTemplate); err != nil {
			t.Fatalf("append failed: %v", err)
		}
	}
	// Prepend.
	expectedPrepends := 3
	for i := 0; i < expectedPrepends; i++ {
		if err := Prepend(w, "prependTarget", contentTemplate); err != nil {
			t.Fatalf("prepend failed: %v", err)
		}
	}
	// Replace.
	expectedReplaces := 4
	for i := 0; i < expectedReplaces; i++ {
		if err := Replace(w, "replaceTarget", contentTemplate); err != nil {
			t.Fatalf("replace failed: %v", err)
		}
	}
	// Update.
	expectedUpdates := 5
	for i := 0; i < expectedUpdates; i++ {
		if err := Update(w, "updateTarget", contentTemplate); err != nil {
			t.Fatalf("update failed: %v", err)
		}
	}
	// Remove.
	expectedRemoves := 6
	for i := 0; i < expectedRemoves; i++ {
		if err := Remove(w, "removeTarget"); err != nil {
			t.Fatalf("remove failed: %v", err)
		}
	}

	doc, err := goquery.NewDocumentFromReader(w.Body)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	if appendCount := doc.Find(`turbo-stream[action="append"][target="appendTarget"]`).Length(); appendCount != expectedAppends {
		t.Errorf("expected %d append actions, but got %d", expectedAppends, appendCount)
	}
	if prependCount := doc.Find(`turbo-stream[action="prepend"][target="prependTarget"]`).Length(); prependCount != expectedPrepends {
		t.Errorf("expected %d prepend actions, but got %d", expectedPrepends, prependCount)
	}
	if replaceCount := doc.Find(`turbo-stream[action="replace"][target="replaceTarget"]`).Length(); replaceCount != expectedReplaces {
		t.Errorf("expected %d replace actions, but got %d", expectedReplaces, replaceCount)
	}
	if updateCount := doc.Find(`turbo-stream[action="update"][target="updateTarget"]`).Length(); updateCount != expectedUpdates {
		t.Errorf("expected %d update actions, but got %d", expectedUpdates, updateCount)
	}
	if removeCount := doc.Find(`turbo-stream[action="remove"][target="removeTarget"]`).Length(); removeCount != expectedRemoves {
		t.Errorf("expected %d remove actions, but got %d", expectedRemoves, removeCount)
	}
	if w.Result().Header.Get("Content-Type") != "text/vnd.turbo-stream.html" {
		t.Errorf("expected Content-Type %q, got %q", "text/vnd.turbo-stream.html", w.Result().Header.Get("Content-Type"))
	}
}

func TestIsTurboRequest(t *testing.T) {
	turboRequest := httptest.NewRequest("GET", "/", nil)
	if IsTurboRequest(turboRequest) {
		t.Error("request was incorrectly recognised as a Turbo stream request")
	}
	turboRequest.Header.Add("accept", "text/vnd.turbo-stream.html")
	if !IsTurboRequest(turboRequest) {
		t.Error("request not correctly recognised as a Turbo stream request")
	}
}
