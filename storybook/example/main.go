package main

import (
	"context"
	"fmt"
	"os"

	"github.com/a-h/templ/storybook"
)

func main() {
	s := storybook.New()
	s.AddComponent("headerTemplate", headerTemplate,
		storybook.TextArg("name", "Page Name"))
	s.AddComponent("footerTemplate", footerTemplate)
	if err := s.ListenAndServeWithContext(context.Background()); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
