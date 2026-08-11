# Runnable conformance examples

Run these commands from the repository root. They exercise the narrow `.gooo`
surface and the current Go-authoritative verifier. The CLI command names are
present for bootstrap planning, but the current `cmd/gooo` implementation is a
stub, so direct `check` and `generate` commands are deferred rather than passing
examples.

## Billing fixture

The canonical fixture is [examples/billing/main.gooo](../examples/billing/main.gooo).
Its activity should derive two `used` facts and one `wasGeneratedBy` fact from
the declared activity contract.

The repository-wide checks and staged semantic wrapper are the expected evidence:

```sh
test -z "$(gofmt -l .)"
go test ./...
go vet ./...
go test -race ./...
GOOO_CONFORMANCE_STAGE=0 ./scripts/semantic-conformance.sh
go run ./scripts/verify
```

## Small independent fixture

The [conformance fixture](../examples/conformance/main.gooo) is intentionally
independent of the billing names. It is a source fixture for the semantic kernel
and future CLI conformance; it is not currently executable through `cmd/gooo`.

```sh
go test ./internal/syntax ./internal/semantic ./internal/verify
```

Generated Go, LSP, cache, and Go-analysis execution remain deferred. Their
reusable contracts are recorded in [docs/contracts.md](contracts.md).

## What these examples prove

- parsing and lowering accept the current compact grammar;
- declared IDs and namespace-qualified names resolve deterministically;
- the Go verifier can run the current policy and evidence checks;
- the examples provide stable inputs for future BX and bootstrap regression tests.

They do not prove a production LSP, automatic Go-to-DSL synchronization, cache
durability, or provenance publishing. Those require separate supported entry
points and evidence.
