package templ

import (
	"errors"
	"strings"
	"testing"
)

type urlTest struct {
	url             string
	expectSanitized bool
}

var urlTests = []urlTest{
	{"//example.com", false},
	{"/", false},
	{"/index", false},
	{"http://example.com", false},
	{"https://example.com", false},
	{"mailto:test@example.com", false},
	{"tel:+1234567890", false},
	{"ftp://example.com", false},
	{"ftps://example.com", false},
	{"irc://example.com", true},
	{"bitcoin://example.com", true},
}

func testURL(t *testing.T, url string, expectSanitized bool) {
	u := URL(url)
	wasSanitized := u == FailedSanitizationURL
	if expectSanitized != wasSanitized {
		t.Errorf("expected sanitized=%v, got %v", expectSanitized, wasSanitized)
	}
}

func TestURL(t *testing.T) {
	for _, test := range urlTests {
		t.Run(test.url, func(t *testing.T) {
			testURL(t, test.url, test.expectSanitized)
		})
		test.url = strings.ToUpper(test.url)
		t.Run(strings.ToUpper(test.url), func(t *testing.T) {
			testURL(t, test.url, test.expectSanitized)
		})
	}
}

func BenchmarkURL(b *testing.B) {
	for range b.N {
		for _, test := range urlTests {
			u := URL(test.url)
			wasSanitized := u == FailedSanitizationURL
			if test.expectSanitized != wasSanitized {
				b.Errorf("expected sanitized=%v, got %v", test.expectSanitized, wasSanitized)
			}
		}
	}
}

func TestJoinURLErrs(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	type CustomString string

	checkResult := func(t *testing.T, result SafeURL, err error, expected SafeURL, expectedErrs ...error) {
		t.Helper()
		if result != expected {
			t.Errorf("expected result %q, got %q", expected, result)
		}
		if len(expectedErrs) == 0 {
			if err != nil {
				t.Errorf("expected nil error, got %v", err)
			}
			return
		}
		if err == nil {
			t.Errorf("expected error(s) %v, got nil", expectedErrs)
			return
		}
		for _, exp := range expectedErrs {
			if !errors.Is(err, exp) {
				t.Errorf("expected error to contain %v, but it did not", exp)
			}
		}
	}
	t.Run("strings are sanitized", func(t *testing.T) {
		result, err := JoinURLErrs("javascript:alert(1)")
		checkResult(t, result, err, FailedSanitizationURL)
	})
	t.Run("custom string types are sanitized", func(t *testing.T) {
		result, err := JoinURLErrs(CustomString("javascript:alert(1)"))
		checkResult(t, result, err, FailedSanitizationURL)
	})
	t.Run("SafeURLs bypass sanitization", func(t *testing.T) {
		safeURL := SafeURL("javascript:alert(1)")
		result, err := JoinURLErrs(safeURL)
		checkResult(t, result, err, safeURL)
	})
	t.Run("safe URL strings are returned unchanged", func(t *testing.T) {
		result, err := JoinURLErrs("https://example.com")
		checkResult(t, result, err, SafeURL("https://example.com"))
	})
	t.Run("single errors are joined", func(t *testing.T) {
		result, err := JoinURLErrs("https://example.com", err1)
		checkResult(t, result, err, SafeURL("https://example.com"), err1)
	})
	t.Run("multiple errors are joined", func(t *testing.T) {
		result, err := JoinURLErrs("https://example.com", err1, err2)
		checkResult(t, result, err, SafeURL("https://example.com"), err1, err2)
	})
	t.Run("nil errors are preserved", func(t *testing.T) {
		result, err := JoinURLErrs("https://example.com", nil)
		checkResult(t, result, err, SafeURL("https://example.com"))
	})
}
