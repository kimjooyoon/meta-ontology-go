# Reusable contract ledger

This ledger separates contracts that are executable on `integration` from
contracts that are defined by research or an implementation branch. It is a
navigation layer for [architecture](architecture.md), [spec](spec.md), and
[governance](governance.md); it does not promote a proposal into a feature.

## Status vocabulary

| Status | Meaning | Evidence required |
| --- | --- | --- |
| `implemented` | Code and tests are present on `integration`. | A repository path and a runnable check. |
| `contract-only` | The reusable boundary is specified by a research or feature lane. | A conformance plan and an owning implementation PR. |
| `deferred` | The entry point or required verifier is intentionally unavailable. | A deterministic deferred result; never a promotion pass. |
| `not-run` | A proposed stage or fixture has no execution result. | The reason and the fallback stage. |

## Contract inventory

| Area | Integration contract | SSOT and trust boundary | Current status |
| --- | --- | --- | --- |
| AST | `internal/syntax` emits a recoverable AST, exact token text, half-open source spans, and stable diagnostics. | `.gooo` source is authoritative; syntax never invents semantic facts. | `implemented` |
| Semantic IR | `internal/semantic` normalizes IDs, namespaces, nodes, relations, facts, and evidence. | Stable IDs and declared `.gooo` intent are authoritative; IR is the comparison form. | `implemented` |
| PROV core | Nodes are Entity, Activity, or Agent. Relations are `used`, `wasGeneratedBy`, `wasDerivedFrom`, and `wasAssociatedWith` with checked directions. | Only the declared activity signature supplies implicit facts; names and helper calls do not infer domain meaning. | `implemented` |
| BX | Get/Put, Project/Lift, explicit deltas, locality, and no-deletion-by-absence are the target lens contract. | The DSL owns intent; a Go observation must be source-backed and accepted before reconciliation. | `contract-only` |
| Codegen | Stable-ID regions, handwritten slots, deterministic imports, source maps, locality, and atomic replacement are the projection contract. | Generated Go is disposable structure; slot bodies are complement data; the DSL/IR owns structure. | `contract-only` |
| LSP | Framed JSON-RPC, lifecycle, versioned document snapshots, UTF-16 ranges, diagnostics, and identity-aware features are the research boundary. | Open buffers are views; semantic IDs and source snapshots remain authoritative. | `deferred` |
| Cache | A cache may store only reconstructable projections, addressed by normalized input and option digests, with integrity and atomic visibility. | Cache metadata is evidence, not SSOT; unknown dependencies are misses. | `deferred` |
| Evidence/CI | `gooo/evidence/v1`, canonical producer-independent payloads, manifests, and stage-0 Go verification are implemented. | Protected CI and the Go verifier decide; a candidate cannot certify itself. | `implemented` / Stage 0 |
| Self-hosting | Go-hosted bootstrap, dual evidence, fallback, promotion, and rollback are staged in the CI plan. | The Go verifier is the current trust anchor; gooo-hosted output is candidate evidence. | Stage 1+ `deferred` |

The absence of an `internal/generator`, `internal/lsp`, or `internal/cache`
package on `integration` is intentional. Their research contracts may guide
future implementation, but they are not current APIs or CI guarantees. The
same rule applies to the `check`, `generate`, `analyze`, and `lsp` command
names: the current `cmd/gooo` adapter is a stub and those commands are
deferred.

## AST and semantic boundary

The syntax contract is parser-facing and source-preserving:

- `Span` is half-open, uses UTF-8 byte offsets, one-based line/column values,
  and preserves an optional filename.
- Tokens retain exact source spelling in `Text`/`Lexeme`; string `Value` is
  decoded. A lexer result ends with one EOF token.
- The parser retains package, namespace, ordered declarations, activity input
  order, result references, and declaration/name/ID spans.
- Recovery may return a partial tree. Diagnostics have stable codes, severity,
  deterministic messages, and spans; parser errors do not become semantic
  facts.

The semantic boundary consumes that AST and establishes the following invariants:

- an ID is an absolute, canonicalizable URI-like identity; a namespace is a
  separate, case-sensitive scope;
- display names and aliases are lookup metadata and cannot change identity;
- node kind and namespace cannot change under an existing ID;
- name lookup is namespace-qualified and collisions fail closed;
- normalization and canonical fingerprints are deterministic and idempotent;
- a missing source span can remain diagnostic/synthetic evidence, but strict
  accepted semantic updates require source-backed evidence.

