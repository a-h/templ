
# browser

Helpers to open URLs, readers, or files in the system default web browser.

This fork adds:

- `OpenReader` error wrapping;
- `ErrNotFound` error wrapping on BSD;
- Go 1.21 support.

## Usage

``` go
import "github.com/cli/browser"

err = browser.OpenURL(url)
err = browser.OpenFile(path)
err = browser.OpenReader(reader)
```
