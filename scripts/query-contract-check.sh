#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fixture="$repo_root/examples/conformance/query-contract.fixture"

fail() {
  echo "query contract: FAIL: $*" >&2
  exit 1
}

section_count() {
  local section="$1"
  awk -v wanted="[$section]" '
    /^\[/ { current=$0; next }
    current == wanted && $0 !~ /^[[:space:]]*#/ && NF { count++ }
    END { print count + 0 }
  ' "$fixture"
}

measurement() {
  local key="$1"
  awk -v wanted="$key" '
    /^\[measurements\]$/ { active=1; next }
    /^\[/ { active=0 }
    active && $1 == wanted { print $2; exit }
  ' "$fixture"
}

canonical_digest() {
  if command -v sha256sum >/dev/null 2>&1; then
    awk '
      /^\[oracle\.canonical\]$/ { active=1; next }
      /^\[/ { active=0 }
      active && NF { print }
    ' "$fixture" | sha256sum | awk '{print $1}'
    return
  fi
  awk '
    /^\[oracle\.canonical\]$/ { active=1; next }
    /^\[/ { active=0 }
    active && NF { print }
  ' "$fixture" | shasum -a 256 | awk '{print $1}'
}

[[ -f "$fixture" ]] || fail "missing fixture $fixture"

line_count="$(wc -l < "$fixture" | tr -d '[:space:]')"
deterministic="$(section_count deterministic)"
candidates="$(section_count candidate)"
oracle_rows="$(section_count oracle.canonical)"
expected_digest="$(measurement oracle_sha256)"
actual_digest="$(canonical_digest)"

[[ "$line_count" == "$(measurement fixture_lines)" ]] || fail "fixture line count changed"
[[ "$deterministic" == "$(measurement deterministic_facts)" ]] || fail "deterministic fact count changed"
[[ "$candidates" == "$(measurement candidate_facts)" ]] || fail "candidate fact count changed"
[[ "$oracle_rows" == "$(measurement oracle_rows)" ]] || fail "oracle row count changed"
[[ "$expected_digest" != "PLACEHOLDER" ]] || fail "oracle digest is not recorded"
[[ "$actual_digest" == "$expected_digest" ]] || fail "canonical oracle digest changed"

grep -Fxq 'exact_input deterministic=1 candidate=0' "$fixture" || fail "exact input oracle missing"
grep -Fxq 'exact_candidate deterministic=0 candidate=1' "$fixture" || fail "candidate oracle missing"
grep -Fxq 'negative_unbounded error=invalid-max-depth' "$fixture" || fail "bounded rejection oracle missing"

printf 'query-contract/v1: fixture oracle PASS\n'
printf 'measurements: lines=%s deterministic=%s candidate=%s oracle_rows=%s sha256=%s\n' \
  "$line_count" "$deterministic" "$candidates" "$oracle_rows" "$actual_digest"
printf 'query_engine_conformance: DEFERRED (implementation adapter not present in this baseline)\n'
printf 'gooo_hosted_stage: NOT_RUN (not a promotion claim)\n'
