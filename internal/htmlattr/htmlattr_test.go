package htmlattr

import "testing"

// TestClassify covers GHSA-94c2-xg24-pwhv: every attribute whose value can
// execute JavaScript in the context of the page must be classified, regardless
// of the case of the element and attribute names. URL attributes are
// classified in the same way as html/template.
func TestClassify(t *testing.T) {
	tests := []struct {
		name        string
		elementName string
		attrName    string
		expected    Context
	}{
		{name: "a href is a URL", elementName: "a", attrName: "href", expected: ContextURL},
		{name: "uppercase HREF is a URL", elementName: "a", attrName: "HREF", expected: ContextURL},
		{name: "mixed case Href is a URL", elementName: "a", attrName: "Href", expected: ContextURL},
		{name: "uppercase element names are compared in lowercase", elementName: "A", attrName: "href", expected: ContextURL},
		{name: "area href is a URL", elementName: "area", attrName: "href", expected: ContextURL},
		{name: "base href is a URL", elementName: "base", attrName: "href", expected: ContextURL},
		{name: "svg a xlink:href is a URL", elementName: "a", attrName: "xlink:href", expected: ContextURL},
		{name: "uppercase XLINK:HREF is a URL", elementName: "a", attrName: "XLINK:HREF", expected: ContextURL},
		{name: "button formaction is a URL", elementName: "button", attrName: "formaction", expected: ContextURL},
		{name: "mixed case input FormAction is a URL", elementName: "input", attrName: "FormAction", expected: ContextURL},
		{name: "form action is a URL", elementName: "form", attrName: "action", expected: ContextURL},
		{name: "action on a custom element is a URL, as in html/template", elementName: "turbo-stream", attrName: "action", expected: ContextURL},
		{name: "object data is a URL", elementName: "object", attrName: "data", expected: ContextURL},
		{name: "data on a div is a URL, as in html/template", elementName: "div", attrName: "data", expected: ContextURL},
		{name: "iframe src is a URL", elementName: "iframe", attrName: "src", expected: ContextURL},
		{name: "uppercase IFRAME SRC is a URL", elementName: "IFRAME", attrName: "SRC", expected: ContextURL},
		{name: "img src is a URL", elementName: "img", attrName: "src", expected: ContextURL},
		{name: "script src is a URL", elementName: "script", attrName: "src", expected: ContextURL},
		{name: "img srcset is a URL", elementName: "img", attrName: "srcset", expected: ContextURL},
		{name: "video poster is a URL", elementName: "video", attrName: "poster", expected: ContextURL},
		{name: "blockquote cite is a URL", elementName: "blockquote", attrName: "cite", expected: ContextURL},
		{name: "body background is a URL", elementName: "body", attrName: "background", expected: ContextURL},
		{name: "object archive is a URL", elementName: "object", attrName: "archive", expected: ContextURL},
		{name: "object classid is a URL", elementName: "object", attrName: "classid", expected: ContextURL},
		{name: "object codebase is a URL", elementName: "object", attrName: "codebase", expected: ContextURL},
		{name: "command icon is a URL", elementName: "command", attrName: "icon", expected: ContextURL},
		{name: "img longdesc is a URL", elementName: "img", attrName: "longdesc", expected: ContextURL},
		{name: "html manifest is a URL", elementName: "html", attrName: "manifest", expected: ContextURL},
		{name: "head profile is a URL", elementName: "head", attrName: "profile", expected: ContextURL},
		{name: "img usemap is a URL", elementName: "img", attrName: "usemap", expected: ContextURL},
		{name: "xmlns is a URL", elementName: "svg", attrName: "xmlns", expected: ContextURL},
		{name: "namespaced xmlns:xlink is a URL", elementName: "svg", attrName: "xmlns:xlink", expected: ContextURL},
		{name: "data-href is a URL, because data- is removed before classification", elementName: "div", attrName: "data-href", expected: ContextURL},
		{name: "data-action is a URL, because data- is removed before classification", elementName: "div", attrName: "data-action", expected: ContextURL},
		{name: "data-url is a URL, because the name contains url", elementName: "div", attrName: "data-url", expected: ContextURL},
		{name: "data-lazy-src is a URL, because the name contains src", elementName: "img", attrName: "data-lazy-src", expected: ContextURL},
		{name: "data-uri is a URL, because the name contains uri", elementName: "div", attrName: "data-uri", expected: ContextURL},
		{name: "custom namespaced g:tweetUrl is a URL, because the name contains url", elementName: "div", attrName: "g:tweetUrl", expected: ContextURL},
		{name: "names that contain uri incidentally are URLs, as in html/template", elementName: "div", attrName: "security", expected: ContextURL},
		{name: "hx-get is HTML-escaped", elementName: "div", attrName: "hx-get", expected: ContextDefault},
		{name: "class is HTML-escaped", elementName: "a", attrName: "class", expected: ContextDefault},
		{name: "title is HTML-escaped", elementName: "a", attrName: "title", expected: ContextDefault},
		{name: "data-testid is HTML-escaped", elementName: "div", attrName: "data-testid", expected: ContextDefault},
		{name: "onclick is an event handler", elementName: "button", attrName: "onclick", expected: ContextEventHandler},
		{name: "camel case onClick is an event handler", elementName: "button", attrName: "onClick", expected: ContextEventHandler},
		{name: "pascal case OnClick is an event handler", elementName: "button", attrName: "OnClick", expected: ContextEventHandler},
		{name: "uppercase ONCLICK is an event handler", elementName: "button", attrName: "ONCLICK", expected: ContextEventHandler},
		{name: "htmx 1 hx-on is an event handler", elementName: "div", attrName: "hx-on", expected: ContextEventHandler},
		{name: "hx-on:click is an event handler", elementName: "div", attrName: "hx-on:click", expected: ContextEventHandler},
		{name: "hx-on::after-request is an event handler", elementName: "div", attrName: "hx-on::after-request", expected: ContextEventHandler},
		{name: "uppercase HX-ON:click is an event handler", elementName: "div", attrName: "HX-ON:click", expected: ContextEventHandler},
		{name: "hx-on-click is an event handler", elementName: "div", attrName: "hx-on-click", expected: ContextEventHandler},
		{name: "hx-on--before-request is an event handler", elementName: "div", attrName: "hx-on--before-request", expected: ContextEventHandler},
		{name: "data-hx-on:click is an event handler", elementName: "div", attrName: "data-hx-on:click", expected: ContextEventHandler},
		{name: "uppercase DATA-HX-ON-CLICK is an event handler", elementName: "div", attrName: "DATA-HX-ON-CLICK", expected: ContextEventHandler},
		{name: "iframe srcdoc is an HTML document", elementName: "iframe", attrName: "srcdoc", expected: ContextSrcdoc},
		{name: "uppercase IFRAME SRCDOC is an HTML document", elementName: "IFRAME", attrName: "SRCDOC", expected: ContextSrcdoc},
		{name: "srcdoc on other elements is an HTML document, as in html/template", elementName: "my-element", attrName: "srcdoc", expected: ContextSrcdoc},
		{name: "data-srcdoc is an HTML document, as in html/template", elementName: "div", attrName: "data-srcdoc", expected: ContextSrcdoc},
		{name: "set to is an animation value", elementName: "set", attrName: "to", expected: ContextAnimationValue},
		{name: "animate values is an animation value", elementName: "animate", attrName: "values", expected: ContextAnimationValue},
		{name: "animate from is an animation value", elementName: "animate", attrName: "from", expected: ContextAnimationValue},
		{name: "animate by is an animation value", elementName: "animate", attrName: "by", expected: ContextAnimationValue},
		{name: "uppercase SET TO is an animation value", elementName: "SET", attrName: "TO", expected: ContextAnimationValue},
		{name: "to on an unknown element is an animation value", elementName: "", attrName: "to", expected: ContextAnimationValue},
		{name: "animate dur is HTML-escaped", elementName: "animate", attrName: "dur", expected: ContextDefault},
		{name: "to on a div is HTML-escaped", elementName: "div", attrName: "to", expected: ContextDefault},
		{name: "animateTransform to is HTML-escaped", elementName: "animateTransform", attrName: "to", expected: ContextDefault},
		{name: "style is CSS", elementName: "div", attrName: "style", expected: ContextStyle},
		{name: "uppercase STYLE is CSS", elementName: "div", attrName: "STYLE", expected: ContextStyle},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Classify(tt.elementName, tt.attrName); got != tt.expected {
				t.Errorf("got %d, expected %d", got, tt.expected)
			}
		})
	}
}

