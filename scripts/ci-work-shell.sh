#!/usr/bin/env bash
set -euo pipefail
test "$#" -eq 1
test -n "${GOOO_CI_DOMAIN:-}"
test -n "${GOOO_CI_STARTED_AT:-}"
test -n "${GOOO_CI_SOURCE_HEAD:-}"
test -n "${GOOO_CI_ORIGIN_ROOT:-}"
test -n "${GOOO_CI_BUDGET_MS:-}"
test -n "${GITHUB_JOB:-}"
test -n "${GITHUB_ACTION:-}"

# GitHub supplies the expanded step body. Preserve that actual command extent.
# The journal and receipts remain outside the active logical source checkout.
body="$(cat -- "$1")"
purpose="${GITHUB_EVENT_NAME}:${GITHUB_BASE_REF:-${GITHUB_REF_NAME}}"
unit="urn:gooo:ci-command:${purpose}:${GITHUB_JOB}:${GITHUB_ACTION}"
receipt="$GOOO_CI_ORIGIN_ROOT/${GITHUB_ACTION}.json"
arguments=(work --root "$PWD" --head "$GOOO_CI_SOURCE_HEAD" --unit "$unit"
  --started-at "$GOOO_CI_STARTED_AT" --budget-ms "$GOOO_CI_BUDGET_MS"
  --journal "$GOOO_CI_ORIGIN_ROOT/command-origins.jsonl")
if [[ -n "${GOOO_CI_ACTIVATION_ORIGIN:-}" ]]; then
  arguments+=(--workspace-origin "$GOOO_CI_ACTIVATION_ORIGIN")
fi
exec "$GOOO_CI_DOMAIN" "${arguments[@]}" -- \
  /bin/bash --noprofile --norc -e -o pipefail -c "$body" > "$receipt"
