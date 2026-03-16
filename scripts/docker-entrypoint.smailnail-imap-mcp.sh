#!/bin/sh
set -eu

if [ "$#" -gt 0 ]; then
  exec /usr/local/bin/smailnail-imap-mcp "$@"
fi

TRANSPORT="${SMAILNAIL_MCP_TRANSPORT:-streamable_http}"
PORT="${SMAILNAIL_MCP_PORT:-3201}"
AUTH_MODE="${SMAILNAIL_MCP_AUTH_MODE:-none}"

set -- /usr/local/bin/smailnail-imap-mcp mcp start \
  --transport "$TRANSPORT" \
  --port "$PORT"

if [ -n "${SMAILNAIL_MCP_INTERNAL_SERVERS:-}" ]; then
  OLD_IFS="$IFS"
  IFS=','
  for internal_server in $SMAILNAIL_MCP_INTERNAL_SERVERS; do
    if [ -n "$internal_server" ]; then
      set -- "$@" --internal-servers "$internal_server"
    fi
  done
  IFS="$OLD_IFS"
fi

if [ "$AUTH_MODE" != "none" ]; then
  set -- "$@" --auth-mode "$AUTH_MODE"
fi

if [ -n "${SMAILNAIL_MCP_AUTH_RESOURCE_URL:-}" ]; then
  set -- "$@" --auth-resource-url "$SMAILNAIL_MCP_AUTH_RESOURCE_URL"
fi

if [ -n "${SMAILNAIL_MCP_OIDC_ISSUER_URL:-}" ]; then
  set -- "$@" --oidc-issuer-url "$SMAILNAIL_MCP_OIDC_ISSUER_URL"
fi

if [ -n "${SMAILNAIL_MCP_OIDC_DISCOVERY_URL:-}" ]; then
  set -- "$@" --oidc-discovery-url "$SMAILNAIL_MCP_OIDC_DISCOVERY_URL"
fi

if [ -n "${SMAILNAIL_MCP_OIDC_AUDIENCE:-}" ]; then
  set -- "$@" --oidc-audience "$SMAILNAIL_MCP_OIDC_AUDIENCE"
fi

if [ -n "${SMAILNAIL_MCP_OIDC_REQUIRED_SCOPES:-}" ]; then
  OLD_IFS="$IFS"
  IFS=','
  for scope in $SMAILNAIL_MCP_OIDC_REQUIRED_SCOPES; do
    if [ -n "$scope" ]; then
      set -- "$@" --oidc-required-scope "$scope"
    fi
  done
  IFS="$OLD_IFS"
fi

if [ -n "${SMAILNAIL_MCP_APP_DB_DRIVER:-}" ]; then
  set -- "$@" --app-db-driver "$SMAILNAIL_MCP_APP_DB_DRIVER"
fi

if [ -n "${SMAILNAIL_MCP_APP_DB_DSN:-}" ]; then
  set -- "$@" --app-db-dsn "$SMAILNAIL_MCP_APP_DB_DSN"
fi

if [ -n "${SMAILNAIL_MCP_APP_ENCRYPTION_KEY_ID:-}" ]; then
  set -- "$@" --app-encryption-key-id "$SMAILNAIL_MCP_APP_ENCRYPTION_KEY_ID"
fi

if [ -n "${SMAILNAIL_MCP_APP_ENCRYPTION_KEY_BASE64:-}" ]; then
  set -- "$@" --app-encryption-key-base64 "$SMAILNAIL_MCP_APP_ENCRYPTION_KEY_BASE64"
fi

if [ -n "${SMAILNAIL_MCP_EXTRA_ARGS:-}" ]; then
  # shellcheck disable=SC2086
  set -- "$@" $SMAILNAIL_MCP_EXTRA_ARGS
fi

exec "$@"
