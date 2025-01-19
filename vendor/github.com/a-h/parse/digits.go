package parse

// ZeroToNine matches a single rune from 0123456789.
var ZeroToNine Parser[string] = RuneIn("0123456789")
