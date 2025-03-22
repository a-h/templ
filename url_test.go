package templ

import (
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
