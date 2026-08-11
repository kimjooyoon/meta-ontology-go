# Cache and freshness research

This note defines a research contract for the `.gooo` compiler cache. It is a
design proposal, not an implementation change. The cache must accelerate
reconstructable projections without becoming a second source of semantic truth.

The repository already names `internal/cache` as a content-addressed,
versioned, atomic incremental cache. The implementation work on the separate
`agent/cache` branch is treated as an input to this note; this branch changes
only this document.

## Executive decision

Use three deliberately separate layers:

1. **Semantic identity and canonical hash**: a deterministic digest of a
   normalized semantic closure, with stable IDs and explicit schema/domain
   versions.
2. **Local projection cache**: immutable, content-addressed bytes plus a
   verifiable metadata envelope. Cache loss must never lose DSL, IR authority,
   provenance records, or durable evidence.
3. **Freshness and evidence**: append-only provenance and verification records
   that prove which inputs, policy, toolchain, and output were checked. A cache
   hit is not evidence and must not be accepted as proof by itself.

The cache key is therefore more than a source-file hash:

```text
key = H(
  domain || key-schema || artifact-kind || projection ||
  semantic-hash || dependency-root || policy-hash ||
  toolchain || target-platform || options
)
```

Every field that can change output must either be in this key or be represented
by a dependency whose digest is included transitively. Observational fields such
as wall-clock time, runner ID, and cache access count must not affect the key.

## Scope and authority boundaries

The cache stores reconstructable projections such as normalized IR, generated
Go, search materializations, diagnostics, or test scaffolding. It does not own:

- `.gooo` business declarations and their stable semantic IDs;
- authoritative handwritten logic;
- ontology vocabulary, verifier semantics, or CI policy;
- append-only provenance facts;
- durable evidence, attestations, or merge decisions.

