#!/usr/bin/env bash

set -euo pipefail

require_env() {
  local name="$1"
  if [[ -z "${!name:-}" ]]; then
    echo "missing required environment variable: ${name}" >&2
    exit 1
  fi
}

require_env KEYCLOAK_URL
require_env KEYCLOAK_ADMIN_REALM
require_env KEYCLOAK_CLIENT_ID
require_env KEYCLOAK_USERNAME
require_env KEYCLOAK_PASSWORD
require_env REALM_NAME
require_env DESIRED_SCOPES_JSON

access_token="$(
  curl -fsS \
    -X POST \
    "${KEYCLOAK_URL}/realms/${KEYCLOAK_ADMIN_REALM}/protocol/openid-connect/token" \
    -H 'Content-Type: application/x-www-form-urlencoded' \
    --data-urlencode "grant_type=password" \
    --data-urlencode "client_id=${KEYCLOAK_CLIENT_ID}" \
    --data-urlencode "username=${KEYCLOAK_USERNAME}" \
    --data-urlencode "password=${KEYCLOAK_PASSWORD}" \
  | jq -r '.access_token'
)"

if [[ -z "${access_token}" || "${access_token}" == "null" ]]; then
  echo "failed to obtain admin access token" >&2
  exit 1
fi

realm_json="$(
  curl -fsS \
    -H "Authorization: Bearer ${access_token}" \
    "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}"
)"

realm_id="$(jq -r '.id' <<<"${realm_json}")"
if [[ -z "${realm_id}" || "${realm_id}" == "null" ]]; then
  echo "failed to resolve realm id for ${REALM_NAME}" >&2
  exit 1
fi

components_json="$(
  curl -fsS \
    -H "Authorization: Bearer ${access_token}" \
    "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}/components?parent=${realm_id}&type=org.keycloak.services.clientregistration.policy.ClientRegistrationPolicy"
)"

component_json="$(
  jq -ce '
    map(select(.name == "Allowed Client Scopes" and .subType == "anonymous"))
    | if length == 1 then .[0] else empty end
  ' <<<"${components_json}"
)"

if [[ -z "${component_json}" ]]; then
  echo "failed to locate anonymous Allowed Client Scopes policy in realm ${REALM_NAME}" >&2
  exit 1
fi

component_id="$(jq -r '.id' <<<"${component_json}")"
updated_json="$(
  jq \
    --argjson desired "${DESIRED_SCOPES_JSON}" \
    '
      .config["allow-default-scopes"] = ["true"]
      | .config["allowed-client-scopes"] = $desired
    ' <<<"${component_json}"
)"

curl -fsS \
  -X PUT \
  -H "Authorization: Bearer ${access_token}" \
  -H 'Content-Type: application/json' \
  -d "${updated_json}" \
  "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}/components/${component_id}" \
  >/dev/null

verified_json="$(
  curl -fsS \
    -H "Authorization: Bearer ${access_token}" \
    "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}/components/${component_id}"
)"

verified_scopes="$(
  jq -c '(.config["allowed-client-scopes"] // []) | sort' <<<"${verified_json}"
)"

expected_scopes="$(
  jq -c 'sort' <<<"${DESIRED_SCOPES_JSON}"
)"

if [[ "${verified_scopes}" != "${expected_scopes}" ]]; then
  echo "verification failed: expected ${DESIRED_SCOPES_JSON}, got ${verified_scopes}" >&2
  exit 1
fi

echo "updated anonymous DCR allowed client scopes for realm ${REALM_NAME}: ${verified_scopes}"
