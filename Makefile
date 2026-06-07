.PHONY: gifs smoke-docker-imap smoke-imap-js-mcp smoke-js-module docker-build-imap-js-mcp \
	dev-backend dev-frontend frontend-build frontend-check build-embed

all: gifs

VERSION=v0.1.14
SQLITE_TAGS ?= sqlite_fts5

TAPES=$(wildcard doc/vhs/*tape)
gifs: $(TAPES)
	@if [ -z "$(TAPES)" ]; then \
		echo "No VHS tapes found under doc/vhs"; \
		exit 0; \
	fi
	for i in $(TAPES); do vhs < $$i; done

docker-lint:
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:latest golangci-lint run -v --build-tags $(SQLITE_TAGS)

lint:
	golangci-lint run -v --build-tags $(SQLITE_TAGS)

test:
	go test -tags "$(SQLITE_TAGS)" ./...

build:
	go generate ./...
	go build -tags "$(SQLITE_TAGS)" ./...

goreleaser:
	goreleaser release --skip=sign --snapshot --clean

tag-major:
	git tag $(shell svu major)

tag-minor:
	git tag $(shell svu minor)

tag-patch:
	git tag $(shell svu patch)

release:
	git push --tags
	GOPROXY=proxy.golang.org go list -m github.com/go-go-golems/smailnail@$(shell svu current)

bump-glazed:
	go get github.com/go-go-golems/glazed@latest
	go get github.com/go-go-golems/clay@latest
	go get github.com/go-go-golems/geppetto@latest
	go get github.com/go-go-golems/go-go-goja@latest
	go get github.com/go-go-golems/go-go-mcp@latest
	go mod tidy

smailnail_BINARY=$(shell which smailnail)
install:
	go build -tags "$(SQLITE_TAGS)" -o ./dist/smailnail ./cmd/smailnail && \
		cp ./dist/smailnail $(smailnail_BINARY)

smoke-docker-imap:
	./scripts/docker-imap-smoke.sh

smoke-imap-js-mcp:
	./scripts/imap-js-mcp-smoke.sh

smoke-js-module:
	./scripts/js-module-smoke.sh

docker-build-imap-js-mcp:
	docker build -t smailnail-imap-mcp:dev .

# --- Frontend dev loop ---

UI_DIR ?= ui
DEV_API_PORT ?= 3001
DEV_UI_PORT ?= 3000

dev-backend:
	go run ./cmd/smailnaild serve --listen-port $(DEV_API_PORT)

dev-frontend:
	pnpm -C $(UI_DIR) dev --host --port $(DEV_UI_PORT)

frontend-build:
	go generate ./pkg/smailnaild/web/

frontend-check:
	pnpm -C $(UI_DIR) run check

build-embed: frontend-build
	go build -tags "embed $(SQLITE_TAGS)" ./...

.PHONY: logcopter-generate
logcopter-generate:
	GOWORK=off GOFLAGS="-tags=$(SQLITE_TAGS)" go tool logcopter-gen -include-main -var zlog -area-prefix go-go-golems.smailnail -strip-prefix github.com/go-go-golems/smailnail ./cmd/... ./pkg/...

.PHONY: logcopter-check
logcopter-check:
	GOWORK=off GOFLAGS="-tags=$(SQLITE_TAGS)" go tool logcopter-gen -include-main -var zlog -area-prefix go-go-golems.smailnail -strip-prefix github.com/go-go-golems/smailnail -check ./cmd/... ./pkg/...

GLAZED_LINT_BIN ?= /tmp/glazed-lint
GLAZED_LINT_PKG ?= github.com/go-go-golems/glazed/cmd/tools/glazed-lint
GLAZED_VERSION ?= v1.3.6

.PHONY: glazed-lint-build glazed-lint

glazed-lint-build:
	@echo "Building glazed-lint from Glazed module..."
	@if [ -n "$(GLAZED_VERSION)" ]; then \
		echo "Installing $(GLAZED_LINT_PKG)@$(GLAZED_VERSION)"; \
		GOBIN=$(dir $(GLAZED_LINT_BIN)) GOWORK=off go install $(GLAZED_LINT_PKG)@$(GLAZED_VERSION); \
	else \
		echo "Installing $(GLAZED_LINT_PKG) from workspace/module"; \
		GOBIN=$(dir $(GLAZED_LINT_BIN)) go install $(GLAZED_LINT_PKG); \
	fi

# Legacy helper commands and MCP identity middleware still use raw environment/flag
# plumbing; keep the rollout analyzer enabled everywhere else.
GLAZED_LINT_ALLOW_PATHS ?= cmd/build-web/,cmd/imap-tests/,pkg/mcp/imapjs/

glazed-lint: glazed-lint-build
	GOWORK=off GOFLAGS="-tags=$(SQLITE_TAGS)" go vet -vettool=$(GLAZED_LINT_BIN) -glazedclilint.allow-paths=$(GLAZED_LINT_ALLOW_PATHS) ./cmd/... ./pkg/...

.PHONY: bump-go-go-golems
bump-go-go-golems:
	@deps="$$(awk '/^require[[:space:]]+github\.com\/go-go-golems\// { print $$2 } /^[[:space:]]*github\.com\/go-go-golems\// { print $$1 }' go.mod | sort -u)"; \
	if [ -z "$$deps" ]; then \
		echo "No github.com/go-go-golems dependencies in go.mod"; \
	else \
		echo "Bumping go-go-golems dependencies:"; \
		echo "$$deps"; \
		for dep in $$deps; do GOWORK=off go get "$${dep}@latest"; done; \
	fi
	GOWORK=off go mod tidy

