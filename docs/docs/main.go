package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/gosimple/slug"
)

type Section struct {
	Name     string
	SubItems []string
}

type ItemToCreate struct {
	Name    string
	Path    string
	IsFile  bool
	Content string
}

func main() {
	sections := []Section{
		{Name: "Quick Start", SubItems: []string{"Installation", "Creating a simple templ component", "Running your first templ application"}},
		{Name: "Syntax and Usage", SubItems: []string{"Basic syntax", "Expressions", "Conditional HTML attribute expressions", "Loops", "Template composition", "CSS style management"}},
		{Name: "Core Concepts", SubItems: []string{"Components", "Template generation", "Conditional rendering", "Rendering lists", "Code-only components"}},
		{Name: "Components", SubItems: []string{"Creating and organizing components", "Adding HTML markup and Go code in templ", "Configuring components with parameters"}},
		{Name: "Using Go Functions and Variables", SubItems: []string{}},
		{Name: "Server-side Rendering", SubItems: []string{"Creating an HTTP server with templ", "Example: Counter application"}},
		{Name: "Static Rendering", SubItems: []string{"Generating static HTML files with templ", "Deploying static files"}},
		{Name: "Hosting and Deployment", SubItems: []string{"Hosting on AWS Lambda", "Hosting using Docker"}},
		{Name: "Commands and Tools", SubItems: []string{"templ generate", "templ fmt", "templ lsp"}},
		{Name: "Advanced Topics", SubItems: []string{"Code-only components", "Source maps", "Storybook integration"}},
		{Name: "Tutorials and Examples", SubItems: []string{"Tutorial: Counter application", "Tutorial: Blog application"}},
		{Name: "API Reference", SubItems: []string{"templ.Component", "templ.Handler"}},
		{Name: "Frequently Asked Questions", SubItems: []string{}},
		{Name: "Contributing and Support", SubItems: []string{}},
		{Name: "Best Practices", SubItems: []string{"Keeping templ components pure and avoiding bugs"}},
		{Name: "Conclusion", SubItems: []string{"Summary and next steps for learning more about templ"}},
	}

	var items []ItemToCreate

	for i, section := range sections {
		i += 1
		current := ItemToCreate{
			Name:   section.Name,
			Path:   fmt.Sprintf("%02d-%s", i+1, slug.Make(section.Name)),
			IsFile: false,
		}
		items = append(items, current)

		// Add category.json
		items = append(items, ItemToCreate{
			Path:    fmt.Sprintf("%s/_category_.json", current.Path),
			IsFile:  true,
			Content: categoryJSON(i+1, section.Name),
		})

		if len(section.SubItems) == 0 {
			fileItem := ItemToCreate{
				Name:    section.Name,
				Path:    fmt.Sprintf("%s/index.md", current.Path),
				IsFile:  true,
				Content: fmt.Sprintf("# %s\n", section.Name),
			}
			items = append(items, fileItem)
		}
		for j, subItem := range section.SubItems {
			fileItem := ItemToCreate{
				Name:    subItem,
				Path:    fmt.Sprintf("%s/%02d-%s.md", current.Path, j+1, slug.Make(subItem)),
				IsFile:  true,
				Content: fmt.Sprintf("# %s\n", subItem),
			}
			items = append(items, fileItem)
		}
	}

	for _, item := range items {
		fmt.Println(item.Path)
		if item.IsFile {
			os.WriteFile(item.Path, []byte(item.Content), 0644)
			continue
		}
		os.Mkdir(item.Path, 0755)
	}
}

func categoryJSON(position int, label string) string {
	sw := new(strings.Builder)
	enc := json.NewEncoder(sw)
	enc.SetIndent("", "  ")
	err := enc.Encode(Category{
		Position: position,
		Label:    label,
	})
	if err != nil {
		panic("failed to create JSON: " + err.Error())
	}
	return sw.String()
}

type Category struct {
	Position int    `json:"position"`
	Label    string `json:"label"`
}
