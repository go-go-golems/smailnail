#!/bin/sh
set -eu

if [ "$#" -gt 0 ]; then
  exec /usr/local/bin/smailnaild "$@"
fi

LISTEN_HOST="${SMAILNAILD_LISTEN_HOST:-0.0.0.0}"
LISTEN_PORT="${SMAILNAILD_LISTEN_PORT:-8080}"
DB_TYPE="${SMAILNAILD_DB_TYPE:-sqlite}"
DATABASE="${SMAILNAILD_DATABASE:-/data/smailnaild.sqlite}"
AUTH_MODE="${SMAILNAILD_AUTH_MODE:-oidc}"
SESSION_COOKIE_NAME="${SMAILNAILD_AUTH_SESSION_COOKIE_NAME:-smailnail_session}"
MCP_ENABLED="${SMAILNAILD_MCP_ENABLED:-1}"
MCP_TRANSPORT="${SMAILNAILD_MCP_TRANSPORT:-streamable_http}"
MCP_AUTH_MODE="${SMAILNAILD_MCP_AUTH_MODE:-none}"

set -- /usr/local/bin/smailnaild serve \
  --listen-host "$LISTEN_HOST" \
  --listen-port "$LISTEN_PORT" \
  --db-type "$DB_TYPE" \
  --auth-mode "$AUTH_MODE" \
  --auth-session-cookie-name "$SESSION_COOKIE_NAME"

if [ -n "${SMAILNAILD_LOG_LEVEL:-}" ]; then
  set -- "$@" --log-level "$SMAILNAILD_LOG_LEVEL"
fi

if [ -n "${SMAILNAILD_DSN:-}" ]; then
  set -- "$@" --dsn "$SMAILNAILD_DSN"
else
  set -- "$@" --database "$DATABASE"
fi

if [ -n "${SMAILNAILD_ENCRYPTION_KEY_ID:-}" ]; then
  set -- "$@" --encryption-key-id "$SMAILNAILD_ENCRYPTION_KEY_ID"
fi

if [ -n "${SMAILNAILD_ENCRYPTION_KEY_BASE64:-}" ]; then
  set -- "$@" --encryption-key-base64 "$SMAILNAILD_ENCRYPTION_KEY_BASE64"
fi

if [ -n "${SMAILNAILD_OIDC_ISSUER_URL:-}" ]; then
  set -- "$@" --oidc-issuer-url "$SMAILNAILD_OIDC_ISSUER_URL"
fi

if [ -n "${SMAILNAILD_OIDC_CLIENT_ID:-}" ]; then
  set -- "$@" --oidc-client-id "$SMAILNAILD_OIDC_CLIENT_ID"
fi

if [ -n "${SMAILNAILD_OIDC_CLIENT_SECRET:-}" ]; then
  set -- "$@" --oidc-client-secret "$SMAILNAILD_OIDC_CLIENT_SECRET"
fi

if [ -n "${SMAILNAILD_OIDC_REDIRECT_URL:-}" ]; then
  set -- "$@" --oidc-redirect-url "$SMAILNAILD_OIDC_REDIRECT_URL"
fi

if [ -n "${SMAILNAILD_OIDC_SCOPES:-}" ]; then
  OLD_IFS="$IFS"
  IFS=','
  for scope in $SMAILNAILD_OIDC_SCOPES; do
    if [ -n "$scope" ]; then
      set -- "$@" --oidc-scopes "$scope"
    fi
  done
  IFS="$OLD_IFS"
fi

case "$MCP_ENABLED" in
  1|true|TRUE|yes|on)
    set -- "$@" --mcp-enabled
    ;;
  *)
    set -- "$@" --mcp-enabled=false
    ;;
esac

set -- "$@" --mcp-transport "$MCP_TRANSPORT" --mcp-auth-mode "$MCP_AUTH_MODE"

if [ -n "${SMAILNAILD_MCP_AUTH_RESOURCE_URL:-}" ]; then
  set -- "$@" --mcp-auth-resource-url "$SMAILNAILD_MCP_AUTH_RESOURCE_URL"
fi

if [ -n "${SMAILNAILD_MCP_OIDC_ISSUER_URL:-}" ]; then
  set -- "$@" --mcp-oidc-issuer-url "$SMAILNAILD_MCP_OIDC_ISSUER_URL"
fi

if [ -n "${SMAILNAILD_MCP_OIDC_DISCOVERY_URL:-}" ]; then
  set -- "$@" --mcp-oidc-discovery-url "$SMAILNAILD_MCP_OIDC_DISCOVERY_URL"
fi

if [ -n "${SMAILNAILD_MCP_OIDC_AUDIENCE:-}" ]; then
  set -- "$@" --mcp-oidc-audience "$SMAILNAILD_MCP_OIDC_AUDIENCE"
fi

if [ -n "${SMAILNAILD_MCP_OIDC_REQUIRED_SCOPES:-}" ]; then
  OLD_IFS="$IFS"
  IFS=','
  for scope in $SMAILNAILD_MCP_OIDC_REQUIRED_SCOPES; do
    if [ -n "$scope" ]; then
      set -- "$@" --mcp-oidc-required-scopes "$scope"
    fi
  done
  IFS="$OLD_IFS"
fi

if [ -n "${SMAILNAILD_EXTRA_ARGS:-}" ]; then
  # shellcheck disable=SC2086
  set -- "$@" $SMAILNAILD_EXTRA_ARGS
fi

exec "$@"
