package example

import (
	"github.com/a-h/templ/storybook"
)

func Storybook() *storybook.Storybook {
	s := storybook.New()
	s.AddComponent("headerTemplate", headerTemplate, storybook.TextArg("name", "Page Name"))
	s.AddComponent("footerTemplate", footerTemplate)
	return s
}
