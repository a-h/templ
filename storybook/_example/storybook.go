package example

import (
	"github.com/a-h/templ/storybook"
)

func Storybook() *storybook.Storybook {
	s := storybook.New()

	header := s.AddComponent("headerTemplate", headerTemplate, storybook.TextArg("name", "Page Name"))
	header.AddStory("Long Name", storybook.TextArg("name", "A Very Long Page Name"))

	s.AddComponent("footerTemplate", footerTemplate)

	return s
}
