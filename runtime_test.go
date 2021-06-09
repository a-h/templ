package templ

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRenderedCSSClassesFromContext(t *testing.T) {
	ctx := context.Background()
	ctx, classes := RenderedCSSClassesFromContext(ctx)
	if classes.Contains("test") {
		t.Fatalf("before the classes have been set, test should not be set")
	}
	classes.Add("test")
	if !classes.Contains("test") {
		t.Errorf("expected 'test' to be present in the context, after setting")
	}
	_, updatedClasses := RenderedCSSClassesFromContext(ctx)
	if !updatedClasses.Contains("test") {
		t.Errorf("expected 'test' to be present in the context with new context, but it wasn't")
	}
	if classes != updatedClasses {
		t.Errorf("expected %v to be the same as %v", classes, updatedClasses)
	}
}

func TestCSSHandler(t *testing.T) {
	var tests = []struct {
		name             string
		input            []CSSClass
		expectedMIMEType string
		expectedBody     string
	}{
		{
			name:             "no classes",
			input:            nil,
			expectedMIMEType: "text/css",
			expectedBody:     "",
		},
		{
			name:             "constant classes don't have CSS to return",
			input:            []CSSClass{ConstantCSSClass("none")},
			expectedMIMEType: "text/css",
			expectedBody:     "",
		},
		{
			name:             "constant classes don't have CSS to return",
			input:            []CSSClass{ConstantCSSClass("none"), ComponentCSSClass{ID: "className", Class: SafeCSS(".className{background-color:white;}")}},
			expectedMIMEType: "text/css",
			expectedBody:     ".className{background-color:white;}",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			h := NewCSSHandler(tt.input...)
			h.ServeHTTP(w, &http.Request{})
			if diff := cmp.Diff(tt.expectedMIMEType, w.Header().Get("Content-Type")); diff != "" {
				t.Errorf(diff)
			}
			if diff := cmp.Diff(tt.expectedBody, w.Body.String()); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}
