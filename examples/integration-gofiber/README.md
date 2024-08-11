## Example

This example demonstrates the usage of templ with gofiber.

As soon as you start the server you can access http://localhost:3000/ and see the rendered page.

If you change the URL to http://localhost:3000/john you will see your parameter printed on the page.

This happens both through parameter passing into the templ component and through context using fiber locals.

## Tasks

### build-templ

```
templ generate
```

### run

```
go run .
```

