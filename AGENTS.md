# Agent and contributor contract

This repository is a semantic compiler, not only a text generator. Every change should
make the source view, semantic IR, Go projection, and verification evidence more
consistent.

## Authority boundaries

- Business intent is authoritative in `.gooo` DSL declarations.
- Stable semantic IDs are authoritative identity; display names and aliases may change.
- Handwritten Go owns irreducible implementation logic only.
- Generated Go owns structural boundaries and must use stable generated-region markers.
- Provenance facts and evidence are append-only records during a build.
- Ontology vocabulary, verifier semantics, and CI policy are protected kernel files.

## Agent roles

Builders may change application DSL and handwritten slots. Guardians may inspect and
verify but must not change the feature. The gate is deterministic: it checks semantic
scope, generated-region integrity, round-trip laws, and evidence freshness.

No single agent should implement, weaken verification, and approve the same change.

## Required checks

```sh
gofmt -w .
go vet ./...
go test ./...
go run ./cmd/gooo check examples/billing/main.gooo
```

Do not hand-edit generated regions. If a generated output needs a structural change,
change the DSL or generator and regenerate it.
