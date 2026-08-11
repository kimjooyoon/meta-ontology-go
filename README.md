# meta-ontology-go

`meta-ontology-go` is an experimental semantic programming language implemented in Go.
Its source files use the `.gooo` extension: **Go Of Ontology**.

The project treats a small business DSL as an authoritative semantic view and lowers it
to a generalized semantic IR. Go code, documentation, search projections, tests, and
verification evidence are derived views. Handwritten Go remains authoritative for
irreducible implementation logic.

The design borrows the provenance vocabulary of [W3C PROV-O] and the graph-first agent
workflow of [zerolang], while adding an explicit bidirectional transformation boundary:

```text
.gooo DSL <-> semantic IR <-> Go projection + handwritten slots
                  |
                  +--> provenance, search, docs, CI evidence
```

The first milestone is intentionally small and inspectable:

- a deterministic lexer/parser and lossless-ish AST with source spans;
- semantic identities and namespace-safe PROV-inspired facts;
- DSL-to-IR lowering and Go symbol lifting;
- stable generated regions and semantic source maps;
- a standard-library-only LSP server;
- content-addressed incremental caches;
- round-trip, scope, provenance, and generated-code checks in GitHub Actions.

This repository is early-stage and the language surface is expected to evolve. Stable
semantic IDs and compatibility rules matter more than preserving provisional syntax.

## Quick start

```sh
go test ./...
go run ./cmd/gooo check examples/billing/main.gooo
go run ./cmd/gooo generate examples/billing/main.gooo --out .gooo-gen
go run ./cmd/gooo lsp
```

See [docs/spec.md](docs/spec.md), [docs/architecture.md](docs/architecture.md), and
[CONTRIBUTING.md](CONTRIBUTING.md) for the current contract.

[W3C PROV-O]: https://www.w3.org/TR/prov-o/
[zerolang]: https://github.com/vercel-labs/zerolang
