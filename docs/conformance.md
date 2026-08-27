# Runnable conformance examples

Run these commands from the repository root. They exercise the narrow `.gooo`
surface and the current CLI check/generation path; they do not imply that the
unstable `analyze` or `lsp` CLI surfaces exist.

## Billing fixture

The canonical fixture is [examples/billing/main.gooo](../examples/billing/main.gooo).
Its activity should derive two `used` facts and one `wasGeneratedBy` fact from
the declared activity contract.

```sh
go run ./cmd/gooo check examples/billing/main.gooo
```

The command should exit zero and print an `ok:` line. The repository-wide checks
are also part of the expected evidence:

```sh
go test ./...
go vet ./...
```

## Small independent fixture

The [conformance fixture](../examples/conformance/main.gooo) is intentionally
independent of the billing names. Run its check and generate a temporary
projection:

```sh
go run ./cmd/gooo check examples/conformance/main.gooo
out="$(mktemp -d)"
go run ./cmd/gooo generate examples/conformance/main.gooo --out "$out"
test -s "$out/semantic.gooo.go"
```

The generated file is temporary output and must not be committed. Its stable
generated markers are evidence for the projection boundary; handwritten logic
belongs in a slot or in the owning example package, not in generated text.

## What these examples prove

- parsing and lowering accept the current compact grammar;
- declared IDs and namespace-qualified names resolve deterministically;
- the CLI can report a semantic check and write a generated projection when that
  command is present;
- the examples provide a stable input for CI and future BX regression tests.

They do not prove a production LSP, automatic Go-to-DSL synchronization, cache
durability, or provenance publishing. Those require separate supported entry
points and evidence.

## External oracle humility

The [external oracle humility example](../examples/external-oracle-humility/README.md)
is verified by the `External oracle humility` Actions workflow. It keeps the
12-indicator denominator and three conformance branches fixed, re-lowers the
source in both producer and consumer packages, and performs Actions-only
current URL retrieval. The report exposes `source_policy=1/1`,
`producer_imports=0/0`, `historical_fixtures=3/3`,
`current_byte_observations=x/3`, `semantic_extraction=0/3`,
`semantic_agreement=0/3`, `semantic_causality=1/1`, and
`nonsemantic_preservation=1/1`. A historical capsule proves metadata
conformance only: every reference remains `HISTORICAL_FIXTURE / UNVERIFIED /
LOWER_RESOLUTION`, and the semantic agreement claim remains `OPEN` without raw
bytes plus a versioned extraction recipe and recipe digest. A successful
Actions retrieval can discharge only the current byte observation; retrieval
failure, digest mismatch, or recipe unavailability remains `OPEN` at the
appropriate `retrieve` or `extract` stage. Repository writes/promotions are
read from an effects snapshot.
