#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

test -z "$(gofmt -l .)"
go vet ./...
go test ./...
go test -race ./...
go test ./internal/verify -count=1
./scripts/semantic-conformance.sh
go run ./scripts/verify
