package render

type Page struct {
	Path            string
	Type            PageType
	Title           string
	Slug            string
	Href            string
	RawContent      string
	RenderedContent string
	Order           int
	Children        []*Page
}

type PageType int

const (
	PageMarkdown PageType = iota
	PageSection  PageType = iota
)

type PageContext struct {
	Title  string
	Active string
}
