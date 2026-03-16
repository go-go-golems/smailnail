.PHONY: gifs smoke-docker-imap smoke-imap-js-mcp smoke-js-module docker-build-imap-js-mcp

all: gifs

VERSION=v0.1.14

TAPES=$(wildcard doc/vhs/*tape)
gifs: $(TAPES)
	@if [ -z "$(TAPES)" ]; then \
		echo "No VHS tapes found under doc/vhs"; \
		exit 0; \
	fi
	for i in $(TAPES); do vhs < $$i; done

docker-lint:
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:latest golangci-lint run -v

lint:
	golangci-lint run -v

test:
	go test ./...

build:
	go generate ./...
	go build ./...

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
	go mod tidy

smailnail_BINARY=$(shell which smailnail)
install:
	go build -o ./dist/smailnail ./cmd/smailnail && \
		cp ./dist/smailnail $(smailnail_BINARY)

smoke-docker-imap:
	./scripts/docker-imap-smoke.sh

smoke-imap-js-mcp:
	./scripts/imap-js-mcp-smoke.sh

smoke-js-module:
	./scripts/js-module-smoke.sh

docker-build-imap-js-mcp:
	docker build -f Dockerfile.smailnail-imap-mcp -t smailnail-imap-mcp:dev .