func TestToLowerASCII(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "ASCII letters are lowercased", input: "HREF", expected: "href"},
		{name: "lowercase letters and punctuation are unchanged", input: "hx-on:click", expected: "hx-on:click"},
		{name: "non-ASCII letters are unchanged, as in the HTML tokenizer", input: "Key", expected: "Key"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toLowerASCII(tt.input); got != tt.expected {
				t.Errorf("got %q, expected %q", got, tt.expected)
			}
		})
	}
}

func TestIsAnimationTargetURL(t *testing.T) {
	tests := []struct {
		name          string
		attributeName string
		expected      bool
	}{
		{name: "href is a URL target", attributeName: "href", expected: true},
		{name: "xlink:href is a URL target", attributeName: "xlink:href", expected: true},
		{name: "uppercase HREF is a URL target", attributeName: "HREF", expected: true},
		{name: "href with leading and trailing spaces is a URL target", attributeName: " href ", expected: true},
		{name: "href with surrounding tabs and newlines is a URL target", attributeName: "\t\nhref\r\f", expected: true},
		{name: "xlink:href with surrounding spaces is a URL target", attributeName: " xlink:href ", expected: true},
		{name: "href with a namespace prefix other than xlink is a URL target", attributeName: "x:href", expected: true},
		{name: "opacity is not a URL target", attributeName: "opacity", expected: false},
		{name: "opacity with surrounding spaces is not a URL target", attributeName: " opacity ", expected: false},
		{name: "hreflang is not a URL target", attributeName: "hreflang", expected: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAnimationTargetURL(tt.attributeName); got != tt.expected {
				t.Errorf("got %t, expected %t", got, tt.expected)
			}
		})
	}
}

