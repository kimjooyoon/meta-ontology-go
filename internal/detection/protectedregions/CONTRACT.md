# Protected-region contract

Status: prototype evidence for the generated-Go boundary. This contract is
intended to be consumed by codegen, AST/BX adapters, LSP diagnostics, cache
keys, provenance records, and CI policy. It does not claim that an unimplemented
`.gooo` stage exists.

## Falsifiable hypothesis

For a valid pair of generated Go projections with stable marker IDs:

1. changing bytes only inside a generated region body is accepted as local;
2. changing a slot or handwritten body is rejected as `protected-change`;
3. changing unmarked text or generated boundaries is rejected as
   `unowned-change`;
4. repeating validation on identical bytes produces identical regions and
   issues.

The fixture in `testdata/locality` and `TestContractGeneratedBodyOnlyChangesAreLocal`
are the minimum counterexample set. A failed case falsifies the hypothesis and
must block an adapter from using the validator as an overwrite guard.

## Input contract

- `Validate(source []byte)` receives one complete Go projection as bytes. It
  does not parse Go AST semantics and does not mutate the input.
- Generated boundaries use stable IDs, for example
  `//gooo:generated:start id="billing://activity/pay-order"`; slot boundaries
  use `//gooo:slot:start/end`; handwritten boundaries use
  `//gooo:handwritten:start/end` or `//gooo:protected:start/end`.
- IDs are semantic identity, not display names. A rename with the same ID must
  not create a new region identity.
- `ValidateLocality(before, after)` receives the previous and candidate bytes;
  both must pass structural validation before locality is evaluated.

## Output contract

- `Report.Regions` contains paired regions in source order with byte offsets and
  a body range. `Report.Issues` is deterministic and classifies malformed
  structure (`nested`, `missing`, `duplicate`, `mismatched`, and ID/scope
  errors).
- `LocalityReport.Valid()` is true only when both reports are structurally
  valid, generated region skeletons are unchanged, and slot/handwritten bodies
  are byte-identical.
- `LocalityReport.Err()` is suitable for a gate and must not be used to infer
  business semantics. A caller must retain the structured report for evidence.

## Consumer obligations

| Consumer | Required use | Deferred boundary |
| --- | --- | --- |
| Go codegen | Validate candidate output; run `CheckLocality` before replacing an existing file. | Generator adapter is deferred until its package is integrated. |
| AST / BX | Carry stable region ID and source/body offsets as a projection delta. | AST semantic lifting is not implemented here. |
| LSP | Map `Issue.Line` and `Region` offsets to diagnostics/code lenses. | JSON-RPC publication is deferred. |
| Cache | Include source bytes and validator version in the cache key; never cache a pass for different bytes. | Content-addressed cache integration is deferred. |
| Provenance | Record validator result as evidence, including input digests and issue list. | PROV entity/activity persistence is deferred. |
| CI | Treat structural or locality errors as a deterministic failure; do not promote deferred `.gooo` stages to pass. | Current CI ownership allowlist must register this package before merge. |

## Evidence criteria

- **Pass:** positive generated-body-only fixture passes; each negative fixture
  reports its expected issue class; repeated runs are equal; benchmark command
  completes under the local measurement protocol.
- **Fail:** any protected/unmarked mutation passes, any generated-only mutation
  fails, or repeated identical inputs produce different structured output.
- **Deferred:** generator/AST/BX/LSP/cache/provenance integrations are not
  counted as implemented until their adapters and independent tests land.

Measurement command:

```sh
go test ./internal/detection/protectedregions -run 'TestContract' -bench 'BenchmarkProtectedRegion' -benchtime=1s -count=3 -benchmem
```

The benchmark measures a 64-region projection pair and is evidence of local
runtime/allocation behavior, not a portable performance threshold.

Sample measurement on Go 1.26.5, darwin/arm64, Apple M4, three one-second
runs: 99.8–141.9 µs/op, 156.6–222.7 MB/s, 302,079–302,083 B/op, and 1,705
allocations/op. The allocation count is a follow-up optimization target, not a
failure condition for this structural contract.
