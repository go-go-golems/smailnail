FROM golang:1.25.8-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/smailnail-imap-mcp ./cmd/smailnail-imap-mcp

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl tzdata \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /out/smailnail-imap-mcp /usr/local/bin/smailnail-imap-mcp
COPY scripts/docker-entrypoint.smailnail-imap-mcp.sh /usr/local/bin/docker-entrypoint.sh

RUN chmod +x /usr/local/bin/docker-entrypoint.sh

EXPOSE 3201

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD []
