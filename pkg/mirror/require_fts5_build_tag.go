//go:build !sqlite_fts5 && !fts5

package mirror

// This undefined sentinel intentionally forces a compile error when this file
// is included without the corresponding sqlite_fts5 or fts5 build tag.
var _ = requires_sqlite_fts5_build_tag
