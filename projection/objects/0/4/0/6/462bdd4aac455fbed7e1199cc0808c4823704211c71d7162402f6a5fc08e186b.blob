#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

conformance_stage="${GOOO_CONFORMANCE_STAGE:-0}"
case "$conformance_stage" in
  0)
    echo "semantic conformance stage 0: Go verifier authoritative"
    ;;
  1|2|3)
    echo "semantic conformance stage $conformance_stage is not enabled; see .github/conformance-plan.md" >&2
    exit 1
    ;;
  *)
    echo "invalid GOOO_CONFORMANCE_STAGE: $conformance_stage" >&2
    exit 1
    ;;
esac

# Stage 0 keeps the deterministic Go verifier as the baseline while gooo is
# bootstrapped. The policy job adds the full changed-path scope check.
go run ./scripts/verify

if [[ ! -f cmd/gooo/main.go ]]; then
  echo "semantic conformance deferred: cmd/gooo is not present in this baseline"
  exit 0
fi

if ! grep -Fq 'case "check":' cmd/gooo/main.go; then
  echo "semantic conformance deferred: cmd/gooo check is not implemented in this baseline"
  exit 0
fi

go run ./cmd/gooo check examples/billing/main.gooo
go test -tags semantic_conformance ./internal/verify -run 'TestSemantic|TestGenerated' -count=1
./scripts/generated-freshness.sh
