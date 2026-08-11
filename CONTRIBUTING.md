# Contributing

The language is deliberately experimental. Proposals should explain the semantic
identity impact, the authoritative view being changed, and the round-trip behavior.

Keep packages dependency-free unless a dependency is justified in the design record.
Prefer deterministic output, explicit source spans, stable IDs, and small compositional
forms. Ambiguous Go facts must remain candidates until a `.gooo` assertion or policy
promotes them.

Before opening a change, run:

```sh
gofmt -w .
go vet ./...
go test ./...
```

Changes to ontology vocabulary, semantic identity rules, verifier semantics, or policy
require a design note and focused regression tests.