func TestIsValidName(t *testing.T) {
	tests := []struct {
		name     string
		attrName string
		expected bool
	}{
		{name: "lowercase names are valid", attrName: "class", expected: true},
		{name: "names with hyphens are valid", attrName: "data-test-id", expected: true},
		{name: "names with colons are valid", attrName: "hx-on:click", expected: true},
		{name: "names with asterisks are valid", attrName: "hx-target-*", expected: true},
		{name: "names with at signs are valid", attrName: "@click", expected: true},
		{name: "names with non-ASCII letters are valid", attrName: "data-größe", expected: true},
		{name: "empty names are invalid", attrName: "", expected: false},
		{name: "names with spaces are invalid", attrName: "x onclick", expected: false},
		{name: "names with tabs are invalid", attrName: "x\tonclick", expected: false},
		{name: "names with newlines are invalid", attrName: "x\nonclick", expected: false},
		{name: "names with form feeds are invalid", attrName: "x\fonclick", expected: false},
		{name: "names with NUL are invalid", attrName: "x\x00", expected: false},
		{name: "names with C1 control characters are invalid", attrName: "x\u0085", expected: false},
		{name: "names with equals signs are invalid", attrName: "x=y", expected: false},
		{name: "names with double quotes are invalid", attrName: `x"`, expected: false},
		{name: "names with single quotes are invalid", attrName: "x'", expected: false},
		{name: "names with greater than signs are invalid", attrName: "x>", expected: false},
		{name: "names with less than signs are invalid", attrName: "x<", expected: false},
		{name: "names with slashes are invalid", attrName: "x/", expected: false},
		{name: "names with backticks are invalid", attrName: "x`", expected: false},
		{name: "names with ampersands are invalid", attrName: "x&y", expected: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidName(tt.attrName); got != tt.expected {
				t.Errorf("got %t, expected %t", got, tt.expected)
			}
		})
	}
}
