# Changelog

## 2026-04-03

- Initial workspace created
- Audited the annotation backend handoff against the current repository and created a corrected implementation guide centered on `smailnail sqlite serve`
- Added repository support for agent-run filtering, batch review updates, and aggregated run summaries/details in commit `4bda44f`
- Added the sqlite annotation UI server, query preset embedding, handler tests, and the `smailnail sqlite serve` command in commit `9a7345a`
- Validated the final implementation with `go test -tags sqlite_fts5 ./...` and a tmux-backed runtime smoke using `go run -tags sqlite_fts5 ./cmd/smailnail sqlite serve --sqlite-path ./smailnail-mirror.sqlite --listen-port 18080`