## PROV and evidence boundary

The current PROV-inspired profile is deliberately smaller than PROV-O. The
activity signature deterministically creates `Activity used Entity` for each
input and `Entity wasGeneratedBy Activity` for its result. It does not infer
derivation, association, attribution, authorization, or runtime execution.
Candidate observations remain separate from deterministic facts.

`internal/semantic` records append-only evidence with stable evidence IDs,
producer IDs, evidence kind, fact status, SHA-256 digest, and span. Semantic
equivalence ignores producer and source-location metadata; provenance
equivalence compares the normalized cross-host claim. Reusing an evidence ID
with different content is a conflict.

`internal/verify` provides the CI comparison envelope:

```json
{
  "producer": "go",
  "bundle": {
    "schema": "gooo/evidence/v1",
    "stage": 0,
    "fixture": "examples/bootstrap/main.gooo",
    "decision": "deferred",
    "facts": []
  }
}
```

The canonical comparison payload is the normalized `bundle` plus a newline;
`producer` is validated but excluded from that payload. Facts sort by
`id/kind/value` and duplicate fact IDs fail closed. The derived manifest binds
the payload SHA-256 to its producer, stage, fixture, and decision. A deferred
or not-run decision is valid evidence of unavailable capability, never a pass.

## Research contracts that are not yet executable

The following are portable rules distilled from the bidirectional, codegen, LSP,
cache, and PROV research lanes. The owning lane must add implementation and
runnable conformance evidence before changing the status above.

### BX and code generation

- Treat DSL lowering, Go projection, Go lifting, and reconciliation as separate
  operations. A partial Go view yields an explicit add/remove delta; absence is
  not deletion.
- Accepted deterministic facts need a source span or an explicitly authorized
  trusted adapter. Ambiguity, unknown IDs, unsupported relations, and conflicts
  fail transactionally without a partial model update.
- Compare semantic graphs by stable IDs, kinds, namespaces, directed relations,
  and semantic attributes. Preserve source complement and evidence separately.
- Generated regions and slots are owned by stable IDs. Preserve slot bodies and
  marker-outside text, preserve declared activity port order, reject malformed or
  duplicate markers, and validate the complete generated package before an
  atomic replacement.

### LSP

- Content-Length counts UTF-8 bytes, requests receive one response, and
  notifications receive none. Lifecycle state must reject or drop pre-init work
  deterministically and require shutdown before exit.
- Document changes are versioned snapshots. Diagnostics replace the prior
  publication, empty diagnostics clear it, and stale snapshots cannot publish
  after a newer change.
- UTF-16 position conversion is shared and must handle non-ASCII text, CRLF,
  invalid boundaries, and EOF without splitting UTF-8.
- Completion, definition, and hover must resolve through namespace-qualified
  stable IDs; display-name coincidence is not cross-file identity.

### Cache

- Hash normalized semantic meaning and all output-affecting inputs, including
  options, policy/schema, toolchain, target, and dependency closure. Presentation
  changes with unchanged IDs should not invalidate the semantic hash.
- A hit requires matching key, metadata, content digest, size, and reconstructable
  artifact type. Corrupt, stale, or uncertain entries are misses or verification
  failures, never fresh hits.
- Publish complete objects atomically, serialize same-key computation, preserve
  append-only provenance outside the cache, and garbage-collect only unreferenced
  unlocked objects after retention policy permits it.

## CI and bootstrap contract

The current workflow runs format, vet, unit test, race test, semantic
conformance, and policy jobs. Semantic conformance pins stage 0, runs the Go
verifier, and exits successfully only after printing an explicit deferred status
when `cmd/gooo check` is still a stub. Stages 1, 2, and 3 are rejected by the
script until a reviewed CI change enables them.

The bootstrap path is monotonic:

```text
Go-hosted seed -> dual evidence -> gooo candidate with Go fallback
              -> gooo authority only after independent comparison and rollback
```

The self-hosting workstream owns the history in
`docs/research/self-hosting.md`; this ledger owns the integration boundary. At
every stage, `.gooo` intent and stable IDs remain SSOT, generated Go/cache/
manifests/evidence remain derived, and the prior accepted verifier remains the
rollback authority. See [bootstrap-evidence.md](bootstrap-evidence.md) and the
[staged CI plan](../.github/conformance-plan.md).
