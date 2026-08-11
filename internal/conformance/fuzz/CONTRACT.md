# `.gooo` conformance contract

This directory is the executable boundary for malformed `.gooo` input. The
contract is deliberately narrower than the full compiler: it makes parser and
diagnostic behavior falsifiable now, while naming the input and output shapes
that later projections must preserve.

## Hypotheses

| ID | Hypothesis | Falsifier | Status |
| --- | --- | --- | --- |
| H1 | Lexing and parsing any input up to 8 KiB return without panic within 250 ms per operation. | A fuzz seed panics or reaches the deadline. | Implemented by `FuzzLexConformance` and `FuzzParseConformance`. |
| H2 | Repeating lex or parse with the same filename and bytes produces deeply equal tokens, AST, and diagnostics. | Any repeated result differs. | Implemented. |
| H3 | Every non-empty token, AST, and diagnostic span has bounded byte offsets and valid line/column positions. | A span escapes the source or has invalid positions. | Implemented. |
| H4 | A valid minimal fixture has no diagnostics; malformed fixtures retain a partial file and identify the expected diagnostic class. | The fixture expectation is not met. | Implemented by `contract.json` and `TestContractFixtures`. |
| H5 | The seed contract is cheap enough for ordinary CI and its cost is measurable without making wall-clock time a hard gate. | The benchmark regresses materially after a parser change. | Measurement only; threshold is deferred. |

The 8 KiB and 250 ms limits are test inputs, not claims about production
throughput. A future implementation may tighten them, but must update the
manifest and evidence together.

## Fixtures and evidence

`testdata/contract.json` is a machine-readable manifest. Its stable fields are:

```json
{
  "schema_version": 1,
  "max_source_bytes": 8192,
  "operation_timeout_ms": 250,
  "fixtures": [{
    "name": "valid-minimal",
    "file": "valid-minimal.gooo",
    "kind": "valid",
    "want_diagnostics": 0,
    "min_diagnostics": 0,
    "declarations": 2,
    "required_codes": []
  }]
}
```

The current minimum corpus contains one valid case and three counterexamples:

| Fixture | Expected evidence |
| --- | --- |
| `valid-minimal.gooo` | Two declarations and zero diagnostics. |
| `negative-unterminated-string.gooo` | `lex.unterminated-string`. |
| `negative-missing-arrow.gooo` | `parse.expected-arrow`. |
| `negative-illegal-character.gooo` | `lex.unexpected-character`. |

Reproduce the fixture evidence with:

```sh
go test ./internal/conformance/fuzz -run TestContractFixtures -count=1
```

On 2026-08-12 with Go 1.26.5 on Apple M4, this completed successfully in
about 0.40 seconds. This is an observation, not a portable CI threshold.

The reproducible measurement command is:

```sh
go test ./internal/conformance/fuzz -run '^$' -bench BenchmarkContractFixtures -benchmem -count=5
```

The observed isolated five-run range was 6.6–6.8 µs/op, with a median of
6.7 µs/op, 26,000 B/op, and 69 allocations/op for one four-fixture benchmark
batch. The benchmark is a regression signal; it is not allowed to convert an
unimplemented compiler stage into a passing result.

For a bounded fuzz measurement, use a fresh `GOCACHE` and run each target
serially rather than launching multiple fuzz processes on one host:

```sh
fuzz_cache_dir=$(mktemp -d /tmp/gooo-fuzz-cache.XXXXXX)
GOCACHE="$fuzz_cache_dir" go test ./internal/conformance/fuzz -run=^$ -fuzz=FuzzLexConformance -parallel=1 -fuzztime=2s
GOCACHE="$fuzz_cache_dir" go test ./internal/conformance/fuzz -run=^$ -fuzz=FuzzParseConformance -parallel=1 -fuzztime=2s
```

On the same host these completed without failures; the lex target processed
49,827 executions and the parse target 40,845 executions in the observed runs.
These counts are measurement evidence, not a fixed performance promise.
Running both fuzz processes concurrently can make the host-level fuzz deadline
expire under load, so concurrent launch is not a CI contract.

