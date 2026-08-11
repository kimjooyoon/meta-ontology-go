# Testing research: invariant-centered quality gates

## Decision

Coverage is useful telemetry, but it is not the merge criterion for a semantic
compiler. A test is a quality gate only when it names an invariant, can fail on a
real regression, and records enough input identity to reproduce the result. The
important evidence is therefore semantic equivalence, locality, provenance, and
determinism—not the percentage of lines executed.

This review is based on the attached design brief, [`docs/spec.md`](../spec.md),
[`docs/architecture.md`](../architecture.md), [`AGENTS.md`](../../AGENTS.md), and
the open PRs present on 2026-08-12. The initial review baseline was `e394af6`,
which contained the repository verifier and a placeholder CLI; the syntax,
semantic, BX, analyzer, generator, LSP, and cache implementations were then on
separate PR branches. No submitted reviews or inline review threads were present
on the inspected open PRs.

Follow-up snapshot: `origin/integration` has since advanced to `74e90c4`, integrating
the syntax layer and staged self-hosting verifier governance. Semantic IR, BX,
analyzer, generator, LSP, cache, and query remain separate from that integration
snapshot. The current trust boundary is still Go-hosted Stage 0; `gooo`-host parity
is not evidence until the dual-evidence comparison in
[`.github/conformance-plan.md`](../../.github/conformance-plan.md) passes. The
remaining lanes should keep publishing independent contracts and fixtures against
minimal interfaces rather than waiting for the full package graph.

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

## Follow-up: living experiment board

This section is an ideation backlog, not a claim that every experiment is already
implemented. Each row should be revisited when its lane changes, when a failure is
minimized, and before promoting a verifier stage. The first executable artifact can
be a standalone fixture or adapter over a small interface; it must record a blocked
dependency instead of reporting an unimplemented lane as passed.

| Lane | Invariant/property experiment | Fuzz seed and minimization | Golden/contract experiment | Race and benchmark experiment | Promotion trigger |
| --- | --- | --- | --- | --- | --- |
| Syntax/AST | Permute comments, whitespace, declaration order, and Unicode while preserving the same normalized AST; assert recovery makes progress. | Empty input, invalid byte, truncated string, unterminated comment, invalid escape, and one valid declaration after an error. Keep the smallest byte seed that reproduces the diagnostic/span issue. | Canonical AST and diagnostic JSON with repository-relative spans; compare parse/lower output once semantic IR exists. | Parse the same immutable source concurrently; benchmark bytes and declarations with allocations reported, not wall-clock pass/fail. | P0 no panic/hang/span regression; P1 stable diagnostic golden. |
| Semantic IR/PROV | Permute nodes/facts, normalize twice, validate candidate/deterministic separation, and check relation-domain/cardinality rules. | Generate bounded IDs, namespaces, relation kinds, missing nodes, duplicate edges, and invalid spans; shrink to the smallest invalid graph. | Canonical IR, stable semantic hash, and PROV validity fixture; include a candidate that must not become truth. | Snapshot reads during promotion/normalization under `-race`; benchmark normalization/hash by node and fact count. | P0 laws and validation; P1 corpus of minimized invalid graphs. |
| BX/lowering/lifting | Check Get-Put, Put-Get, Diff/Apply, delta composition, transactional rejection, and locality closure over generated documents. | Malformed references, missing source spans, duplicate/removal conflicts, and mixed candidate/deterministic deltas; preserve the smallest failing delta. | Billing DTO contract: `.gooo -> IR -> generated DTO -> lifted delta -> IR fingerprint`; compare source-backed evidence. | Reconcile independent models concurrently and verify no shared mutation; benchmark lower/lift/reconcile by graph size. | P0 round-trip/locality; Stage 1 requires parity with the independent Go verifier. |
| Go analyzer | Permute registration order, import aliases, receiver names, local shadowing, and unknown helpers; only registered semantic symbols may lift. | Seed valid and malformed Go snippets, unresolved selectors, generic/type-alias forms, and nested functions; minimize source plus registry. | Canonical fact/candidate/implementation-detail JSON and a real `go/parser`/`go/types` contract when available. | Analyze immutable source concurrently with independent registries; benchmark AST/type resolution by file and symbol count. | P0 conservative lifting; P1 type-resolution and analyzer-to-BX contract. |
| Generator | Equal normalized IR/options produce equal bytes; changing one activity preserves unrelated regions, slots, and source-map ranges. | Empty IR, duplicate IDs, invalid Go names/types, malformed prior markers, nested/mismatched markers, and oversized names. | Versioned generated Go/source-map golden that passes `gofmt`, `go/parser`, and compilation; never store absolute paths. | Concurrent generation into separate temp dirs; benchmark generation by entities, ports, and marker count with operation counts. | P0 deterministic/local output and fail-closed markers; P1 golden review. |
| Query/search | Inverse and derived relations are order-independent, duplicate-free, bounded, and do not cross namespaces without an explicit edge. | Cycles, disconnected graphs, recursive rules, unknown predicates, depth limits, and duplicate facts; shrink graph and rule set together. | Canonical query-result JSON and bounded closure fixture tied to semantic IDs; compare with the rule engine's independent result. | Concurrent read-only queries over a snapshot; benchmark traversal/closure by graph size and depth, not timeout thresholds. | P0 termination and namespace safety; P1 derived-result golden. |
| LSP/JSON-RPC | Frame round-trip, lifecycle state, UTF-16 edits, document versions, diagnostics ordering, and invalid-request behavior are deterministic. | Empty/partial headers, oversized length, truncated payload, invalid JSON/params, unknown method, and split-at-every-byte seeds. | Canonical protocol transcript with IDs, paths, and timestamps normalized; include open/change/close/shutdown. | Run independent documents/servers under `-race`; benchmark framing and diagnostics by payload size with bounded resource limits. | P0 no desynchronization; P1 transcript and version-conflict contract. |
| Cache/freshness | Canonical key equality, first-writer immutability, digest validation, dependency invalidation, cancellation, and stale-evidence rejection hold. | Corrupt metadata/data, interrupted temp entries, path traversal/symlink candidates, unsupported/cyclic values, and dependency changes. | Cache manifest golden containing input/projection/evidence digests; restart fixture must distinguish miss, corruption, and hit. | Same-key and independent-key barriers plus cross-process restart; benchmark hit/miss/recompute and lock contention with operation counts. | P0 integrity/race; P1 restart/dependency contract; scheduled durability stress. |
| Provenance/evidence | Evidence is append-only, source-backed, digest-linked, and cannot certify a missing or stale projection. | Missing parent, wrong digest, duplicate sequence, future stage, malformed signature/manifest, and disconnected evidence paths. | Canonical evidence bundle for parse, generate, verify, and build; compare two independent runs byte-for-byte. | Concurrent append uses explicit serialization or immutable segments; benchmark evidence construction by edge count. | P0 freshness and chain validity; Stage 1 requires independent bundle comparison. |
| CLI/CI/self-hosting | Commands have stable exit/error contracts; scope policy is fail-closed; Go-hosted and gooo-hosted decisions agree before authority changes. | Arg permutations, missing files, malformed diffs, unknown branch aliases, unavailable revisions, and unsupported stage values. | End-to-end billing transcript plus Stage 0/1 evidence manifest; no deferred or unavailable result is rendered as pass. | Run isolated CLI processes in parallel temp dirs; benchmark verifier operations and fixture size, never CI success on latency alone. | Stage 0 now; Stage 1 only after dual evidence, reproducibility, comparator independence, and rollback rehearsal. |

## Flaky-test prevention and reproducibility

The following rules apply to every new experiment:

- Use channels, barriers, `sync.WaitGroup`, and explicit operation counters for
  synchronization. Do not use `time.Sleep` to prove ordering or mutual exclusion.
- Use `t.TempDir`, repository-relative paths, UTC, fixed locale, stable environment
  variables, and local random sources. Never depend on the developer's home,
  timezone, map iteration, wall-clock timestamp, or process scheduling.
- Every randomized property and fuzz failure records a seed, generator version,
  input digest, lane, invariant ID, and exact command. A minimized seed becomes a
  normal regression fixture before the failure is closed.
- Timeouts are watchdogs that classify a hang; they are not a correctness oracle.
  A timeout must preserve the input and stack/log evidence for triage.
- Golden output excludes absolute paths, timestamps, random IDs, hostnames, and
  toolchain-specific noise. Canonicalize before comparison and require an explicit
  update action for changes.
- Benchmarks report `testing.B` samples, allocations, and deterministic operation
  counts. PR CI should detect missing or pathological work; performance budgets and
  regression thresholds belong in scheduled runs with repeated samples.
- Race tests use bounded barriers and repeated runs. A retry may collect diagnosis,
  but must never hide a first failure or turn a flaky result into a pass.

## Failure triage protocol

The first failing signal is preserved as evidence, even if a rerun passes. Triage
follows this order:

1. Classify the lane, invariant ID, test kind, commit, base, and whether the failure
   is deterministic, seed-dependent, race-only, resource-bound, or environment-bound.
2. Re-run the exact command once with the recorded seed, `-count=1`, pinned toolchain,
   `GOOS/GOARCH`, `GOMAXPROCS`, and input digest. Do not broaden the scope by merging
   another lane to make the failure disappear.
3. Minimize the input: fuzz corpus to one seed, property data to the smallest model,
   golden output to the smallest semantic delta, or benchmark to the smallest
   operation count that still reproduces the anomaly.
4. Compare authoritative views in order: source/AST, IR/facts, projection/source map,
   lifted delta, cache/evidence digest, and CI decision. The first mismatch owns the
   defect; downstream symptoms are linked evidence.
5. Assign severity: P0 blocks a semantic law, safety boundary, integrity check, or
   deterministic output; P1 blocks a package promotion or creates a reproducibility
   gap; P2 is scheduled performance, corpus, or observability debt.
6. Add the minimized failure as a deterministic regression test or seed, then update
   the experiment board and promotion status. “Flaky” is a diagnosis requiring a
   cause and quarantine owner, not a disposition that permits merging.

The minimum triage record is:

```text
lane, invariant, test-kind, commit, base, input-digest, seed
go-version, GOOS/GOARCH, GOMAXPROCS, command, first-output
classification, severity, minimized-fixture, next-promotion-gate
```

## Self-hosting evidence ladder

The current [staged conformance plan](../../.github/conformance-plan.md) makes Go
the authoritative verifier at Stage 0. The testing strategy must remain comparable
as the implementation moves toward `gooo`:

- **Stage 0 — Go-hosted:** run the Go verifier, all deterministic gates, and explicit
  deferred markers for unavailable `gooo` checks. Deferred is not semantic success.
- **Stage 1 — dual evidence:** run Go and `gooo` verifiers on the same pinned
  checkout and compare normalized semantic results, scope decisions, generated
  freshness, and evidence manifests with an independent comparator. Any mismatch,
  unavailable input, or non-reproducible output blocks promotion.
- **Stage 2 — gooo authority with fallback:** require `gooo` while retaining the Go
  verifier as an independently runnable fallback; preserve both evidence bundles
  and rehearse rollback.
- **Stage 3 — fallback removal:** remove the fallback only after an independent
  parity audit, reproducible rebuild, forced-mismatch recovery, and reviewed
  governance change. The former verifier remains pinned as a recovery artifact.

For every stage, run the same fixture twice from the same commit and compare semantic
IDs, canonical bytes, digests, evidence decisions, and failure classifications. The
stage variable is fail-closed: an unimplemented future stage must fail or remain at
the prior stage, never silently pass.

## CI cadence and promotion contract

The living board maps experiments to three cadences:

| Cadence | Required evidence | Allowed variability | Promotion rule |
| --- | --- | --- | --- |
| Every PR | Unit/table tests, bounded properties, fuzz seed replay, reviewed goldens, package contract, repository vet/test/race, DAMP/DRY, and scope policy. | No wall-clock threshold, unpinned random seed, or retry-only pass. | A new P0 failure blocks; a P1 failure blocks the owning lane's promotion; P2 is recorded with an owner. |
| Scheduled | Longer fuzz budgets, repeated race stress, cross-process cache/restart, benchmark samples, dependency invalidation, and Stage 1 dual-evidence rehearsal. | Bounded duration and fixed environment matrix; raw samples retained. | A reproducible regression becomes a PR fixture; threshold changes require baseline and rationale. |
| Stage promotion | Full cross-package contract, two-run reproducibility, independent comparator, evidence freshness, rollback rehearsal, and explicit governance review. | None for semantic decisions; unavailable evidence is a block. | Move Go-hosted -> dual -> gooo authority only when the conformance plan's gate is met. |

When a dependency lane is absent from `integration`, its contract can use a small
DTO, interface, or fixture under the owning research scope. The result must say
“blocked: dependency X absent” and still prove the adapter's local law. Once the lane
is integrated, re-run the exact saved command and compare the evidence rather than
replacing it with a new unconnected smoke test.
