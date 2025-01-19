# htmlformat

htmlformat is a Go package and CLI tool used to format HTML.

It is forked and simplified from the https://github.com/ericchiang/pup package.

It does not aim to:

* Colorize output.
* Modify the input HTML except for formatting (i.e. no HTML escaping will be applied).
* Provide any facilities to query the content.

## Installation

To use the CLI, you can install with Go > 1.20.

```
go install github.com/a-h/htmlformat/cmd/htmlformat@latest
```

## Usage

### CLI

```bash
echo '<ol><li style="&">A</li><li>B</li></ol>' | htmlformat
<ol>
 <li style="&">
  A
 </li>
 <li>
  B
 </li>
</ol>
```

### Package

```go
r := strings.NewReader(`<ol><li style="&">A</li><li>B</li></ol>`)
w := os.Stdout
if err := Fragment(w, r); err != nil {
  log.Fatalf("failed to format: %v", err)
}
```
