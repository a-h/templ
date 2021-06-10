package safehtml

import "testing"

func TestSanitizeCSS(t *testing.T) {
	var tests = []struct {
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

//TODO: Integrate tests from safehtml and Go's html/template.
//{
//"styleURLQueryEncoded",
//`<p style="background: url(/img?name={{"O'Reilly Animal(1)<2>.png"}})">`,
//`<p style="background: url(/img?name=O%27Reilly%20Animal%281%29%3c2%3e.png)">`,
//},
//{
//"styleQuotedURLQueryEncoded",
//`<p style="background: url('/img?name={{"O'Reilly Animal(1)<2>.png"}}')">`,
//`<p style="background: url('/img?name=O%27Reilly%20Animal%281%29%3c2%3e.png')">`,
//},
//{
//"styleStrQueryEncoded",
//`<p style="background: '/img?name={{"O'Reilly Animal(1)<2>.png"}}'">`,
//`<p style="background: '/img?name=O%27Reilly%20Animal%281%29%3c2%3e.png'">`,
//},
//{
//"styleURLBadProtocolBlocked",
//`<a style="background: url('{{"javascript:alert(1337)"}}')">`,
//`<a style="background: url('#ZgotmplZ')">`,
//},
//{
//"styleStrBadProtocolBlocked",
//`<a style="background: '{{"vbscript:alert(1337)"}}'">`,
//`<a style="background: '#ZgotmplZ'">`,
//},
//{
//"styleStrEncodedProtocolEncoded",
//`<a style="background: '{{"javascript\\3a alert(1337)"}}'">`,
//// The CSS string 'javascript\\3a alert(1337)' does not contain a colon.
//`<a style="background: 'javascript\\3a alert\28 1337\29 '">`,
//},
//{
//"styleURLGoodProtocolPassed",
//`<a style="background: url('{{"http://oreilly.com/O'Reilly Animals(1)<2>;{}.html"}}')">`,
//`<a style="background: url('http://oreilly.com/O%27Reilly%20Animals%281%29%3c2%3e;%7b%7d.html')">`,
//},
//{
//"styleStrGoodProtocolPassed",
//`<a style="background: '{{"http://oreilly.com/O'Reilly Animals(1)<2>;{}.html"}}'">`,
//`<a style="background: 'http\3a\2f\2foreilly.com\2fO\27Reilly Animals\28 1\29\3c 2\3e\3b\7b\7d.html'">`,
//},
//{
//"styleURLEncodedForHTMLInAttr",
//`<a style="background: url('{{"/search?img=foo&size=icon"}}')">`,
//`<a style="background: url('/search?img=foo&amp;size=icon')">`,
//},
//{
//"styleURLNotEncodedForHTMLInCdata",
//`<style>body { background: url('{{"/search?img=foo&size=icon"}}') }</style>`,
//`<style>body { background: url('/search?img=foo&size=icon') }</style>`,
//},
//{
//"styleURLMixedCase",
//`<p style="background: URL(#{{.H}})">`,
//`<p style="background: URL(#%3cHello%3e)">`,
//},
//{
//"stylePropertyPairPassed",
//`<a style='{{"color: red"}}'>`,
//`<a style='color: red'>`,
//},
//{
//"styleStrSpecialsEncoded",
//`<a style="font-family: '{{"/**/'\";:// \\"}}', &quot;{{"/**/'\";:// \\"}}&quot;">`,
//`<a style="font-family: '\2f**\2f\27\22\3b\3a\2f\2f  \\', &quot;\2f**\2f\27\22\3b\3a\2f\2f  \\&quot;">`,
//},
//{
//"styleURLSpecialsEncoded",
//`<a style="border-image: url({{"/**/'\";:// \\"}}), url(&quot;{{"/**/'\";:// \\"}}&quot;), url('{{"/**/'\";:// \\"}}'), 'http://www.example.com/?q={{"/**/'\";:// \\"}}''">`,
//`<a style="border-image: url(/**/%27%22;://%20%5c), url(&quot;/**/%27%22;://%20%5c&quot;), url('/**/%27%22;://%20%5c'), 'http://www.example.com/?q=%2f%2a%2a%2f%27%22%3b%3a%2f%2f%20%5c''">`,
//},
//{
//desc:  "angle brackets 1",
//input: `width: x<;`,
//want:  `contains angle brackets`,
//},
//{
//desc:  "angle brackets 2",
//input: `width: x>;`,
//want:  `contains angle brackets`,
//},
//{
//desc:  "angle brackets 3",
//input: `</style><script>alert('pwned')</script>`,
//want:  `contains angle brackets`,
//},
//{
//desc:  "no ending semicolon",
//input: `width: 1em`,
//want:  `must end with ';'`,
//},
//{
//desc:  "no colon",
//input: `width= 1em;`,
//want:  `must contain at least one ':' to specify a property-value pair`,
//},

//{
//desc: "BackgroundImageURLs single URL",
//input: StyleProperties{
//BackgroundImageURLs: []string{"http://goodUrl.com/a"},
//},
//want: `background-image:url("http://goodUrl.com/a");`,
//},
//{
//desc: "BackgroundImageURLs multiple URLs",
//input: StyleProperties{
//BackgroundImageURLs: []string{"http://goodUrl.com/a", "http://goodUrl.com/b"},
//},
//want: `background-image:url("http://goodUrl.com/a"), url("http://goodUrl.com/b");`,
//},
//{
//desc: "BackgroundImageURLs invalid runes in URL escaped",
//input: StyleProperties{
//BackgroundImageURLs: []string{"http://goodUrl.com/a\"\\\n"},
//},
//want: `background-image:url("http://goodUrl.com/a\000022\00005C\00000A");`,
//},
//{
//desc: "FontFamily unquoted names",
//input: StyleProperties{
//FontFamily: []string{"serif", "sans-serif", "GulimChe"},
//},
//want: `font-family:serif, sans-serif, GulimChe;`,
//},
//{
//desc: "FontFamily quoted names",
//input: StyleProperties{
//FontFamily: []string{"\nserif", "serif\n", "Goudy Bookletter 1911", "New Century Schoolbook", `"sans-serif"`},
//},
//want: `font-family:"\00000Aserif", "serif\00000A", "Goudy Bookletter 1911", "New Century Schoolbook", "sans-serif";`,
//},
//{
//desc: "FontFamily quoted and unquoted names",
//input: StyleProperties{
//FontFamily: []string{"sans-serif", "Goudy Bookletter 1911", "GulimChe", `"fantasy"`, "Times New Roman"},
//},
//want: `font-family:sans-serif, "Goudy Bookletter 1911", GulimChe, "fantasy", "Times New Roman";`,
//},
//{
//desc: "Display",
//input: StyleProperties{
//Display: "inline",
//},
//want: "display:inline;",
//},
//{
//desc: "BackgroundColor",
//input: StyleProperties{
//BackgroundColor: "red",
//},
//want: "background-color:red;",
//},
//{
//desc: "BackgroundPosition",
//input: StyleProperties{
//BackgroundPosition: "100px -110px",
//},
//want: "background-position:100px -110px;",
//},
//{
//desc: "BackgroundRepeat",
//input: StyleProperties{
//BackgroundRepeat: "no-repeat",
//},
//want: "background-repeat:no-repeat;",
//},
//{
//desc: "BackgroundSize",
//input: StyleProperties{
//BackgroundSize: "10px",
//},
//want: "background-size:10px;",
//},
//{
//desc: "Color",
//input: StyleProperties{
//Color: "#000",
//},
//want: "color:#000;",
//},
//{
//desc: "Height",
//input: StyleProperties{
//Height: "100px",
//},
//want: "height:100px;",
//},
//{
//desc: "Width",
//input: StyleProperties{
//Width: "120px",
//},
//want: "width:120px;",
//},
//{
//desc: "Left",
//input: StyleProperties{
//Left: "140px",
//},
//want: "left:140px;",
//},
//{
//desc: "Right",
//input: StyleProperties{
//Right: "160px",
//},
//want: "right:160px;",
//},
//{
//desc: "Top",
//input: StyleProperties{
//Top: "180px",
//},
//want: "top:180px;",
//},
//{
//desc: "Bottom",
//input: StyleProperties{
//Bottom: "200px",
//},
//want: "bottom:200px;",
//},
//{
//desc: "FontWeight",
//input: StyleProperties{
//FontWeight: "100",
//},
//want: "font-weight:100;",
//},
//{
//desc: "Padding",
//input: StyleProperties{
//Padding: "5px 1em 0 2em",
//},
//want: "padding:5px 1em 0 2em;",
//},
//{
//desc: "ZIndex",
//input: StyleProperties{
//ZIndex: "-2",
//},
//want: "z-index:-2;",
//},
//{
//desc: "multiple properties",
//input: StyleProperties{
//BackgroundImageURLs: []string{"http://goodUrl.com/a", "http://goodUrl.com/b"},
//FontFamily:          []string{"serif", "Goudy Bookletter 1911", "Times New Roman", "monospace"},
//BackgroundColor:     "#bbff10",
//BackgroundPosition:  "100px -110px",
//BackgroundRepeat:    "no-repeat",
//BackgroundSize:      "10px",
//Width:               "12px",
//Height:              "10px",
//},
//want: `background-image:url("http://goodUrl.com/a"), url("http://goodUrl.com/b");` +
//`font-family:serif, "Goudy Bookletter 1911", "Times New Roman", monospace;` +
//`background-color:#bbff10;` +
//`background-position:100px -110px;` +
//`background-repeat:no-repeat;` +
//`background-size:10px;` +
//`height:10px;` +
//`width:12px;`,
//},
//{
//desc: "multiple properties, some empty and unset",
//input: StyleProperties{
//BackgroundImageURLs: []string{"http://goodUrl.com/a", "http://goodUrl.com/b"},
//BackgroundPosition:  "100px -110px",
//BackgroundSize:      "",
//Width:               "12px",
//Height:              "10px",
//},
//want: `background-image:url("http://goodUrl.com/a"), url("http://goodUrl.com/b");` +
//`background-position:100px -110px;` +
//`height:10px;` +
//`width:12px;`,
//},
//{
//desc:  "no properties set",
//input: StyleProperties{},
//want:  "",
//},
//{
//desc: "sanitize comment in regular value",
//input: StyleProperties{
//BackgroundRepeat:   "// This is bad",
//BackgroundPosition: "/* This is bad",
//BackgroundSize:     "This is bad */",
//},
//want: "background-position:zGoSafezInvalidPropertyValue;" +
//`background-repeat:zGoSafezInvalidPropertyValue;` +
//`background-size:zGoSafezInvalidPropertyValue;`,
//},
//{
//desc: "sanitize comment in middle of regular value",
//input: StyleProperties{
//BackgroundRepeat:   "10px /* This is bad",
//BackgroundPosition: "10px // This is bad",
//BackgroundSize:     "10px */ This is bad",
//},
//want: "background-position:zGoSafezInvalidPropertyValue;" +
//`background-repeat:zGoSafezInvalidPropertyValue;` +
//`background-size:zGoSafezInvalidPropertyValue;`,
//},
//{
//desc: "sanitize bad rune in regular value",
//input: StyleProperties{
//BackgroundSize: "This&is$bad",
//},
//want: "background-size:zGoSafezInvalidPropertyValue;",
//},
//{
//desc: "sanitize invalid enum value",
//input: StyleProperties{
//Display: "badValue123",
//},
//want: "display:zGoSafezInvalidPropertyValue;",
//},
//{
//desc: "sanitize unsafe URL value",
//input: StyleProperties{
//BackgroundImageURLs: []string{"javascript:badJavascript();"},
//},
//want: `background-image:url("about:invalid#zGoSafez");`,
//},
//{
//desc: "sanitize regular and enum properties with newline prefix",
//input: StyleProperties{
//Display:         "\nfoo",
//BackgroundColor: "\nfoo",
//},
//want: "display:zGoSafezInvalidPropertyValue;background-color:zGoSafezInvalidPropertyValue;",
//},
//{
//desc: "sanitize regular and enum properties with newline suffix",
//input: StyleProperties{
//Display:         "foo\n",
//BackgroundColor: "foo\n",
//},
//want: "display:zGoSafezInvalidPropertyValue;background-color:zGoSafezInvalidPropertyValue;",
//},
//{
//desc: "regular value symbols in value",
//input: StyleProperties{
//BackgroundSize: "*+/-.!#%_ \t",
//},
//want: "background-size:*+/-.!#%_ \t;",
//},
//{
//desc: "quoted and unquoted font family names CSS-escaped",
//input: StyleProperties{
//FontFamily: []string{
//`"`,
//`""`,
//`serif\`,
//`"Gulim\Che"`,
//`"Gulim"Che"`,
//`New Century Schoolbook"`,
//`"New Century Schoolbook`,
//`New Century" Schoolbook`,
//`sans-"serif`,
//},
//},
//want: `font-family:"\000022", ` +
//`"\000022\000022", ` +
//`"serif\00005C", ` +
//`"Gulim\00005CChe", ` +
//`"Gulim\000022Che", ` +
//`"New Century Schoolbook\000022", ` +
//`"\000022New Century Schoolbook", ` +
//`"New Century\000022 Schoolbook", ` +
//`"sans-\000022serif";`,
//},
//{
//desc: "less-than rune CSS-escaped",
//input: StyleProperties{
//BackgroundImageURLs: []string{`</style><script>evil()</script>`},
//FontFamily:          []string{`</style><script>evil()</script>`},
//},
//want: `background-image:url("\00003C/style>\00003Cscript>evil()\00003C/script>");` +
//`font-family:"\00003C/style>\00003Cscript>evil()\00003C/script>";`,
//},V
