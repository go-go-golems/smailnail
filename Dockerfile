FROM node:22-bookworm AS ui-builder

WORKDIR /src/ui

COPY ui/package.json ui/pnpm-lock.yaml ./
RUN corepack enable && pnpm install --frozen-lockfile

COPY ui/ ./
RUN pnpm run build

FROM golang:1.26.1-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=ui-builder /src/ui/dist/public /src/pkg/smailnaild/web/embed/public

RUN GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -tags embed -o /out/smailnaild ./cmd/smailnaild

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl tzdata \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /out/smailnaild /usr/local/bin/smailnaild
COPY scripts/docker-entrypoint.smailnaild.sh /usr/local/bin/docker-entrypoint.sh

RUN chmod +x /usr/local/bin/docker-entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD []
