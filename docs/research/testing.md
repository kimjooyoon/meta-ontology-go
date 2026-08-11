# Testing research: invariant-centered quality gates

## Decision

Coverage is useful telemetry, but it is not the merge criterion for a semantic
compiler. A test is a quality gate only when it names an invariant, can fail on a
real regression, and records enough input identity to reproduce the result. The
important evidence is therefore semantic equivalence, locality, provenance, and
determinism—not the percentage of lines executed.

This review is based on the attached design brief, [`docs/spec.md`](../spec.md),
[`docs/architecture.md`](../architecture.md), [`AGENTS.md`](../../AGENTS.md), and
the open PRs present on 2026-08-12. The current `origin/integration` baseline is
`e394af6`. It contains the repository verifier and a placeholder CLI; the syntax,
semantic, BX, analyzer, generator, LSP, and cache implementations are currently
separate PR branches. No submitted reviews or inline review threads were present
on the inspected open PRs.

The current CI already runs formatting, vet, unit tests, race tests, DAMP/DRY caps,
changed-path scope, and branch policy. Its semantic job explicitly defers when
`cmd/gooo check` is not implemented. A deferred job is useful transition evidence,
but must not be treated as semantic conformance once the owning implementation is
available.

## Quality model

Every gate should answer four questions:

1. Which authoritative view was the input: `.gooo`, semantic IR, handwritten Go,
   generated Go, cache metadata, or a verification policy?
2. Which semantic invariant was exercised?
3. Which semantic delta, source span, generated-region map, or cache digest was
   produced?
4. Can the result be replayed from the commit, input digest, test command, and
   toolchain version?

The test suite should use the standard library only. Table-driven tests cover named
examples; properties cover algebraic laws over generated values; fuzzing attacks
parser and transport boundaries; golden files protect stable projections; contract
tests exercise the complete projection chain; race tests protect shared mutable
state. These are complementary forms of evidence.

## Invariant matrix

| ID | Surface | Invariant and failure meaning | Minimum evidence | Priority |
| --- | --- | --- | --- | --- |
| SYN-DET | Lexer/parser | Same source and filename produce byte-for-byte equal tokens, diagnostics, AST, and spans. | Repeated parse plus declaration-order and whitespace fixtures. | P0 |
| SYN-SPAN | Lexer/parser | Token and AST spans are monotonic, bounded by the source, and use the documented byte/line/column convention. | Unicode, escaped strings, comments, EOF, and invalid UTF-8 fixtures. | P0 |
| SYN-REC | Parser diagnostics | Error recovery always makes progress, terminates, and never panics; valid declarations after an error remain inspectable. | Malformed byte fuzz target and recovery-progress property. | P0 |
| SEM-ID | Semantic IR | Identity is absolute and stable across display-name, alias, source-file, and namespace presentation changes; equal names in different namespaces do not collide. | Rename/alias/namespace table plus canonical identity property. | P0 |
| SEM-NORM | Semantic IR | Normalization is idempotent and insertion-order independent; duplicate facts and aliases have one defined result. | Permutation property and `Normalize(Normalize(x)) == Normalize(x)`. | P0 |
| SEM-PROV | Semantic IR/PROV | Only declared node kinds and allowed relations validate; deterministic facts, candidates, and evidence remain distinct. Missing source attribution is rejected. | Valid/invalid graph fixtures, candidate promotion transaction, provenance path assertions. | P0 |
| BX-GP | BX lens | Get-Put does not invent semantic facts or lose authoritative declarations. | DSL document -> model -> document, compared semantically. | P0 |
| BX-PG | BX lens | An accepted Go semantic delta is visible after Put-Get and retains source-backed evidence. | Delta reconciliation followed by a fresh lower and semantic fingerprint comparison. | P0 |
| BX-RT | BX lens | DSL -> IR -> Go -> lift -> IR is semantically equivalent; textual equality is not required. | Cross-package contract fixture using real syntax, analyzer, and generator. | P0 |
| BX-LOC | BX lens | An implementation-only edit has no semantic delta; a semantic edit touches only its relation closure and generated regions. | Unrelated-node fixture, source-map locality assertion, and preserved handwritten text. | P0 |
| AN-NS | Go analyzer | Only registered semantic symbols in the configured namespace become semantic facts; standard helpers remain implementation details. | Imports, aliases, methods, receivers, local shadowing, and unknown-call fixtures. | P0 |
| AN-CAND | Go analyzer | Ambiguous resolution remains a candidate with stable sorted options; it is never silently promoted. | Registration-order permutation and explicit-promotion contract. | P0 |
| GEN-DET | Generator | Equal normalized IR and options produce equal bytes, independent of map/order/process state. Output is gofmt-able and parseable. | Repeat generation, canonical input permutation, `format.Source`, and `go/parser`. | P0 |
| GEN-MARK | Generator | Generated start/end markers are balanced, unique, correctly identified, and never overwrite handwritten slots. Malformed markers fail closed. | Missing, duplicate, nested, mismatched, orphan, and slot-preservation fixtures. | P0 |
| GEN-LOC | Generator | A changed activity does not rewrite unrelated generated regions, source-map ranges, or handwritten material. | Region byte comparison plus semantic source-map range validation. | P0 |
| LSP-FRAME | LSP | Partial reads, short writes, malformed headers, truncated payloads, and arbitrary message boundaries produce bounded errors or valid frames without stream desynchronization. | Framing unit tests and byte fuzz target. | P0 |
| LSP-STATE | LSP | Initialize/open/change/close/shutdown/exit obey the JSON-RPC lifecycle; diagnostics and UTF-16 ranges describe the current document version. | Protocol transcript, invalid-request fixtures, versioned incremental edits, and cancellation cases. | P1 |
| CACHE-KEY | Cache | Canonical values with the same semantic content have the same key; type, nil, signed zero, time, version, options, and namespace distinctions are preserved. | Canonicalization table and permutation property. | P0 |
| CACHE-ATOMIC | Cache | A committed entry is complete and digest-valid; interrupted or tampered writes are misses/corruption, never valid data. First-writer immutability is explicit. | Temporary-file, corruption, recovery, and restart fixtures. | P0 |
| CACHE-CONC | Cache | Same-key computation has one winner and no race; independent keys may progress independently; cancellation never commits partial output. | `-race` stress test with barriers, errors, cancellation, and multiple keys. | P0 |
| POLICY-SCOPE | CI verifier | Actual changed paths and semantic deltas stay within allowed ownership; PRs use `agent/* -> integration`; exceptions are narrow and explicit. | Positive/negative policy tests at branch, path, and diff boundaries. | P0 |
| POLICY-CAPS | CI verifier | Tracked Go files are at most 300 lines and functions/methods/literals are at most 75 lines, including exact boundary cases. | 299/300/301 and 74/75/76-line fixtures, with and without final newline. | P1 |
| E2E-CHAIN | CLI/CI | The example can be checked, lowered, generated, analyzed, reconciled, and checked again without a semantic drift or stale projection. | Versioned conformance fixture and generated/cache freshness evidence. | P0 |

