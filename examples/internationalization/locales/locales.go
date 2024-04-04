// Package locales provides content for translations.
package locales

import "embed"

//go:embed en
//go:embed de
//go:embed es
//go:embed fr
//go:embed it

var Content embed.FS
