# Parse

A set of parsing tools for Go inspired by [Sprache](https://github.com/sprache/Sprache/).

Build up complex parsers from small, simple functions that chomp away at the input.

## Input

The input moves along as the parser succeeds.

```go
input := parse.NewInput("ABCD")
item, ok, err := parse.String("A").Parse(input)
// Input is now at index 1.
item, ok, err := parse.String("B").Parse(input)
// Input is now at index 2.
item, ok, err := parse.String("XYZ").Parse(input)
// Input index didn't change and ok=false.
item, ok, err := parse.String("CD").Parse(input)
// Input is now at index 4.
```

## Design

A parser must match the `parse.Parser` interface, or be created by the use of the `parser.Func` helper. These 3 parsers are equivalent.

```go
parse.String("<")
```

```go
parse.Func(func(in *parse.Input) (item string, ok bool, err error) {
	item, _ = in.Peek(1)
	ok = item == "<"
	return
})
```

```go
type lessThanParser struct{}

func (ltp lessThanParser) Parse(in *parse.Input) (item string, ok bool, err error) {
	item, _ = in.Peek(1)
	ok = item == "<"
	return
}
```

## Functions

Parser functions provide a way of matching patterns in a given input. They are designed to be able to be composed together to make more complex operations.

The [examples](./examples) directory contains several examples of composing the primitive functions.

* `Any`
    * Parse any of the provided parse functions, or roll back.
* `AnyRune`
    * Parse any rune.
* `AtLeast`
    * Parse the provided function at least the number of times specified, or roll back.
* `AtMost`
    * Parse the provided function at least once, and at most the number of times specified, or roll back.
* `Letter`
    * Parse any letter in the Unicode Letter range or roll back.
* `Many`
    * Parse the provided parse function a number of times or roll back.
* `Optional`
    * Attempt to parse, but don't roll back if a match isn't found.
* `Or`
    * Return the first successful result of the provided parse functions, or roll back.
* `Rune`
    * Parse the specified rune (character) or fallback.
* `RuneIn`
    * Parse a rune from the input stream if it's in the specified string, or roll back.
* `RuneInRanges`
    * Parse a rune from the input stream if it's in the specified Unicode ranges, or roll back.
* `RuneNotIn`
    * Parse a rune from the input stream if it's not in the specified string, or roll back.
* `RuneWhere`
    * Parse a rune from the input stream if the predicate function passed in succeeds, or roll back.
* `String`
    * Parse a string from the input stream if it exactly matches the provided string, or roll back.
* `StringUntil`
    * Parse a string from the input stream until the specified _until_ parser is matched.
* `Then`
    * Return the results of the first and second parser passed through the combiner function which converts the two results into a single output (a map / reduce operation), or roll back if either doesn't match.
* `Times`
    * Parse using the specified function a set number of times or roll back.
* `Until`
    * Parse from the input stream until the specified _until_ parser is matched.
* `ZeroToNine`
    * Parse a rune from the input stream if it's within the set of 1234567890.

## More complex parsers

More complex parsers will need to store the start position, and rollback by using the input's `Seek` function if the current parser does not match the input.

```go
func ExampleParser() {
	type GotoStatement struct {
		Line int64
	}
	gotoParser := parse.Func(func(in *parse.Input) (item GotoStatement, ok bool, err error) {
		start := in.Index()

		if _, ok, err = parse.String("GOTO ").Parse(in); err != nil || !ok {
			// Rollback, and return.
			in.Seek(start)
			return
		}

		// Read until the next newline or the EOF.
		until := parse.Any(parse.NewLine, parse.EOF[string]())
		var lineNumber string
		if lineNumber, ok, err = parse.StringUntil(until).Parse(in); err != nil || !ok {
			err = parse.Error("Syntax error: GOTO is missing line number", in.Position())
			return
		}
		// We must have a valid line number now, or there is a syntax error.
		item.Line, err = strconv.ParseInt(lineNumber, 10, 64)
		if err != nil {
			return item, false, parse.Error("Syntax error: GOTO has invalid line number", in.Position())
		}

		// Chomp the newline we read up to.
		until.Parse(in)

		return item, true, nil
	})

	inputs := []string{
		"GOTO 10",
		"GOTO abc",
		"FOR i = 0",
	}
	for _, input := range inputs {
		stmt, ok, err := gotoParser.Parse(parse.NewInput(input))
		fmt.Printf("%+v, ok=%v, err=%v\n", stmt, ok, err)
	}
	// Output:
	// {Line:10}, ok=true, err=<nil>
	// {Line:0}, ok=false, err=Syntax error: GOTO has invalid line number: line 0, col 8
	// {Line:0}, ok=false, err=<nil>
}
```
