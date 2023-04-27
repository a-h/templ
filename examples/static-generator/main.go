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

type Post struct {
	Date    time.Time
	Title   string
	Content string
}

func Unsafe(html string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		_, err = io.WriteString(w, html)
		return
	})
}

func main() {
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
* Enjoying the day off as a public holiday, known as Early May Bank Holiday
`,
		},
	}

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
