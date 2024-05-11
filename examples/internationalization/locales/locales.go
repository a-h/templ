package locales

import "embed"

//go:embed en
//go:embed de
//go:embed zh-cn

var Content embed.FS
