#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
output="${1:-${RUNNER_TEMP:-/tmp}/language-package-execution}"
mkdir -p "$output"
cd "$root"

case "$(go env GOVERSION)" in
  go1.27*) ;;
  *) echo "expected Go 1.27, observed $(go env GOVERSION)" >&2; exit 1 ;;
esac

unformatted="$(gofmt -l internal/packageruntime/packageexecution internal/meta/languagepackageexecution cmd/language-package-execution-witness cmd/gooo/run_package_source.go cmd/gooo/run_package_source_test.go)"
test -z "$unformatted" || { printf 'unformatted Go files:\n%s\n' "$unformatted" >&2; exit 1; }

go vet ./internal/packageruntime/packageexecution ./internal/meta/languagepackageexecution ./cmd/language-package-execution-witness ./cmd/gooo
go test ./internal/packageruntime/packageexecution ./internal/meta/languagepackageexecution ./cmd/gooo
go fix ./internal/packageruntime/packageexecution ./internal/meta/languagepackageexecution ./cmd/language-package-execution-witness
git diff --exit-code -- internal/packageruntime/packageexecution internal/meta/languagepackageexecution cmd/language-package-execution-witness

go build -o "$output/gooo" ./cmd/gooo
"$output/gooo" run --entry PayOrder examples/billing-package > "$output/cli-receipt.json"
go run ./cmd/language-package-execution-witness --head "${EXACT_SHA:-${GITHUB_SHA:-0000000000000000000000000000000000000000}}" --root "$root" --out "$output/report.json"

if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
  jq -r '"### Gooo package execution\n\n| Metric | Observed | Target |\n| --- | ---: | ---: |\n" + ([.indicators[] | "| `\(.id)` | \(.value) | \(.target) |"] | join("\n")) + "\n\nDecision: `\(.decision)` / `\(.resolution)`"' "$output/report.json" >> "$GITHUB_STEP_SUMMARY"
fi