This follows the repository's multi-SSOT model: the DSL is authoritative for
business intent, the semantic IR is authoritative for generalized meaning, and
PROV facts are authoritative for derivation history. The cache is only a
reconstructable optimization over those authorities. This boundary is also
consistent with Bazel's distinction between an action cache and a content
addressable store: action results can be reused, but their declared inputs and
outputs still define the computation ([Bazel remote caching](https://bazel.build/remote/caching)).

## Canonical semantic hashing

### Hash the normalized meaning, not source presentation

The semantic hash should be computed after parsing and lowering, before any
projection-specific formatting. Its input is a canonical semantic closure:

```text
SemanticClosure {
  root IDs:        sorted stable semantic IDs
  entities:        sorted (id, semantic type, declared constraints)
  activities:      sorted (id, semantic type, declared effects)
  relations:       sorted (subject, predicate, object, qualifiers)
  assertions:      sorted explicit domain assertions
  dependencies:    sorted referenced semantic IDs and their hashes
  schema:          canonicalizer and IR schema versions
}
```

The following must not change the semantic hash when the stable IDs and
meaning remain unchanged:

- declaration order when the language defines declarations as a set;
- display names, aliases, preferred names, or source formatting;
- comments and whitespace;
- source spans, unless the requested artifact is explicitly a source map;
- map/dictionary insertion order.

The following must change it:

- a stable semantic ID or namespace;
- a relation endpoint, predicate, qualifier, or assertion;
- an input/output port or constraint;
- an ontology or verifier rule that changes lowering;
- a generator, compiler, policy, target, or option that changes bytes.

This makes a rename with an unchanged semantic ID a cache-preserving operation,
while preventing namespace collisions such as `billing::Payment` and
`settlement::Payment` from sharing a result accidentally.

### Canonical byte format

Do not use ordinary Go map JSON as the semantic hash format. JSON object order,
number handling, omitted defaults, and language-specific values can silently
create multiple encodings for one value or one encoding for different values.
For interchange-oriented JSON, [RFC 8785 JCS](https://www.rfc-editor.org/rfc/rfc8785.html)
provides deterministic property sorting and primitive serialization. The IR
hash should instead use a small versioned typed encoding with these rules:

1. Prefix every value and relation with a domain and type tag.
2. Encode strings as validated UTF-8 with an explicit byte length.
3. Encode sequences in semantic order; sort set-like collections by their
   canonical encoded bytes.
4. Encode maps as sorted key/value pairs and reject duplicate canonical keys.
5. Distinguish null, absent, empty, signed zero, integer widths, and floating
   point values where the IR permits them; otherwise normalize them before
   hashing.
6. Normalize URI/semantic-ID syntax according to the language contract, but do
   not apply unadvertised Unicode normalization to user strings.
7. Reject unsupported values, cycles, NaN policy violations, and unexported
   implementation fields at the canonicalization boundary.
8. Include the canonicalizer version and IR schema version in the domain.

The recommended digest envelope is:

```text
gooo-semantic-hash\0
algorithm=sha256\0
ir-schema=v1\0
canonicalizer=v1\0
<length-prefixed canonical closure bytes>
```

SHA-256 is adequate for the first implementation and is already used by the
candidate cache. Keep the algorithm in the serialized envelope so a future
algorithm migration creates a new namespace instead of silently reusing old
objects. A digest is an address and integrity check, not a formal proof that a
collision is impossible.

### Separate semantic hash from cache key

`SemanticHash` answers “what meaning is this?” A cache key answers “which exact
projection computation is this?” They should not be conflated.

```text
SemanticHash = H(canonical semantic closure)
ProjectionKey = H(
  cache-domain ||
  artifact-kind || projection ||
  SemanticHash || dependency-root || policy-hash ||
  compiler/toolchain || target || options
)
```

The current candidate `KeySpec` shape—key version, namespace, tool version,
canonical inputs, and canonical options—is a useful storage API. The semantic
layer should add explicit fields for `SemanticHash`, `DependencyRoot`,
`PolicyHash`, artifact kind, projection, and target when those values affect
the result. `CreatedAt` and access timestamps belong only in metadata.

## Dependency graph and invalidation

### Model dependencies as semantic edges

File paths are useful discovery hints but are not sufficient invalidation
identity. Build a dependency graph with nodes for:

- source files and declaration spans;
- stable semantic IDs and normalized relations;
- projection actions such as `lower`, `generate-go`, `index`, and `verify`;
- policy, ontology, generator, and toolchain versions;
- produced artifacts and evidence records.

An edge says that the downstream node's bytes or validity depend on the
upstream node. The artifact graph should be acyclic; cycles in the semantic
relation graph must be collapsed into explicitly identified strongly connected
components before dependency hashing.

Each projection records both direct input digests and a deterministic
`DependencyRoot`, for example:

```text
DependencyRoot = H(
  sorted (dependency-id, dependency-kind, dependency-digest, edge-qualifiers)
)
```

This gives two invalidation mechanisms:

- **Key invalidation**: a changed dependency produces a new key, so stale bytes
  cannot be a hit even if old objects remain on disk.
- **Reachability invalidation**: reverse edges identify affected projections
  for eager recomputation, selective deletion, or diagnostics.

Key invalidation is the correctness mechanism. Deletion is a space-management
policy and must never be required for correctness.

### Incremental algorithm

The proposed build sequence is:

1. Parse changed source and retain source spans for diagnostics only.
2. Lower the affected declarations into normalized facts.
3. Compare canonical fact digests with the previous dependency index.
4. Mark changed, added, and deleted stable IDs. A deletion is a change, not an
   absent event that can be ignored.
5. Traverse reverse dependency edges to find affected projections.
6. Recompute each projection's `DependencyRoot` and exact key.
7. Read and validate exact-key entries; compute only misses or corrupt entries.
8. Publish new immutable entries, then append provenance and verification
   evidence for the build that used them.
9. Garbage-collect unreachable objects only after a separate retention and
   locking check.

The index itself is reconstructable. If it is lost, a cold scan or full build
can recreate it from authoritative DSL/IR and provenance records. A dependency
edge that cannot be resolved must conservatively produce a miss or a failed
verification, never a fresh hit.

### Policy and ontology changes

Protected ontology, canonicalizer, verifier semantics, and CI policy files are
dependencies of every affected projection. Their digest must be in the policy
or schema component of the key. This prevents an old generated file or an old
verification result from being accepted after a rule change.

Toolchain and target changes follow the same rule. At minimum include Go
version, compiler version, relevant build tags, GOOS, GOARCH, and generator
version for Go projections. Avoid using a broad “invalidate everything” switch
when a precise policy digest can identify the affected closure.

## Provenance and evidence freshness

### Cache metadata is not evidence

Every cache object should carry a small integrity envelope containing at least:

```text
format/schema version
projection and artifact kind
exact projection key
semantic hash
dependency root
input and options digests
policy/ontology hash
compiler and toolchain identity
target platform
content digest and byte size
reconstructable=true
```

The envelope can detect corruption and explain a miss. It does not establish
that a verifier ran, that a claim was approved, or that the output is safe to
ship. `CreatedAt` is useful for operations but is not a freshness proof.

PROV validation is a useful model here: the W3C specification defines
normalization, validity, equivalence, uniqueness, and event-order constraints
for provenance records ([PROV constraints](https://www.w3.org/TR/prov-constraints/)).
The compiler should apply the same discipline to its build evidence rather than
treating a cached byte blob as a provenance record.

### Freshness predicate

Define a verification result as fresh only when all of the following match the
current expected state:

```text
Fresh(E, T) iff
  E.subject_digest       == T.expected_output_digest
  E.semantic_hash        == T.semantic_hash
  E.dependency_root      == T.dependency_root
  E.policy_hash          == T.policy_hash
  E.canonicalizer/schema == T.canonicalizer/schema
  E.toolchain/target     == T.toolchain/target
  E.verifier_identity    is trusted for T
  every required input and predecessor evidence is fresh
  any declared time window has not expired
```

The last two conditions are important. An evidence record can have a matching
artifact digest but still be invalid because it was produced under an old
policy, by an untrusted verifier, or from a stale predecessor. Freshness must
be evaluated over the evidence dependency graph, not only over the leaf file.

Recommended provenance topology:

```text
DSL / handwritten logic
  -> parse evidence
  -> normalized IR
  -> generated projection
  -> verification activity
  -> verification result / evidence
  -> build artifact
  -> optional signed attestation
```

Each activity records `used` input digests and `wasGeneratedBy` output
relationships. Evidence records are append-only; a failed or superseded result
is retained with its status and reason. A new result supersedes an old result
through a new provenance edge rather than mutating history. For release-facing
artifacts, signed attestations can bind a subject digest to its build context;
GitHub documents artifact attestations as provenance claims that should be
verified by consumers ([artifact attestations](https://docs.github.com/en/actions/how-tos/secure-your-work/use-artifact-attestations/use-artifact-attestations)).

Freshness failure reasons should be machine-readable: `missing-input`,
`dependency-changed`, `semantic-hash-mismatch`, `policy-changed`,
`toolchain-changed`, `schema-changed`, `verifier-untrusted`, `evidence-expired`,
`content-corrupt`, and `unknown-dependency`. The CI gate can then explain a
miss or reject a stale claim without guessing.

## Storage, publication, and concurrent access

### Correctness without a global lock

Cache objects should be immutable after publication. Write into a private
temporary directory, write and sync the data and metadata, then publish the
complete directory with one same-filesystem rename. Readers accept only a
complete object whose metadata and content digests agree. Incomplete temporary
directories are ignored or removed during maintenance.

Go documents that `os.Rename` is not atomic on every platform, so the storage
implementation must either constrain the publication contract to supported
filesystems or add a platform-specific fallback and test it
([`os.Rename`](https://go.dev/pkg/os/#Rename)). Never claim crash atomicity for
cross-device moves, network filesystems, or unsupported Windows behavior.

With immutable content-addressed objects, two processes that compute the same
key concurrently may both finish safely: they must produce byte-identical
content or the second publication must be rejected as a key/content mismatch.
Correctness therefore does not depend on a process-wide mutex.

### Stampede control and lock hierarchy

Locking is an efficiency feature and must be layered on top of atomic
publication:

- process-local, per-key mutex for same-process `GetOrCompute` calls;
- optional cross-process per-key lease for expensive computations;
- short global maintenance lock for invalidation/index rewrite/GC;
- no long-held global lock around ordinary reads.

The cross-process lease must use an atomic acquisition operation, include owner
metadata and a bounded wait, and leave a recoverable record on crash. Stale
lease reclamation is risky because PIDs can be reused and clocks can jump; do
not delete a live owner's lock based only on age. If the platform cannot offer a
safe lease, allow duplicate computation and rely on immutable publication.

`GetOrCompute` should double-check after acquiring a stampede lock. Context
cancellation must stop waiting and computing where possible. A lock must never
be part of a semantic hash, evidence claim, or durable provenance fact. The
lock controls scheduling, not meaning.

The minimum concurrency tests are:

- many readers while one writer publishes the same key;
- many writers for one key with identical bytes;
- competing writers that intentionally produce different bytes;
- process restart with temporary directories and stale leases;
- invalidation racing with reads and GC;
- `go test -race` for in-process state.

## GitHub Actions cache integration

GitHub Actions cache is a remote CI acceleration layer, not the authoritative
semantic cache or evidence store. GitHub scopes caches by key, cache version,
and branch; existing cache contents cannot be changed in place. An exact key is
therefore the safe semantic-cache lookup. The documented `restore-keys`
behavior intentionally permits partial matches and may return the most recent
matching prefix ([dependency caching reference](https://docs.github.com/en/actions/reference/workflows-and-actions/dependency-caching)).

The current workflow's `actions/setup-go@v5` `cache: true` is treated as the
ordinary Go dependency/build cache. It should remain separate from a future
`.gooo` projection cache, whose manifest and semantic key require the stronger
validation rules below.

Recommended workflow rules:

1. Use the built-in Go module/build cache for ordinary Go dependencies and a
   separate path for `.gooo` reconstructable projections.
2. Include OS, architecture, Go/toolchain version, workflow version, cache
   schema, projection name, and semantic/dependency root in the primary key.
3. Treat an exact primary-key hit as a candidate only after local metadata and
   content verification. Treat a `restore-keys` partial hit as a warm seed or
   miss, never as a fresh semantic projection or evidence result.
4. Do not store secrets, credentials, mutable authority, or attestations in the
   Actions cache. GitHub warns that restored cache contents are untrusted input
   and that pull requests can expose cache contents to workflows with access
   ([caching concepts](https://docs.github.com/en/actions/concepts/workflows-and-actions/dependency-caching)).
5. Do not let an untrusted pull request publish trusted release evidence. A
   cache restored in a low-trust workflow must be revalidated in the trusted
   integration/release job.
6. Keep the cache key below GitHub's length limit and version the path/archive
   format whenever compression or layout changes.
7. Use artifacts, not caches, to pass build outputs or test reports between jobs
   when retention and explicit provenance matter. Use attestations only for
   release-facing subjects that will be verified.

The repository already uses workflow-level concurrency to cancel superseded CI
runs. That reduces redundant work but is not a filesystem lock and does not
provide cache correctness. GitHub's concurrency groups are a scheduler control
with cancellation/queue semantics ([workflow concurrency](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/control-workflow-concurrency)).

An initial CI key shape is:

```text
gooo-${runner.os}-${arch}-${go-version}-${cache-schema}-
  ${projection}-${semantic-hash}-${dependency-root}
```

Do not add a broad prefix fallback for generated semantic outputs until the
workflow verifies the restored manifest and recomputes the exact key. A partial
cache can be useful for dependency downloads, but it is not proof of semantic
freshness.

## Cache invariants

These are proposed conformance invariants. The first five are correctness
gates; the remainder are safety, operational, or performance gates.

| ID | Invariant | Required check |
| --- | --- | --- |
| C1 | Determinism: equal canonical semantic closures and computation inputs produce the same semantic hash and key. | Repeat with reordered maps, declarations, and processes. |
| C2 | Identity stability: display-name or formatting-only changes preserve the semantic hash when stable IDs and meaning are unchanged. | Rename/format round-trip fixture. |
| C3 | Namespace isolation: distinct namespaces or stable IDs cannot share a key accidentally. | Same display names in two namespaces. |
| C4 | Completeness: every output-affecting input is represented directly or transitively in the key. | Mutate each field in a dependency fixture and require a changed key. |
| C5 | No false hit: a changed dependency, policy, schema, toolchain, target, or option cannot return the old projection as an exact hit. | Mutation matrix over all key components. |
| C6 | Integrity: a hit requires valid metadata, matching key, matching content digest, and expected byte size. | Tamper data and metadata independently. |
| C7 | Atomic visibility: readers see either the previous complete object or the new complete object, never a partial entry. | Kill/restart and concurrent read/write tests. |
| C8 | Reconstructability: deleting all cache objects does not delete authoritative DSL, IR, provenance, or evidence. | Cold rebuild from the repository and durable records. |
| C9 | Fresh evidence: verification results match current subject, dependency root, policy, schema, toolchain, and trusted verifier. | Stale predecessor and policy-change fixtures. |
| C10 | Append-only history: failures and superseded evidence remain queryable and are not overwritten by cache writes. | Compare provenance record count/content before and after rebuild. |
| C11 | Concurrency safety: same-key concurrent computes cannot publish conflicting bytes under one key; duplicate work is bounded or observable. | N-process same-key and divergent-output stress tests. |
| C12 | Invalidation precision: an unrelated semantic subgraph does not invalidate an unrelated projection. | Disjoint namespace/subgraph fixture. |
| C13 | Conservative uncertainty: unknown or missing dependencies cause a miss or verification failure, never a fresh hit. | Delete dependency-index entries and simulate unknown policy. |
| C14 | Safe GC: only unreachable, unlocked, unreferenced objects are collected after the retention policy permits it. | GC with active readers, leases, and evidence references. |

## Benchmark plan

Benchmarks should measure both speed and the cost of making a hit trustworthy.
The acceptance baseline is zero false hits and zero corrupted publications;
latency targets should be set after the first measurements rather than hiding
correctness regressions behind a speed target.

### Benchmark matrix

| Area | Fixtures | Metrics |
| --- | --- | --- |
| Canonicalization | 10/100/1,000/10,000 entities and relations; nested qualifiers; maps; long IDs; empty vs absent values. | ns/op, bytes/op, allocations/op, digest throughput. |
| Key construction | Same semantic closure with changed display names, IDs, options, policy, and toolchain. | key determinism, changed-key rate, CPU and allocation cost. |
| Local cache | Cold miss, warm exact hit, corrupt entry, metadata-only read, 1 KiB/64 KiB/1 MiB/100 MiB artifacts. | p50/p95 latency, fsync cost, read/write throughput, disk amplification. |
| Invalidation | DAGs with depth 2/5/10, fan-out 2/8/32, 0.1%/1%/10% changed roots, deletions, and SCCs. | affected-node precision, recomputation count, index scan time. |
| Concurrency | 1/2/4/8/32 workers and separate processes for one key and disjoint keys. | wall time, duplicate computations, lock wait, contention, error rate. |
| Evidence | Fresh leaf, stale predecessor, changed policy, changed verifier, expired window. | freshness decision latency and explanation completeness. |
| Actions cache | Cold run, exact hit, partial restore, changed semantic root, changed toolchain, branch/default-branch lookup. | download/upload time, hit classification, recomputation time, cache size. |

### Protocol

Record commit, Go version, compiler/generator versions, OS/architecture,
filesystem, cache schema, fixture dimensions, and whether the run is cold or
warm. Run local benchmarks with `go test -bench . -benchmem -count=10` and
compare revisions with `benchstat`. Run filesystem cases on a local SSD and a
representative CI runner; do not mix network-cache latency into the local
algorithm benchmark. Run concurrency cases under `go test -race` and with
multiple OS processes, because an in-process race detector cannot prove
cross-process lock safety.

The benchmark harness should emit machine-readable records that include:

```text
benchmark name, fixture dimensions, key/schema versions,
hit/miss classification, recompute count, lock wait,
bytes read/written, error and freshness reason
```

The first performance review should answer:

- Is canonicalization linear in the number of semantic facts for ordinary
  graphs, apart from explicitly documented sorting costs?
- Does incremental invalidation recompute only the reverse dependency closure?
- Does same-key locking reduce stampedes enough to justify its wait time?
- Does metadata verification dominate warm-hit latency for small projections?
- Does a GitHub Actions cache hit still save time after exact-key validation?

### Measured local baseline

The following is an observed baseline, not an acceptance decision. It was run
on 2026-08-12 against `agent/cache` commit `1f225a0`, with Go 1.26.5 on
`darwin/arm64` Apple M4. The benchmark command was:

```sh
go test ./internal/cache \
  -bench 'Benchmark(CanonicalSemanticHash|EvidenceHash|CacheHit|SameKeyStampede)$' \
  -benchmem -benchtime=200ms -count=3
```

| Experiment | Observed result | Interpretation |
| --- | --- | --- |
| Canonical hash, 10 facts | 8.4–10.1 µs/op; 15,344 B/op; 81 allocs/op | Small closure overhead is measurable but bounded. |
| Canonical hash, 100 facts | 75.2–82.7 µs/op; 125,680 B/op; 624 allocs/op | Approximately linear growth at this size. |
| Canonical hash, 1,000 facts | 742–781 µs/op; 1,107,314 B/op; 6,027 allocs/op | Allocation cost is already material. |
| Canonical hash, 10,000 facts | 7.36–8.06 ms/op; 14,161,788 B/op; 60,031 allocs/op | Current reflection encoder is not yet a large-graph budget winner. |
| Evidence envelope hash | 1.0–1.17 µs/op; 1,952 B/op; 16 allocs/op | Evidence metadata is cheap relative to semantic closure hashing. |
| Warm exact cache hit | 40.6–46.1 µs/op; 12,453–12,470 B/op; 103 allocs/op | Includes filesystem metadata/content verification. |
| Same-key stampede, 16 workers, 1 ms compute | 14.0–14.8 ms/op; exactly 1.000 compute/op | Same-instance lock coalesces the miss, but `Clear` and publication dominate the batch. |
| Same-key latency, 100 rounds, 16 workers | p50 13.62 ms; p95 19.59 ms; max 26.00 ms; average computes 1.00 | Current local contention baseline; not a cross-process result. |
| Stale/corrupt recovery | 300/300 test invocations passed in 6.34 s wall time | Data and metadata tampering were rejected and recomputed. This is a correctness observation, not a latency SLO. |

The measurements use a temporary fixture and are not committed as production
tests. They must be reproduced by a future benchmark package with stable
machine-readable output before becoming a CI performance gate.

### Semantic versus evidence invalidation result

The four-case mutation fixture separates meaning invalidation from evidence
invalidation:

| Mutation | Semantic digest | Evidence digest | Required decision |
| --- | --- | --- | --- |
| Same canonical closure replay / display-only change | unchanged | unchanged | Cache may hit; existing evidence remains eligible. |
| Semantic fact added | changed | unchanged until evidence is regenerated | Projection key must miss; old evidence must be `stale`, never silently fresh. |
| Policy hash changed | unchanged | changed | Semantic projection may remain reusable, but verification evidence must be rechecked. |
| Evidence observation time changed | unchanged | changed | Projection may hit; evidence record is a new append-only observation. |

This is the intended split: evidence must bind the semantic/output digest and
its policy context, but an unchanged evidence record must not be rewritten just
because a cache object was recomputed. The current fixture records the expected
digest transitions; it does not yet implement a provenance graph or freshness
verifier, so those rows are contract evidence rather than feature completion.

### Acceptance budgets

Budgets are explicit gates for the future implementation. A measured value that
does not meet a budget is a performance or design failure to investigate, not a
reason to weaken correctness or accept stale output.

| Area | Acceptance budget | Current status |
| --- | --- | --- |
| Semantic invalidation | 100% of the mutation matrix produces the expected semantic/projection decision; 0 false exact hits. | Fixture transitions pass; projection dependency graph is not implemented yet. |
| Evidence freshness | 0 stale or policy-mismatched evidence accepted as fresh; every stale result has a machine-readable reason. | Hash transition fixture passes; freshness evaluator is not implemented yet. |
| Canonical determinism | 100% identical digests for the same vector on each supported OS/architecture; 100% expected digest changes for meaning-bearing mutations. | Native Darwin vector passes; Linux/Windows were compile-only, so runtime parity is unverified. |
| Canonical hash latency | p95 ≤1 ms for 1,000 facts and p95 ≤10 ms for 10,000 facts on the reference runner, with an allocation budget tracked separately. | Current 1,000/10,000 means are below targets; p95 and allocation reduction are not yet gated. |
| Warm exact hit | p95 ≤100 µs for a ≤64 KiB local projection on the reference SSD. | Current means are 40.6–46.1 µs; p95 is not yet emitted by the benchmark harness. |
| Same-key stampede | Exactly 1 compute/op for one process; 16-worker p95 ≤25 ms with a 1 ms compute; no wrong bytes. | 1.000 compute/op and 19.59 ms p95 observed locally; cross-process behavior is unverified. |
| Stale/corrupt entries | 100% of data, metadata, missing-file, and interrupted-publication faults become miss/recompute or an explicit error; 0 false accepts. | 300/300 repeated candidate tests pass. |
| Cross-platform publication | No partial visible object; supported-platform runtime tests pass; unsupported filesystem behavior is explicit. | Linux/Windows test binaries compile; they were not executed here. |
| GitHub Actions exact hit | 10/10 unchanged warm runs report an exact hit and pass manifest verification; 0 partial restores are treated as fresh. | No `.gooo` Actions cache job exists yet. |
| GitHub Actions value | For projections larger than 5 MiB, warm validated runs should be at least 30% faster than cold recomputation in 8/10 runs; otherwise retain the cache only as a measured optimization. | Not measured. A prior PR run only exercised `setup-go` cache and warned that `go.sum` was absent, so it is not a projection-cache result. |
| Cache safety | No secret or durable evidence in cache paths; untrusted restores are locally verified before use. | Design rule only; workflow fixture pending. |

### Experiment protocols

#### 1. Semantic/evidence mutation matrix

Use one fixed normalized closure and one evidence envelope. Apply one mutation
at a time: display name, stable ID, relation endpoint, policy digest, verifier
identity, observation time, toolchain, and output bytes. Record semantic hash,
projection key, evidence hash, freshness result, and stale reason. Run the same
matrix before and after each canonicalizer or verifier change. The hard gate is
zero false fresh results, not a minimum hash-change percentage.

#### 2. Lock contention and crash safety

Run 1, 2, 4, 8, 16, and 32 workers against one missing key and against disjoint
keys. Repeat 100 batches per worker count, with compute delays of 0 ms, 1 ms,
and 100 ms. Record p50/p95/p99 batch latency, lock wait, compute count,
publication errors, and byte equality. Repeat with separate OS processes to
expose the fact that a process-local mutex is not a cross-process lock. Kill a
writer at staging, after data sync, and after metadata sync; reopen the cache
and require either the old complete object or a clean miss.

The acceptance budget is one compute for same-key calls within one cache
instance, no partial objects, and no divergent bytes under one key. A
cross-process lease may reduce duplicate work, but duplicate work is safer than
unsafe stale-lock reclamation when the filesystem cannot provide a portable
lease.

#### 3. Stale-cache and evidence fault injection

For each cached object, independently mutate or remove data, metadata, the
dependency index, policy digest, semantic digest, and predecessor evidence.
Also leave a temporary directory or lock record behind and restart the reader.
The expected outcomes are `miss`, `corrupt`, or an explicit freshness failure;
never a fresh projection or fresh evidence result. Keep the original evidence
record append-only and emit a new recovery/verification record after rebuild.

#### 4. Cross-platform canonicalization

Publish a versioned vector set containing map order permutations, empty versus
absent values, Unicode strings without implicit normalization, signed zero,
float edge cases, URI/ID values, `time.Time`, and nested relation sets. Execute
the same vectors on Darwin, Linux, and Windows with the same Go and
canonicalizer versions, then compare digest files byte-for-byte. Compile-only
cross builds are useful smoke checks but do not satisfy this experiment.

The current candidate already exposes one policy decision: the same instant
represented as UTC and as a fixed KST location produces different digests. If
semantic timestamps mean instants, canonicalization must normalize to UTC; if
location is meaning-bearing, the distinction must be explicit in the IR. This
must be decided before declaring cross-platform canonicalization complete.
Go also documents that `os.Rename` is not atomic on every platform, so the
publication portion of this experiment must report filesystem support
separately ([`os.Rename`](https://go.dev/pkg/os/#Rename)).

#### 5. GitHub Actions cache hit/miss

The future workflow experiment should be a manually triggered, read-only-safe
matrix over Linux/macOS/Windows and the supported Go versions. For each cell,
run at least 10 cold keys, 10 unchanged warm keys, 10 semantic mutations, 10
policy/toolchain mutations, and 10 partial-restore cases. The primary key must
include OS, architecture, toolchain, cache schema, projection, semantic hash,
and dependency root. A restored manifest must contain the expected exact key
and content digest before the projection is used.

Record `cache-hit`, primary versus prefix match, lookup/download/upload time,
manifest validation, recomputation time, total job time, cache size, and whether
the run is trusted or low-trust. GitHub defines `cache-hit` for an exact key;
prefix/`restore-keys` matches are a different condition and can return the most
recent matching cache ([dependency caching reference](https://docs.github.com/en/actions/reference/workflows-and-actions/dependency-caching)).
Therefore a partial match may seed dependencies but must count as a semantic
miss. The current workflow's `setup-go@v5 cache: true` warning about absent
`go.sum` is recorded as an ordinary dependency-cache miss, not as evidence
about a future `.gooo` projection cache.

Workflow concurrency is also measured separately from cache locking. The
existing `cancel-in-progress: true` can suppress duplicate CI runs, but it
does not prove object publication safety. GitHub describes concurrency groups
as scheduler controls that cancel or queue runs, with at most one running job
per group by default ([workflow concurrency](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/control-workflow-concurrency)).

### Hosting transition contract

The cache/evidence contract must survive the transition from a Go-hosted
compiler to a future gooo-hosted compiler:

| Phase | Host contract | Required evidence |
| --- | --- | --- |
| Go-hosted initial | Go parser, canonicalizer, verifier, and generator are the executable host. | Host language/version, toolchain digest, canonicalizer schema, semantic vectors, and verifier identity. |
| gooo-hosted future | A `.gooo` compiler can parse/lower/build the same contract and, where applicable, host the next compiler stage. | Bootstrap parent artifact digest, host semantic hash, conformance-suite digest, toolchain/policy digests, and replay result. |

The transition gate is vector parity, not a label: both hosts must produce the
same canonical semantic hashes, invalidation decisions, freshness decisions,
and projection keys for the shared fixture set. The future host must also emit
a provenance edge back to the Go-hosted bootstrap artifact. No gooo-hosted
capability is claimed by this research document; its status remains planned
until an independent replay and evidence chain pass.

## Falsifiable hypotheses and minimum fixtures

The following hypotheses are deliberately stated so an implementation can
disprove them. A green unit test is evidence for one fixture, not proof that
the hypothesis holds for all IR shapes.

| ID | Hypothesis | Minimal falsifier | Pass condition | Current status |
| --- | --- | --- | --- | --- |
| H1 | A meaning-bearing semantic mutation changes `SemanticHash` and every affected projection key, while an evidence-only mutation does not change `SemanticHash`. | Add one relation endpoint, then change only evidence observation time. | Expected digest changes and unchanged values match the mutation matrix; 0 false exact hits. | Partially measured; dependency graph/freshness evaluator deferred. |
| H2 | Canonicalization is representation-independent for declared equivalences. | Reorder a map/set, format source differently, or rename a display-only alias. | Equal meaning yields byte-identical canonical bytes and digest on every supported host. | Map/order fixture passes; UTC/KST same-instant counterexample remains. |
| H3 | Every output-affecting dependency is in the key directly or through `DependencyRoot`. | Omit policy, toolchain, generator, or target from a test key, then mutate it. | The completeness test fails for the intentionally incomplete key and passes for the complete key. | Contract only; no dependency index implementation yet. |
| H4 | Same-key concurrent computation is safe even when lock scheduling changes. | Start 16 workers on one miss; kill one writer during publication; run separate processes. | No partial object, no divergent bytes, and one compute/op for one in-process cache. | In-process baseline passes; cross-process and kill-point runs deferred. |
| H5 | A stale or corrupt cache object can never become a fresh projection or evidence result. | Flip data bytes, metadata bytes, policy digest, predecessor evidence, or remove the index. | Result is miss/corrupt/stale with a reason, followed by explicit recompute or verification. | 300/300 local tamper/recovery runs pass; freshness graph deferred. |
| H6 | Invalidation is precise for disjoint semantic subgraphs. | Mutate `billing://` while querying an unrelated `fraud://` projection. | Unrelated exact hit remains valid; affected closure is recomputed. | Fixture contract only; reverse-edge index deferred. |
| H7 | GitHub Actions `restore-keys` partial matches are useful warm seeds but never semantic fresh hits. | Restore a prefix entry after a semantic or toolchain mutation. | `cache-hit`/manifest classification is partial or miss and forces exact validation/recompute. | Workflow experiment not executed; current CI only observes `setup-go` cache. |
| H8 | Go-hosted and future gooo-hosted stages can share the same cache/evidence contract. | Run the same vector set through two hosts and compare digests and decisions. | 100% vector parity plus a bootstrap provenance edge; otherwise the future host is deferred. | Future host is explicitly deferred. |

### Fixture manifest

Every fixture should have a stable ID and a small manifest that can be consumed
by tests, benchmark runners, CI, and future host implementations. The manifest
is input data, not an assertion hidden in a test body:

```text
Fixture {
  id: "cache-split-001"
  schema: "cache-contract-v1"
  semantic_closure: canonical facts and stable IDs
  dependency_edges: sorted direct semantic/policy/toolchain edges
  projection_request: artifact kind, projection, target, options
  evidence_chain: optional append-only predecessor records
  mutation: one named change or "base"
  expected: semantic hash/key/freshness/status/reasons
}
```

The minimum fixture set is:

| Fixture ID | Contents | Negative/counterexample variant | Expected output |
| --- | --- | --- | --- |
| `cache-split-001` | Two entities, one activity, one `uses` relation, one evidence record. | Add a relation; change only policy; change only observation time; rename display alias. | Semantic/projection/evidence digest transitions match H1; stale evidence is not fresh. |
| `canonical-vectors-001` | Map/set order, nil/empty, signed zero, float edge values, Unicode, URI IDs, `time.Time`, nested qualifiers. | Same instant in UTC and fixed KST; omitted policy default; duplicate canonical map keys. | Declared equivalences share a digest; ambiguous or unsupported values fail or are explicitly distinct. |
| `dependency-closure-001` | `billing` and `fraud` namespaces with disjoint projection roots and one shared policy node. | Mutate one namespace; delete one edge; introduce unknown dependency. | Only reverse closure invalidates; unknown dependency is miss/blocked, never fresh. |
| `lock-stampede-001` | One key, 16 workers, 1 ms compute, deterministic output. | 0/100 ms compute, 32 workers, separate processes, writer killed at each publication stage. | Compute count, p50/p95/p99, lock wait, byte equality, and recovery status are recorded. |
| `stale-object-001` | Complete data/metadata pair and append-only evidence. | Data flip, metadata flip, missing file, stale policy, stale predecessor, orphan temp directory. | `corrupt`, `miss`, or `stale` with reason; no false fresh result. |
| `actions-cache-001` | Exact key, older prefix key, semantic mutation, toolchain mutation, trusted and low-trust runs. | Treat prefix restore as exact hit; restore from an untrusted branch; exceed key schema. | Exact hit only after manifest verification; partial/low-trust paths recompute or remain read-only. |
| `host-parity-001` | Shared canonical vectors and expected contract version. | Go-hosted output differs from future gooo-hosted output or lacks bootstrap edge. | `PASS` only on full parity and evidence; current future stage remains `DEFERRED`. |

Fixtures must include a negative expected result. A fixture that can only pass
does not test the safety boundary. For example, `canonical-vectors-001` must
retain the UTC/KST case until the IR makes the timestamp policy explicit; the
current candidate's differing digest is a useful falsifier, not a flaky test.

## Shared implementation contracts

The cache contract is useful only if each compiler surface consumes the same
typed facts. These records are language-neutral names for the first Go
implementation; fields may be serialized as canonical JSON or the versioned
binary form, but their meaning and omission rules must stay stable.

### Input records

```text
CanonicalClosureInput {
  schema_version
  stable_roots[]
  entities[]
  activities[]
  relations[]
  explicit_assertions[]
  dependency_edges[]
  canonicalizer_version
}

ProjectionRequest {
  artifact_kind
  projection
  semantic_hash
  dependency_root
  policy_hash
  compiler_version
  toolchain
  target
  options
}

EvidenceFreshnessInput {
  subject_digest
  semantic_hash
  dependency_root
  policy_hash
  schema_version
  toolchain
  target
  verifier_identity
  predecessor_evidence_ids[]
  expires_at?
}
```

Rules:

- AST source spans and display names may be carried for diagnostics, but must
  not enter `CanonicalClosureInput` unless the requested artifact is a source
  map or diagnostic projection.
- `dependency_edges` are sorted and include edge kind and qualifiers; an
  unresolved edge is an explicit error/unknown state, not an empty list.
- Defaults are normalized before hashing. “Absent” and “empty” are equal only
  when the IR schema says they are semantically equal.
- `ProjectionRequest` contains every value that can alter output. A caller may
  not silently add ambient environment variables after key construction.
- Evidence freshness is evaluated against the current request and all required
  predecessors; a matching leaf digest alone is insufficient.

### Output records

```text
CanonicalClosureOutput {
  canonical_bytes_digest
  semantic_hash
  dependency_root
  normalized_facts[]
  direct_edges[]
  source_map_digest?
}

ProjectionResult {
  status: exact_hit | miss | partial_seed | corrupt | stale | error
  projection_key
  content_digest?
  bytes?
  manifest
  recomputed: boolean
  reasons[]
}

FreshnessResult {
  status: fresh | stale | invalid | deferred
  subject_digest?
  evidence_id?
  reasons[]
  checked_predecessors[]
}

VerificationDecision {
  status: pass | fail | deferred | blocked
  semantic_delta
  evidence_ids[]
  scope_result
  reasons[]
}
```

`partial_seed` is intentionally not an exact cache hit. `deferred` means the
requested check is not executable in the current host or fixture environment;
it must retain a command, reason, and required follow-up. `blocked` means an
external owner or dependency is required. Neither status may be rendered as
`pass` by a CI adapter or evidence publisher.

### Consumer mapping

| Consumer | Reads | Produces | Must preserve |
| --- | --- | --- | --- |
| AST/parser | `.gooo` source, spans, declarations | source facts and diagnostics | Stable IDs and authority annotations; spans remain diagnostic metadata. |
| Semantic IR | normalized facts, explicit assertions, dependency edges | `CanonicalClosureOutput` | Deterministic ordering, namespace isolation, candidate-vs-authoritative fact distinction. |
| BX/lifting | AST/Go symbol facts, stable semantic registry | semantic delta with `added/removed/changed`, authority, spans | No promotion of ambiguous calls to business truth without an assertion. |
| Codegen | `ProjectionRequest`, canonical closure, handwritten slots | `ProjectionResult`, generated-region manifest, source map | Locality, generated markers, exact key, content digest. |
| LSP | document version, semantic hash, projection/freshness status | diagnostics, cache hints, stale reasons | Never hide stale/blocked status behind a successful completion response. |
| Cache | `ProjectionRequest`, object bytes, metadata | `ProjectionResult` | Immutable publication, integrity verification, reconstructability. |
| Provenance | activity inputs/outputs, evidence predecessors | append-only `FreshnessResult` and evidence records | `used`/`wasGeneratedBy` links, no mutation of history, verifier identity. |
| CI/gate | semantic delta, scope, `VerificationDecision`, evidence | pass/fail/deferred/blocked decision and machine-readable report | Protected policy, no cache-only approval, no unimplemented-stage success. |

This mapping is the reuse contract for later AST, IR, BX, generator, LSP,
cache, provenance, and CI work. An implementation can add fields, but it must
not change the meaning of an existing status or omit a reason required by a
negative fixture.

## Pass, fail, deferred, and blocked semantics

Each experiment and implementation stage must publish one of four statuses:

| Status | Meaning | Evidence required | Merge implication |
| --- | --- | --- | --- |
| `PASS` | The fixture and acceptance budget ran and all hard invariants held. | Command, environment, fixture IDs, measurements, and output digest. | May be considered by the gate if scope and policy also pass. |
| `FAIL` | A falsifier or hard invariant was observed. | Counterexample, expected/actual output, reproduction command, and severity. | Must not be hidden by fallback or cache reuse. |
| `DEFERRED` | The check is not executable yet or a supported host is unavailable. | Explicit reason, exact follow-up command, owner, and no-success disclaimer. | Cannot satisfy a required gate by itself. |
| `BLOCKED` | Work depends on an external owner, branch, service, or policy registration. | Dependency, owner/thread, and unblock condition. | Do not weaken the gate; continue independent scope. |

Current classifications for this research are:

- `PASS`: local Go-hosted semantic/evidence mutation fixture, stale-object
  recovery fixture, in-process same-key compute-once check, and repository
  `gofmt`/vet/test/race checks.
- `FAIL`: the current candidate's same-instant UTC/KST digest equality
  hypothesis, because it produces different digests before timestamp semantics
  are normalized or declared location-bearing.
- `DEFERRED`: Linux/Windows runtime vector parity, cross-process lease behavior,
  the `.gooo` projection Actions cache job, the freshness evaluator, and the
  gooo-hosted stage.
- `BLOCKED`: CI ownership registration for `docs/research/cache.md`; it has
  been delegated to the CI workflow owner because no ownership alias exists in
  the current repository.

No status above claims that the unimplemented `cmd/gooo check` or future
gooo-hosted compiler exists. The status is part of the evidence contract and
must be checked by reviewers and automation.

## Phased implementation plan

1. **Contract fixtures**: add canonical IR fixtures and the C1–C5 mutation
   matrix before changing storage behavior.
2. **Canonical semantic hash**: implement typed, versioned normalization with
   stable-ID and namespace tests; keep source-map hashes separate.
3. **Dependency index**: record direct edges and a deterministic dependency
   root; add deletion, policy, and toolchain invalidation tests.
4. **Local storage**: publish immutable objects atomically, validate metadata
   and bytes, clean crash leftovers, and keep duplicate computation safe.
5. **Freshness verifier**: evaluate evidence over the predecessor graph and
   return machine-readable stale reasons.
6. **CI integration**: add exact-key Actions cache usage only after local
   validation; keep evidence and release artifacts in explicit artifact or
   attestation channels.
7. **Benchmark gate**: establish a baseline, then add regression thresholds
   for canonicalization, hit latency, invalidation precision, and stampede
   behavior without weakening correctness gates.

## Open risks

- Semantic equivalence is harder than byte canonicalization. Every new IR field
  needs an explicit “meaning-bearing or presentation-only” decision.
- Reflection-based canonicalization is convenient in Go but can accidentally
  expose implementation details or diverge from future non-Go projections. A
  schema-driven encoder is preferable for the stable IR boundary.
- Filesystem rename and locking semantics differ across platforms and network
  filesystems. Supported environments must be stated and tested explicitly.
- A cached verification result can be cryptographically intact yet semantically
  stale. Freshness must remain a graph predicate tied to policy and verifier
  identity.
- Remote cache contents are untrusted input. Every restored object needs local
  integrity and schema validation before it influences generation or CI.
- Garbage collection must account for provenance/evidence references, active
  leases, and in-flight builds; deleting an unreferenced cache object is safe,
  deleting durable evidence is not.

## References

- [Bazel Remote Caching](https://bazel.build/remote/caching)
- [GitHub dependency caching reference](https://docs.github.com/en/actions/reference/workflows-and-actions/dependency-caching)
- [`actions/cache` documentation](https://github.com/actions/cache)
- [GitHub dependency caching security concepts](https://docs.github.com/en/actions/concepts/workflows-and-actions/dependency-caching)
- [GitHub workflow concurrency](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/control-workflow-concurrency)
- [GitHub artifact attestations](https://docs.github.com/en/actions/how-tos/secure-your-work/use-artifact-attestations/use-artifact-attestations)
- [RFC 8785: JSON Canonicalization Scheme](https://www.rfc-editor.org/rfc/rfc8785.html)
- [W3C PROV constraints](https://www.w3.org/TR/prov-constraints/)
- [Go `os.Rename` documentation](https://go.dev/pkg/os/#Rename)
