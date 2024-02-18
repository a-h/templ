package render

type Page struct {
	Path            string
	Type            int
	Title           string
	Slug            string
	Href            string
	RawContent      string
	RenderedContent string
	Order           int
	Children        []*Page
}

const (
	PageMarkdown = iota
	PageSection
)

type PageContext struct {
	Title  string
	Active string
}
