# Blog example

This example demonstrates building a static blog with templ.

## Create a blog template

Create a template for the site header and site content. Then, create a template for the content page and index page.

```templ title="blog.templ"
package main

import "path"
import "github.com/gosimple/slug"

templ headerComponent(title string) {
	<head><title>{ title }</title></head>
}

templ contentComponent(title string, body templ.Component) {
	<body>
		<h1>{ title }</h1>
		<div class="content">
			@body
		</div>
	</body>
}

templ contentPage(title string, body templ.Component) {
	<html>
		@headerComponent(title)
		@contentComponent(title, body)
	</html>
}

templ indexPage(posts []Post) {
	<html>
		@headerComponent("My Blog")
		<body>
			<h1>My Blog</h1>
			for _, post := range posts {
				<div><a href={ templ.SafeURL(path.Join(post.Date.Format("2006/01/02"), slug.Make(post.Title), "/")) }>{ post.Title }</a></div>
			}
		</body>
	</html>
}
```

In the Go code, create a `Post` struct to store information about a blog post.

```go
type Post struct {
	Date    time.Time
	Title   string
	Content string
}
```

Create some pretend blog posts.

```go
posts := []Post{
	{
		Date:  time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
		Title: "Happy New Year!",
		Content: `New Year is a widely celebrated occasion in the United Kingdom, marking the end of one year and the beginning of another.

Top New Year Activities in the UK include:

* Attending a Hogmanay celebration in Scotland
* Taking part in a local First-Foot tradition in Scotland and Northern England
* Setting personal resolutions and goals for the upcoming year
* Going for a New Year's Day walk to enjoy the fresh start
* Visiting a local pub for a celebratory toast and some cheer
`,
	},
	{
		Date:  time.Date(2023, time.May, 1, 0, 0, 0, 0, time.UTC),
		Title: "May Day",
		Content: `May Day is an ancient spring festival celebrated on the first of May in the United Kingdom, embracing the arrival of warmer weather and the renewal of life.

Top May Day Activities in the UK:

* Dancing around the Maypole, a traditional folk activity
* Attending local village fetes and fairs
* Watching or participating in Morris dancing performances
* Enjoying the public holiday known as Early May Bank Holiday
`,
	},
}
```

## Rendering HTML directly

The example blog posts contain markdown, so we'll use `github.com/yuin/goldmark` to convert the markdown to HTML.

We can't use a string containing HTML directly in templ, because all strings are escaped in templ. So we'll create an `Unsafe` code component to write the HTML directly to the output writer without first escaping it.

```go
func Unsafe(html string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		_, err = io.WriteString(w, html)
		return
	})
}
```

## Creating the static pages

The code creates the index page. The code then iterates through the posts, creating an output file for each blog post.

```go title="main.go"
package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"path"
	"time"

	"github.com/a-h/templ"
	"github.com/gosimple/slug"
	"github.com/yuin/goldmark"
)

func main() {
	// Output path.
	rootPath := "public"
	if err := os.Mkdir(rootPath, 0755); err != nil {
		log.Fatalf("failed to create output directory: %v", err)
	}

	// Create an index page.
	name := path.Join(rootPath, "index.html")
	f, err := os.Create(name)
	if err != nil {
		log.Fatalf("failed to create output file: %v", err)
	}
	// Write it out.
	err = indexPage(posts).Render(context.Background(), f)
	if err != nil {
		log.Fatalf("failed to write index page: %v", err)
	}

	// Create a page for each post.
	for _, post := range posts {
		// Create the output directory.
		dir := path.Join(rootPath, post.Date.Format("2006/01/02"), slug.Make(post.Title))
		if err := os.MkdirAll(dir, 0755); err != nil && err != os.ErrExist {
			log.Fatalf("failed to create dir %q: %v", dir, err)
		}

		// Create the output file.
		name := path.Join(dir, "index.html")
		f, err := os.Create(name)
		if err != nil {
			log.Fatalf("failed to create output file: %v", err)
		}

		// Convert the markdown to HTML, and pass it to the template.
		var buf bytes.Buffer
		if err := goldmark.Convert([]byte(post.Content), &buf); err != nil {
			log.Fatalf("failed to convert markdown to HTML: %v", err)
		}

		// Create an unsafe component containing raw HTML.
		content := Unsafe(buf.String())

		// Use templ to render the template containing the raw HTML.
		err = contentPage(post.Title, content).Render(context.Background(), f)
		if err != nil {
			log.Fatalf("failed to write output file: %v", err)
		}
	}
}
```

## Results

After generating Go code from the templates, and running it with `templ generate` followed by `go run *.go`, the following files will be created.

```
public/index.html
public/2023/01/01/happy-new-year/index.html
public/2023/05/01/may-day/index.html
```

The `index.html` contains links to all of the posts.

```html title="index.html"
<title>
 My Website
</title>
<h1>
 My Website
</h1>
<div>
 <a href="2023/01/01/happy-new-year/">
  Happy New Year!
 </a>
</div>
<div>
 <a href="2023/05/01/may-day/">
  May Day
 </a>
</div>
```

While each content page contains the HTML generated from the markdown, and the surrounding template.

```html title="2023/05/01/may-day/index.html"
<title>
 May Day
</title>
<h1>
 May Day
</h1>
<div class="content">
 <p>
  May Day is an ancient spring festival celebrated on the first of May in the United Kingdom, embracing the arrival of warmer weather and the renewal of life.
 </p>
 <p>
  Top May Day Activities in the UK:
 </p>
 <ul>
  <li>
   Dancing around the Maypole, a traditional folk activity
  </li>
  <li>
   Attending local village fetes and fairs
  </li>
  <li>
   Watching or participating in Morris dancing performances
  </li>
  <li>
   Enjoying the public holiday known as Early May Bank Holiday
  </li>
 </ul>
</div>
```

The files in the `public` directory can then be hosted in any static website hosting provider.
