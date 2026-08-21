#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
generated_root="$(mktemp -d "${TMPDIR:-/tmp}/gooo-generated.XXXXXX")"
trap 'rm -rf -- "$generated_root"' EXIT

first="$generated_root/first"
second="$generated_root/second"
mkdir -p "$first" "$second"

cd "$repo_root"
go run ./cmd/gooo generate examples/billing/main.gooo --out "$first"
go run ./cmd/gooo generate examples/billing/main.gooo --out "$second"

generated_file="semantic.gooo.go"
cmp "$first/$generated_file" "$second/$generated_file"
test -z "$(gofmt -l "$first/$generated_file")"
grep -Fq '//gooo:generated:start' "$first/$generated_file"
grep -Fq '//gooo:generated:end' "$first/$generated_file"
