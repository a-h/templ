package turbo

import (
	"bytes"
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/a-h/templ/internal/htmlfind"
	"golang.org/x/net/html"
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
	doc, err := html.Parse(bytes.NewReader(bdy))
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	// Assert.
	match := htmlfind.Element("turbo-stream", htmlfind.Attr("action", "replace"), htmlfind.Attr("target", "replaceTarget"))
	var count int
	for _, node := range htmlfind.All(doc, match) {
		count++
		if strings.TrimSpace(node.FirstChild.FirstChild.Data) != "content" {
			t.Errorf("expected 'content' to be the text, but got %q\n%s", node.FirstChild.Data, bdy)
		}
	}
	if count != 1 {
		t.Errorf("expected to find a single replace action, but found %d", count)
	}
}

func TestStream(t *testing.T) {
	w := httptest.NewRecorder()

	// Append.
	expectedAppends := 2
	for range expectedAppends {
		if err := Append(w, "appendTarget", contentTemplate); err != nil {
			t.Fatalf("append failed: %v", err)
		}
	}
	// Prepend.
	expectedPrepends := 3
	for range expectedPrepends {
		if err := Prepend(w, "prependTarget", contentTemplate); err != nil {
			t.Fatalf("prepend failed: %v", err)
		}
	}
	// Replace.
	expectedReplaces := 4
	for range expectedReplaces {
		if err := Replace(w, "replaceTarget", contentTemplate); err != nil {
			t.Fatalf("replace failed: %v", err)
		}
	}
	// Update.
	expectedUpdates := 5
	for range expectedUpdates {
		if err := Update(w, "updateTarget", contentTemplate); err != nil {
			t.Fatalf("update failed: %v", err)
		}
	}
	// Remove.
	expectedRemoves := 6
	for range expectedRemoves {
		if err := Remove(w, "removeTarget"); err != nil {
			t.Fatalf("remove failed: %v", err)
		}
	}

	doc, err := html.Parse(w.Body)
	if err != nil {
		t.Fatalf("failed to read response: %v", err)
	}

	appends := htmlfind.All(doc, htmlfind.Element("turbo-stream", htmlfind.Attr("action", "append"), htmlfind.Attr("target", "appendTarget")))
	if len(appends) != expectedAppends {
		t.Errorf("expected %d append actions, but got %d", expectedAppends, len(appends))
	}
	prepends := htmlfind.All(doc, htmlfind.Element("turbo-stream", htmlfind.Attr("action", "prepend"), htmlfind.Attr("target", "prependTarget")))
	if len(prepends) != expectedPrepends {
		t.Errorf("expected %d prepend actions, but got %d", expectedPrepends, len(prepends))
	}
	replaces := htmlfind.All(doc, htmlfind.Element("turbo-stream", htmlfind.Attr("action", "replace"), htmlfind.Attr("target", "replaceTarget")))
	if len(replaces) != expectedReplaces {
		t.Errorf("expected %d replace actions, but got %d", expectedReplaces, len(replaces))
	}
	updates := htmlfind.All(doc, htmlfind.Element("turbo-stream", htmlfind.Attr("action", "update"), htmlfind.Attr("target", "updateTarget")))
	if len(updates) != expectedUpdates {
		t.Errorf("expected %d update actions, but got %d", expectedUpdates, len(updates))
	}
	removes := htmlfind.All(doc, htmlfind.Element("turbo-stream", htmlfind.Attr("action", "remove"), htmlfind.Attr("target", "removeTarget")))
	if len(removes) != expectedRemoves {
		t.Errorf("expected %d remove actions, but got %d", expectedRemoves, len(removes))
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