P0 means the implementation is not merge-ready without the invariant. P1 means it
can initially be package-local but must be promoted before the feature is relied on
by the CLI or editor. P2 work such as long-running fuzzing, cross-process durability
stress, and performance budgets should be scheduled after the P0 chain is working;
it must not replace P0 evidence.

## Open PR test-gap review

The following is a reviewer inventory, not approval of any PR. The PR descriptions
correctly report several repository-wide failures caused by unintegrated companion
packages; those failures must be re-run on the merged integration graph.

| PR | Existing proof | Important gap before integration approval |
| --- | --- | --- |
| [#1 generator](https://github.com/kimjooyoon/meta-ontology-go/pull/1) | Repeated generation, slot preservation, locality, malformed marker, and one formatted-output path. | No golden projection, no generated-source compile/type-check, no duplicate/nested marker matrix, and no fuzzing of IR or prior source. The test named `Formatted` does not itself call `gofmt`; formatting is asserted only in the locality test. |
| [#2 docs](https://github.com/kimjooyoon/meta-ontology-go/pull/2) | Documents SSOT/BX/governance and adds a conformance example. | Documentation claims need a runnable conformance consumer. A fixture alone does not prove parser-to-IR-to-Go-to-analyzer consistency. |
| [#3 semantic](https://github.com/kimjooyoon/meta-ontology-go/pull/3) | Identity, namespace isolation, normalization, candidate promotion, graph validation, and stable hash unit tests. | No property/fuzz corpus, no explicit relation-cardinality/PROV constraint matrix, no parser/lowering contract, and no concurrent snapshot or evidence-freshness contract. |
| [#4 CLI](https://github.com/kimjooyoon/meta-ontology-go/pull/4) | Smoke tests for check/generate/query/inspect/analyze/version/LSP and exit codes. | Output checks use substrings instead of deterministic goldens; invalid input, generated compilation, cache freshness, full BX round-trip, and cross-command error contracts are untested. It depends on companion package APIs, so package-local green is not repository green. |
| [#5 analyzer](https://github.com/kimjooyoon/meta-ontology-go/pull/5) | Registered-only lifting, candidates, unknown calls, registration order, namespace ordering, method mismatch, and local shadowing. | No type-check/import-alias/generic/type-alias matrix, malformed Go fuzzing, analyzer-to-BX reconciliation, or test that implementation details cannot leak into deterministic semantic facts. |
| [#7 bidir](https://github.com/kimjooyoon/meta-ontology-go/pull/7) | Get-Put, Put-Get, round-trip, fact layers, transactional reconciliation, diff/apply, normalization, and locality examples. | Most inputs are hand-built documents; there is no generated Go/analyzer end-to-end loop, randomized delta composition, removal/conflict permutation property, or fuzzing of malformed deltas. |
| [#8 cache](https://github.com/kimjooyoon/meta-ontology-go/pull/8) | Canonicalization, metadata integrity, corruption recovery, invalidation, limits, cancellation, same-key concurrency, and targeted race execution. | No process-crash/restart or cross-process writer test, dependency/evidence invalidation contract, path/symlink/permission threat matrix, independent-key concurrency property, or benchmark budget. |
| [#9 Go version](https://github.com/kimjooyoon/meta-ontology-go/pull/9) | One-line toolchain change; correctness is delegated to CI. | Must be validated only after the narrow `go.mod` exception is present; it should not be used to broaden dependency or policy scope. |
| [#10 LSP](https://github.com/kimjooyoon/meta-ontology-go/pull/10) | Partial framing, short writes, malformed/partial input, lifecycle, diagnostics, hover, completion, definition, UTF-16 edits, and parser seam. | No golden protocol transcript, malformed JSON/params/unknown-method matrix, document close/version conflict behavior, cancellation, multi-document or concurrent-client test, or bounded large-input policy. |
| [#11 syntax](https://github.com/kimjooyoon/meta-ontology-go/pull/11) | Lexer/parser tables, deterministic spans, diagnostics, comments/escapes, Unicode identifiers, recovery examples, and billing shape. | No fuzz target for arbitrary bytes, recovery-progress assertion, invalid UTF-8 policy, parse/lower equivalence, semantic-ID retention through AST, or golden diagnostic output. |
| [#12 CI policy](https://github.com/kimjooyoon/meta-ontology-go/pull/12) | Branch/path/toolchain exception unit tests and DAMP/DRY rejection fixture. | Add exact boundary tests, push-event versus pull-request environment coverage, and a test that the exception cannot admit `go.sum`, dependency, or unrelated-file changes. Keep the semantic defer visibly temporary. |

The PR set has good local unit coverage, but its current evidence is intentionally
fragmented. No PR currently supplies the complete `E2E-CHAIN`, and no PR adds a
golden or fuzz gate to the workflow. Integration should therefore be treated as a
new verification state, not as the sum of package-local green checks.

## DAMP/DRY and deterministic output gates

The repository verifier already checks the declared 300-line file and 75-line
function limits for tracked Go files. The negative fixture in `internal/verify`
proves rejection, but a durable gate also needs the boundary matrix:

- 299, 300, and 301 file lines, including a file without a final newline;
- 74, 75, and 76 lines for functions, methods, and function literals;
- generated Go and test fixtures, not only handwritten production files;
- deterministic, repository-relative violation ordering;
- the narrow `agent/go-version` `go.mod` exception without admitting `go.sum` or
  dependency changes.

For generated output, “deterministic” should mean all of the following:

- identical normalized IR and options produce identical bytes in repeated runs;
- map insertion order, declaration order, process working directory, and timezone do
  not affect output;
- the output passes `gofmt` and `go/parser`, and then compiles in the conformance
  fixture;
- generated-region markers and source-map ranges are stable; source paths in goldens
  are normalized to repository-relative paths;
- handwritten slots and unrelated regions remain byte-stable when the semantic delta
  is local.

Golden files should cover generated Go, canonical semantic IR, CLI JSON, and an LSP
protocol transcript. Goldens must not contain timestamps, absolute paths, random IDs,
or machine-specific formatting. Updating a golden should require an explicit local
`-update` action and a semantic-ID/change note; CI must compare, never update.

## Property and fuzz plan

Properties should use deterministic generators and a bounded case count in ordinary
PR tests. Good generators include declaration permutations, valid names and stable
IDs, fact-set permutations, source-backed deltas, marker streams, and canonical cache
values. The highest-value properties are:

- `Normalize(Normalize(x)) == Normalize(x)` and normalization is permutation-invariant;
- `Apply(Diff(a, b), a)` is semantically equivalent to `b`;
- Get-Put and Put-Get preserve the semantic fingerprint;
- a presentation-only edit has an empty semantic delta;
- reconciliation is transactional on any rejected delta;
- equal IR/options generate equal bytes and preserve unrelated regions;
- cache keys distinguish all documented type/nil/version distinctions and never
  accept a digest-invalid entry.

Fuzz targets should be small, panic-oriented, and seed-replayable in PR CI:

- syntax: arbitrary bytes must not panic, hang, or move spans backwards;
- analyzer: arbitrary Go source must not panic, and only registered symbols may lift;
- LSP: arbitrary headers/payload boundaries must either decode one frame or return a
  bounded error without desynchronizing the reader;
- generator: malformed IR/marker input must fail closed without overwriting a slot;
- cache: supported canonical values and corrupted metadata must never produce an
  invalid successful read.

The separate fuzz-conformance work should own the fuzz harness path. Once merged,
ordinary `go test` should replay its seed corpus, while a scheduled job may run a
bounded fuzz budget and upload new corpus entries as reviewable evidence. A fuzz job
must not be the only proof of a named law.

## Race, contract, and evidence plan

`go test -race ./...` is already a repository gate and should remain so. Targeted
stress tests should additionally cover cache locks, analyzer registries, LSP document
state, and any incremental compiler cache. Use barriers and operation counts rather
than sleeps where possible; a sleep-only test is evidence of timing, not mutual
exclusion. Run targeted race tests repeatedly during development, but keep the PR
gate bounded and reproducible.

The first cross-package contract fixture should be one small billing program:

```text
.gooo -> AST/parser -> semantic IR/PROV facts -> generated Go
       -> Go analyzer delta -> BX reconciliation -> semantic fingerprint
```

The fixture should assert stable IDs, expected `used`/`wasGeneratedBy` facts, source
spans for lifted relations, generated markers, handwritten-slot preservation, and an
empty delta after analyzing an unchanged projection. It should also fail on stale
generated output or stale cache metadata. This is the missing `E2E-CHAIN` evidence.

Every CI evidence record should include at least:

- commit SHA and toolchain version;
- input and projection digests;
- invariant ID and exact command;
- pass/fail result and relevant semantic delta/source-map/cache digest;
- whether the result was a unit, property, fuzz-seed, golden, contract, race, or
  deferred transition check.

Evidence is append-only for a build. A “deferred” result must remain distinguishable
from a passed conformance result, and a cached result must retain the input digest so
that freshness can be checked rather than inferred from a successful exit code.

## CI promotion order

The following order keeps the gate useful at every integration stage:

1. **Baseline, always required:** `gofmt -l .`, `go vet ./...`, `go test ./...`,
   `go test -race ./...`, DAMP/DRY caps, changed-path scope, and `agent/* ->
   integration` policy. No verifier or protected-kernel change may be smuggled in
   through a feature PR.
2. **Package invariants:** require the relevant P0 table/unit tests for every
   package that exists on the branch. Do not claim absent companion packages are
   tested merely because a scoped package test passed.
3. **Deterministic properties:** enable bounded normalization, diff/apply, lens,
   marker, canonicalization, and locality properties. Repeat deterministic tests
   with a fixed seed or `-count`, not an unbounded wall-clock loop.
4. **Seed corpus:** replay syntax/analyzer/LSP/generator/cache fuzz seeds in normal
   PR tests. Keep the seed corpus versioned and attach newly minimized failures to a
   change that explains the invariant.
5. **Golden projections:** compare canonical IR, generated Go/source maps, CLI JSON,
   and LSP transcripts. Golden updates require explicit review of semantic meaning.
6. **Cross-package conformance:** run the billing `E2E-CHAIN`, the example check,
   generation twice, analyzer lifting, BX reconciliation, and cache freshness. At
   this point the semantic job may no longer defer.
7. **Scheduled depth:** run bounded fuzzing, repeated race stress, cross-process
   cache/restart tests, and benchmark/operation-budget checks. Promote a discovered
   regression to a deterministic P0/P1 test before changing the schedule.

The immediate integration blocker is step 6, not lack of line coverage: until the
compiler packages are on one graph, repository-wide `go test`, `go vet`, race, and
the example CLI check remain the authoritative evidence to report as blocked or
passed. A green package-local PR should not be merged as if it proved the complete
semantic pipeline.

## Reviewer checklist

- [ ] The change names its authoritative view and the invariant IDs it exercises.
- [ ] Tests assert semantic fingerprints/deltas, locality, provenance, or stable
  bytes—not only non-empty output or line coverage.
- [ ] Deterministic output is checked across repeated runs and normalized inputs.
- [ ] Generated regions, handwritten slots, source maps, and cache metadata are
  tested for freshness and locality.
- [ ] Fuzz seeds, golden files, and contract fixtures have an explicit owner and CI
  promotion stage.
- [ ] `go test -race ./...`, `go vet ./...`, DAMP/DRY, and branch/path policy remain
  enabled; no deferred check is mistaken for a pass.
- [ ] The integration result records the exact blocked dependency or full
  end-to-end evidence, rather than relying on a scoped PR description.
