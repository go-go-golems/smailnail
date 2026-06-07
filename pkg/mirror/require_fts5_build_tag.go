//go:build !sqlite_fts5 && !fts5

package mirror

// requires_sqlite_fts5_build_tag is a sentinel value that forces a compile
// error when this file is included without the corresponding sqlite_fts5
// or fts5 build tag. The variable is defined in the tagged build file.
var requires_sqlite_fts5_build_tag = true
