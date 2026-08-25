#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 5 ]]; then
  echo "usage: main.sh BINARY ROOT HEAD OUTPUT REPEATS" >&2
  exit 2
fi
binary="$(realpath "$1")"
root="$(realpath "$2")"
head="$3"
output="$4"
repeats="$5"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
samples="$work/samples.jsonl"

profile() {
  local operation="$1"
  shift
  local arguments='[]'
  for argument in "$@"; do
    arguments="$(jq -c --arg value "$argument" '. + [$value]' <<<"$arguments")"
  done
  for sequence in $(seq 1 "$repeats"); do
    local sample="$work/${operation}-${sequence}"
    mkdir -p "$sample"
    local started finished exit_code
    started="$(date +%s%N)"
    set +e
    (cd "$root" && /usr/bin/time -f '%M' -o "$sample/rss" env LANG=C LC_ALL=C TZ=UTC NO_COLOR=1 "$binary" "$@" >"$sample/stdout" 2>"$sample/stderr")
    exit_code=$?
    set -e
    finished="$(date +%s%N)"
    jq -cn --arg operation "$operation" --argjson arguments "$arguments" \
      --argjson sequence "$sequence" --argjson exit_code "$exit_code" \
      --argjson wall_ms "$(((finished-started)/1000000))" \
      --argjson max_rss_kib "$(cat "$sample/rss")" \
      --argjson stdout_bytes "$(wc -c <"$sample/stdout")" \
      --argjson stderr_bytes "$(wc -c <"$sample/stderr")" \
      --arg stdout_digest "sha256:$(sha256sum "$sample/stdout" | cut -d' ' -f1)" \
      --arg stderr_digest "sha256:$(sha256sum "$sample/stderr" | cut -d' ' -f1)" \
      '{operation:$operation,arguments:$arguments,sequence:$sequence,exit_code:$exit_code,wall_ms:$wall_ms,max_rss_kib:$max_rss_kib,stdout_bytes:$stdout_bytes,stderr_bytes:$stderr_bytes,stdout_digest:$stdout_digest,stderr_digest:$stderr_digest}' >>"$samples"
  done
}

source='examples/billing/main.gooo'
profile VERSION_TEXT version
profile VERSION_JSON version --json
profile CHECK_TEXT check "$source"
profile CHECK_JSON check --json "$source"
profile ROUNDTRIP_JSON roundtrip --json "$source"
profile SEMANTIC_CHECK check --semantic "$source"
profile RUN_SOURCE run --json --entry PayOrder "$source"

mkdir -p "$(dirname "$output")"
jq -s --arg subject_sha "$head" --arg os "${ImageOS:-ubuntu24}" \
  --arg architecture "${RUNNER_ARCH:-$(uname -m)}" --arg image "ubuntu-24.04" \
  --arg image_version "${ImageVersion:-unknown}" --arg go_version "$(go env GOVERSION)" \
  --arg executable_digest "sha256:$(sha256sum "$binary" | cut -d' ' -f1)" \
  --argjson executable_size "$(stat -c%s "$binary")" --arg source_path "$source" \
  --arg source_digest "sha256:$(sha256sum "$root/$source" | cut -d' ' -f1)" \
  '{schema:"gooo/user-journey-resource-observation/v1",subject_sha:$subject_sha,runner:{os:$os,architecture:$architecture,image:$image,image_version:$image_version,go_version:$go_version},executable:{digest:$executable_digest,size_bytes:$executable_size},source_path:$source_path,source_digest:$source_digest,samples:.}' \
  "$samples" >"$output"
