package annotationui

import "embed"

//go:embed queries/*/*.sql
var embeddedQueries embed.FS
