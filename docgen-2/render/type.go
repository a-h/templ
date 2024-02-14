package render

type Renderer interface {
	GetPath() string
	Title() string
	Href() string
	Render() (string, error)
}
