# meta-ontology-go

`meta-ontology-go` is an experimental semantic compiler implemented in Go. Its
surface files use the `.gooo` extension: **Go Of Ontology**.

The project has one deliberate authority boundary: `.gooo` declarations express
business intent. The compiler lowers that view to a normalized semantic IR and
projects structural boundaries to Go. Handwritten Go remains the source of truth
for irreducible implementation logic; generated Go, query output, documentation,
and CI results are derived views or evidence.

```text
.gooo intent ──lower──> semantic IR ──project──> generated Go
     │                         │                    │
     │                         └── facts/evidence   └── handwritten slots
     └── source spans, IDs, and explicit assertions
```

The repository is intentionally small. The supported language sketch currently
covers packages, namespaces, entities with URI-like IDs, and activities with
entity inputs and an entity result. The semantic kernel also defines a
PROV-inspired vocabulary, deterministic normalization, explicit candidate facts,
and marker-based generated regions. See [the architecture](docs/architecture.md)
and [the language sketch](docs/spec.md).

## Quick start

Run the repository checks from the project root:

```sh
test -z "$(gofmt -l .)"
go vet ./...
go test ./...
go test -race ./...
GOOO_CONFORMANCE_STAGE=0 ./scripts/semantic-conformance.sh
go run ./scripts/verify
```

The conformance walkthrough in [docs/conformance.md](docs/conformance.md) uses
the checked-in examples and shows the expected command shapes. It also calls out
which CLI surfaces are not yet supported. The semantic wrapper reports the
current `cmd/gooo check` stub as deferred; that output is not a semantic pass.
The current GitHub Actions workflow is documented in [CONTRIBUTING.md](CONTRIBUTING.md)
and the implementation/deferred ledger is in [docs/contracts.md](docs/contracts.md).

## Project status

This is an experimental language, not a stable application framework. In
particular, the repository does not currently promise a production LSP, a stable
`check`, `generate`, or `analyze` CLI, automatic promotion of ambiguous Go
observations, a code generator, a projection cache, or durable provenance
publishing. Internal research contracts may describe those directions, but a
feature is supported only when its command/API and conformance evidence are
present.

## Governance

[AGENTS.md](AGENTS.md) defines authority boundaries and agent roles.
[CONTRIBUTING.md](CONTRIBUTING.md) defines branch, PR, review, and CI workflow.
[docs/governance.md](docs/governance.md) records the SSOT boundary, BX laws, line
caps, and evidence policy. [docs/conformance.md](docs/conformance.md) is the
runnable example index.

[W3C PROV-O]: https://www.w3.org/TR/prov-o/
