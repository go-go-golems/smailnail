//go:build !sqlite_fts5 && !fts5

package mirror

var _ = requires_sqlite_fts5_build_tag
