package render

type Page struct {
	Path     string
	Type     int
	Title    string
	Slug     string
	Href     string
	Html     string
	Order    int
	Children []*Page
}

const (
	PageMarkdown = iota
	PageSection
)

type PageContext struct {
	Title  string
	Active string
}