## Reusable pipeline contract

Every stage receives a source-backed envelope and emits deterministic records.
The parser stage is the only implemented projection in this directory.

| Stage | Input | Required output | Current status |
| --- | --- | --- | --- |
| AST | `filename`, UTF-8 bytes, source digest, language version | Ordered nodes with half-open spans plus ordered diagnostics | Implemented. |
| IR | AST and diagnostics | Stable semantic IDs, declaration kind, source span, and explicit diagnostic gate | Deferred; no IR success is claimed here. |
| BX | IR plus a source-backed Go delta | Deterministic `Added`, `Removed`, and `Candidate` facts with evidence spans | Deferred. |
| Go projection | IR and handwritten-slot map | Stable generated-region bytes, semantic IDs, and source map | Deferred. |
| LSP | Document URI, version, and bytes | Same diagnostics as AST parsing, with URI/version echoed | Deferred. |
| Cache | Stage name, schema version, source digest, policy digest | Hit/miss, key, artifact digest, and invalidation reason | Deferred. |
| Provenance | Activity, used inputs, generated outputs, source/evidence digests | Append-only derivation record with verification status | Deferred. |
| CI gate | Allowed semantic scope, actual delta, fresh artifacts, evidence | Pass, fail with counterexample, or explicit deferred status | Partial; changed-path ownership for this directory is awaiting CI registration. |

The canonical source envelope for later stages is:

```json
{
  "schema_version": 1,
  "filename": "billing.gooo",
  "source_sha256": "<lowercase hex digest>",
  "source_bytes": "<UTF-8 bytes>",
  "language": "gooo"
}
```

The parser output must never silently become an IR fact when diagnostics are
present. A lowering stage may return a partial AST for editor use, but must
mark the resulting IR as incomplete and preserve the diagnostic evidence.

## Hosting transition

The current evidence is explicitly **Go-hosted**: Go owns the parser API, test
harness, fixture loader, and benchmark. That stage is passing only for the
syntax/diagnostic contracts listed above. A future **gooo-hosted** stage may
describe the contract and its projections in `.gooo`, but it is not implemented
and is therefore deferred, not passing.

The transition gate is comparable evidence, not a rewrite of the expected
results. A gooo-hosted implementation must consume the same manifest, preserve
fixture names and source digests, emit equivalent diagnostic codes/spans, and
publish its own runtime/toolchain evidence. Any difference must be reported as
a semantic delta or a deferred capability; it must not be hidden by changing
the fixture expectations.

## Pass, fail, and deferred rules

- **Pass:** all fixture tests pass; repeated outputs are equal; no fuzz input
  panics or exceeds its per-operation deadline; measurements are recorded with
  command, toolchain, and machine context.
- **Fail:** any panic, timeout, nondeterministic output, out-of-bounds span,
  missing required diagnostic, or fixture manifest mismatch.
- **Deferred:** IR, BX, codegen, LSP, cache, provenance, and semantic CLI checks
  remain deferred until their implementations exist. Deferred is not success.
- **Ownership blocker:** the CI policy currently recognizes only its protected
  aliases. The CI owner must register `internal/conformance/fuzz/**`; this lane
  does not change protected CI policy files to make its own gate pass.

## Follow-up implementation contract

1. The parser keeps source bytes, filename, and spans as authoritative input;
   diagnostics are ordered and machine-coded.
2. Lowering consumes the AST plus diagnostic state and emits stable semantic
   IDs without inventing facts from recovery-only nodes.
3. BX and Go analysis attach every accepted fact to a source span and classify
   unsupported inferences as candidates.
4. Projections and caches include schema/source/policy digests, so stale output
   fails freshness checks instead of being treated as evidence.
5. CI compares actual semantic delta with allowed scope and reports the first
   counterexample path; it does not weaken the gate for a new projection.
