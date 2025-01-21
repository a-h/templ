package safehtml

import "testing"

func TestSanitizeCSS(t *testing.T) {
	tests := []struct {
		name             string
		inputProperty    string
		expectedProperty string
		inputValue       string
		expectedValue    string
	}{
		{
			name:             "directions are allowed",
			inputProperty:    "dir",
			expectedProperty: "dir",
			inputValue:       "ltr",
			expectedValue:    "ltr",
		},
		{
			name:             "border-left allowed",
			inputProperty:    "border-left",
			expectedProperty: "border-left",
			inputValue:       "0",
			expectedValue:    "0",
		},
		{
			name:             "border can contain multiple values",
			inputProperty:    "border",
			expectedProperty: "border",
			inputValue:       `1 1 1 1`,
			expectedValue:    `1 1 1 1`,
		},
		{
			name:             "properties are case corrected",
			inputProperty:    "Border",
			expectedProperty: "border",
			inputValue:       `1 1 1 1`,
			expectedValue:    `1 1 1 1`,
		},
		{
			name:             "expressions are not allowed",
			inputProperty:    "width",
			expectedProperty: "width",
			inputValue:       `expression(alert(1337))`,
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "font-family standard values are allowed",
			inputProperty:    "font-family",
			expectedProperty: "font-family",
			inputValue:       `sans-serif`,
			expectedValue:    `sans-serif`,
		},
		{
			name:             "font-family values with spaces are allowed",
			inputProperty:    "font-family",
			expectedProperty: "font-family",
			inputValue:       `Akzidenz Grotesk`,
			expectedValue:    `Akzidenz Grotesk`,
		},
		{
			name:             "font-family multiple standard values are allowed",
			inputProperty:    "font-family",
			expectedProperty: "font-family",
			inputValue:       `sans-serif, monospaced`,
			expectedValue:    `sans-serif, monospaced`,
		},
		{
			name:             "font-family multiple quoted and non-quoted values are allowed",
			inputProperty:    "font-family",
			expectedProperty: "font-family",
			inputValue:       `"Georgia", monospaced, sans-serif`,
			expectedValue:    `"Georgia", monospaced, sans-serif`,
		},
		{
			name:             "font-family Chinese names are allowed",
			inputProperty:    "font-family",
			expectedProperty: "font-family",
			inputValue:       `"中易宋体", monospaced`,
			expectedValue:    `"中易宋体", monospaced`,
		},
		{
			name:             "font-family quoted values must be terminated",
			inputProperty:    "font-family",
			expectedProperty: "font-family",
			inputValue:       `"quotes`,
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "font-family non standard names are not allowed",
			inputProperty:    "font-family",
			expectedProperty: "font-family",
			inputValue:       `foo@bar`,
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "obfuscated values are not allowed",
			inputProperty:    "width",
			expectedProperty: "width",
			inputValue:       `  e\\78preS\x00Sio/**/n(alert(1337))`,
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "moz binding blocked",
			inputProperty:    "-moz-binding(alert(1337))",
			expectedProperty: InnocuousPropertyName,
			inputValue:       `something`,
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "obfuscated moz-binding blocked",
			inputProperty:    "  -mo\\7a-B\x00I/**/nding(alert(1337))",
			expectedProperty: InnocuousPropertyName,
			inputValue:       `something`,
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "angle brackets in property value",
			inputProperty:    "background-image",
			expectedProperty: "background-image",
			inputValue:       `url(/img?name=O'Reilly Animal(1)<2>.png)`,
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "angle brackets in quoted property value",
			inputProperty:    "background-image",
			expectedProperty: "background-image",
			inputValue:       `url("/img?name=O'Reilly Animal(1)<2>.png")`,
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "background",
			inputProperty:    "background",
			expectedProperty: "background",
			inputValue:       "url(/img?name=O%27Reilly%20Animal%281%29%3c2%3e.png)",
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "background-image JS URL",
			inputProperty:    "background-image",
			expectedProperty: "background-image",
			inputValue:       `url(javascript:alert(1337))`,
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "background-image VBScript URL",
			inputProperty:    "background-image",
			expectedProperty: "background-image",
			inputValue:       `url(vbscript:alert(1337))`,
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "background-image absolute FTP URL",
			inputProperty:    "background-image",
			expectedProperty: "background-image",
			inputValue:       `url("ftp://safe.example.com/img.png")`,
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "background-image invalid URL",
			inputProperty:    "background-image",
			expectedProperty: "background-image",
			inputValue:       `url("` + string([]byte{0x7f}) + `")`,
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "background-image invalid prefix",
			inputProperty:    "background-image",
			expectedProperty: "background-image",
			inputValue:       `/img.png")`,
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "background-image invalid suffix",
			inputProperty:    "background-image",
			expectedProperty: "background-image",
			inputValue:       `url("/img.png`,
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "background-image safe URL",
			inputProperty:    "background-image",
			expectedProperty: "background-image",
			inputValue:       `url("/img.png")`,
			expectedValue:    `url("/img.png")`,
		},
		{
			name:             "background-image safe URL - two slashes",
			inputProperty:    "background-image",
			expectedProperty: "background-image",
			inputValue:       `url("//img.png")`,
			expectedValue:    `url("//img.png")`,
		},
		{
			name:             "background-image safe HTTP URL",
			inputProperty:    "background-image",
			expectedProperty: "background-image",
			inputValue:       `url("http://safe.example.com/img.png")`,
			expectedValue:    `url("http://safe.example.com/img.png")`,
		},
		{
			name:             "background-image safe mailto URL",
			inputProperty:    "background-image",
			expectedProperty: "background-image",
			inputValue:       `url("mailto:foo@bar.foo")`,
			expectedValue:    `url("mailto:foo@bar.foo")`,
		},
		{
			name:             "background-image multiple URLs",
			inputProperty:    "background-image",
			expectedProperty: "background-image",
			inputValue:       `url("http://safe.example.com/img.png"), url("https://safe.example.com/other.png")`,
			expectedValue:    `url("http://safe.example.com/img.png"), url("https://safe.example.com/other.png")`,
		},
		{
			name:             "-webkit-text-stroke-color safe webkit",
			inputProperty:    "-webkit-text-stroke-color",
			expectedProperty: "-webkit-text-stroke-color",
			inputValue:       `#000`,
			expectedValue:    `#000`,
		},
		{
			name:             "escape attempt property name",
			inputProperty:    "</style><script>alert('hello')</script><style>",
			expectedProperty: InnocuousPropertyName,
			inputValue:       "test",
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "escape attempt property value",
			inputProperty:    "bottom",
			expectedProperty: "bottom",
			inputValue:       "</style><script>alert('hello')</script><style>",
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "encoded protocol is blocked",
			inputProperty:    "bottom",
			expectedProperty: "bottom",
			inputValue:       "javascript\\3a alert(1337)",
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "angle brackets 1",
			inputProperty:    "bottom",
			expectedProperty: "bottom",
			inputValue:       "x<",
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "angle brackets 2",
			inputProperty:    "bottom",
			expectedProperty: "bottom",
			inputValue:       "x>",
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "contains colon",
			inputProperty:    "bottom",
			expectedProperty: "bottom",
			inputValue:       ":",
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "contains semicolon",
			inputProperty:    "bottom",
			expectedProperty: "bottom",
			inputValue:       ";",
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "comments are removed #1",
			inputProperty:    "background-repeat",
			expectedProperty: "background-repeat",
			inputValue:       "// removed",
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "comments are removed #2",
			inputProperty:    "background-repeat",
			expectedProperty: "background-repeat",
			inputValue:       "/* removed",
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "comments are removed #3",
			inputProperty:    "background-repeat",
			expectedProperty: "background-repeat",
			inputValue:       "/* removed",
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "comments are sanitized mid value #1",
			inputProperty:    "background-repeat",
			expectedProperty: "background-repeat",
			inputValue:       "repeat-none // removed",
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "comments are sanitized mid value #2",
			inputProperty:    "background-repeat",
			expectedProperty: "background-repeat",
			inputValue:       "repeat-none /* removed",
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "comments are sanitized mid value #3",
			inputProperty:    "background-repeat",
			expectedProperty: "background-repeat",
			inputValue:       "repeat-none /* removed",
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "bad characters in value",
			inputProperty:    "background-repeat",
			expectedProperty: "background-repeat",
			inputValue:       "This&is$bad",
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "enum values must be valid",
			inputProperty:    "display",
			expectedProperty: "display",
			inputValue:       "badValue123",
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "values cannot contain newlines at start",
			inputProperty:    "display",
			expectedProperty: "display",
			inputValue:       "\ntest",
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "values cannot contain newlines at end",
			inputProperty:    "display",
			expectedProperty: "display",
			inputValue:       "test\n",
			expectedValue:    InnocuousPropertyValue,
		},
		{
			name:             "some symbols are allowed",
			inputProperty:    "display",
			expectedProperty: "display",
			inputValue:       "*+/-.!#%_ \t",
			expectedValue:    InnocuousPropertyValue,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			actualProperty, actualValue := SanitizeCSS(tt.inputProperty, tt.inputValue)
			if actualProperty != tt.expectedProperty {
				t.Errorf("%s: mismatched property - expected %q, actual %q", tt.name, tt.expectedProperty, actualProperty)
			}
			if actualValue != tt.expectedValue {
				t.Errorf("%s: mismatched value - expected %q, actual %q", tt.name, tt.expectedValue, actualValue)
			}
		})
	}
}
