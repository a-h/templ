package render

import (
	"encoding/json"
	"io/fs"
	"path"
	"path/filepath"
	"strings"
)

type SectionPage struct {
	Fsys     fs.FS
	Path     string
	Children []MarkdownPage
}

func (p SectionPage) GetPath() string {
	return p.Path
}

func (p SectionPage) Title() string {
	b, err := fs.ReadFile(p.Fsys, filepath.Join(p.Path, "_category_.json"))
	if err != nil {
		return titleFromPath(p.Path)
	}

	var category struct {
		Label string `json:"label"`
	}

	if err := json.Unmarshal(b, &category); err != nil {
		return titleFromPath(p.Path)
	}

	if category.Label != "" {
		return category.Label
	}

	return titleFromPath(p.Path)
}

func (p SectionPage) Href() string {
	if p.Path == "index.md" {
		return "index"
	}

	noExt := strings.TrimSuffix(p.Path, filepath.Ext(p.Path))

	htmlPath := ""

	for _, r := range strings.Split(noExt, "/") {
		name, _ := baseParts(r)
		htmlPath = path.Join(htmlPath, name)
	}

	for _, r := range []string{"\\", " ", ".", "_"} {
		htmlPath = strings.ReplaceAll(htmlPath, r, "-")
	}

	htmlPath = strings.ToLower(htmlPath)
	htmlPath = strings.Trim(htmlPath, "-")

	return htmlPath
}
