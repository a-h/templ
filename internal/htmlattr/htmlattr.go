// Package htmlattr classifies HTML attributes by the escaping and
// sanitization that their values require. The generator uses it for
// attributes with constant names, and the runtime uses it for attributes with
// dynamic names, and spread attributes, so that both apply the same rules.
//
// See GHSA-94c2-xg24-pwhv, where only a fixed list of lowercase attribute
// names was sanitized, so values in attributes such as <iframe src>,
// <button formaction>, and <a HREF> could execute JavaScript.
package htmlattr

import "strings"

// Context describes the position that an attribute value is written into,
// which determines how the value must be escaped and sanitized.
type Context int

const (
	// ContextDefault values are HTML-escaped.
	ContextDefault Context = iota
	// ContextURL values are URLs. Values are sanitized with templ.URL.
	ContextURL
	// ContextEventHandler values are executed as JavaScript by the browser, or
	// by htmx. Values must be a templ.ComponentScript.
	ContextEventHandler
	// ContextSrcdoc values are parsed as an HTML document in the context of the
	// page. String values are escaped as text.
	ContextSrcdoc
	// ContextAnimationValue values are SVG animation values, which can set the
	// href of the parent element to a javascript: URL. Each item in the
	// semicolon separated list is sanitized with templ.URL.
	ContextAnimationValue
	// ContextStyle values are CSS.
	ContextStyle
)

// Classify returns the context of the value of the attribute on the element.
//
// HTML element and attribute names are case insensitive, so Classify
// lowercases both names before comparing them. Browsers only lowercase the
// ASCII letters in names, so Classify does the same.
//
// URL attributes are classified in the same way as html/template's attrType,
// i.e. by attribute name on any element, including data-* attributes and
// custom attributes whose names contain "src", "uri", or "url".
//
// An empty element name means that the element is unknown, so the
// SVG animation value rules apply.
func Classify(elementName, attrName string) Context {
	elementName = toLowerASCII(elementName)
	attrName = toLowerASCII(attrName)

	if strings.HasPrefix(attrName, "on") || strings.HasPrefix(attrName, "hx-on") || strings.HasPrefix(attrName, "data-hx-on") {
		return ContextEventHandler
	}
	if attrName == "style" {
		return ContextStyle
	}
	if elementName == "" || elementName == "set" || elementName == "animate" {
		switch attrName {
		case "to", "from", "values", "by":
			return ContextAnimationValue
		}
	}

	// The rules below match html/template's attrType.
	name := attrName
	if after, ok := strings.CutPrefix(name, "data-"); ok {
		name = after
	} else if prefix, local, ok := strings.Cut(name, ":"); ok {
		if prefix == "xmlns" {
			return ContextURL
		}
		name = local
	}
	switch name {
	case "srcdoc":
		return ContextSrcdoc
	case "action", "archive", "background", "cite", "classid", "codebase", "data", "formaction",
		"href", "icon", "longdesc", "manifest", "poster", "profile", "src", "usemap", "xmlns":
		return ContextURL
	}
	if strings.Contains(name, "src") || strings.Contains(name, "uri") || strings.Contains(name, "url") {
		return ContextURL
	}
	return ContextDefault
}

// toLowerASCII lowercases the ASCII letters in s, in the same way that the
// HTML tokenizer lowercases element and attribute names. strings.ToLower also
// lowercases other letters, e.g. the Kelvin sign is lowercased to "k".
func toLowerASCII(s string) string {
	var sb strings.Builder
	sb.Grow(len(s))
	for i := range len(s) {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		sb.WriteByte(c)
	}
	return sb.String()
}

// IsAnimationTargetURL returns true if the value of an SVG animation element's
// attributeName attribute may be an attribute that contains a URL.
//
// Browsers may ignore whitespace around the name, and a namespace prefix other
// than xlink may refer to the XLink namespace, so IsAnimationTargetURL trims
// HTML whitespace, and returns true for a href attribute with any prefix.
func IsAnimationTargetURL(attributeName string) bool {
	name := toLowerASCII(strings.Trim(attributeName, " \t\n\f\r"))
	if _, local, ok := strings.Cut(name, ":"); ok {
		name = local
	}
	return name == "href"
}

// IsValidName returns true if the name can be written as an HTML attribute
// name without changing the structure of the element. It rejects empty names,
// and names that contain whitespace, control characters, quotes, the
// characters that end an attribute name, or an ampersand, because browsers do
// not decode character references in attribute names.
//
// html/template's htmlNameFilter is stricter: it only allows ASCII letters and
// digits, so it rejects common names such as data-testid.
func IsValidName(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		if r <= 0x20 || (r >= 0x7f && r <= 0x9f) {
			return false
		}
		switch r {
		case '"', '\'', '>', '/', '=', '<', '`', '&':
			return false
		}
	}
	return true
}
