# Architecture

The compiler is organized around explicit semantic boundaries. The `.gooo` view
is authoritative for business intent; the IR is the normalized interchange form;
Go projection is structural output; handwritten slots are the implementation
escape hatch. Evidence records describe what a build observed and do not become
new intent by themselves.

Self-hosting is not part of the current implementation contract. The checked-in
bootstrap fixtures are non-promoting evidence shapes; this document does not
define a self-hosted compiler or verifier authority, and it does not link to a
research note that is not present in this repository.

```text
main.gooo
   │ parse: syntax tree + source spans
   ▼
semantic IR ── normalize ──> stable nodes + PROV-inspired facts
   │                                  │
   │                                  ├── query/search projections
   │                                  ├── provenance/evidence records
   │                                  └── locality and CI decisions
   ▼
generated Go ── marker regions ──> handwritten implementation slots
```

## Package boundaries

The current implementation is dependency-free and keeps adapters at the edges:

- `internal/syntax` owns tokens, parsing, AST nodes, spans, and diagnostics;
- `internal/semantic` owns IDs, namespaces, PROV-inspired node/relation kinds,
  normalization, validation, and canonical fingerprints;
- `internal/bidir` owns parser-neutral `Get`/`Put`, fact layers, reconciliation,
  source requirements, deltas, locality, and BX checks;
- `internal/generator` owns deterministic Go projection, generated markers,
  handwritten slots, and source maps;
- `internal/analyzer` observes registered semantic Go symbols and emits facts;
- `internal/cache` stores reconstructable projections only, addressed by input and
  option digests;
- `cmd/gooo` is the command-line adapter.

These are internal package boundaries, not promises that every package has a
stable public API. The CLI is the supported user boundary. `analyze` and `lsp`
are not stable CLI commands, and no documentation should imply that they are.

## SSOT and provenance matrix

| Artifact | Authority | Allowed derived consumers | Write-back rule |
| --- | --- | --- | --- |
| `.gooo` declarations | Business intent, explicit IDs, names, contracts | IR, Go structure, queries, docs | Change here for intent; do not infer intent from helper Go |
| Semantic ID | Stable identity | Every projection and evidence record | Never replace an ID because a display name changed |
| Semantic IR | Normalized intermediate meaning | Generator, analyzer reconciliation, query, cache | Must remain semantically equivalent across projections |
| Handwritten Go slot | Irreducible implementation logic | Build and runtime behavior | Preserve outside regeneration; do not use it as structural SSOT |
| Generated Go region | Structural projection | Go compiler and tools | Regenerate from DSL/IR; never hand-edit |
| Go observation | Syntactic, candidate, or deterministic evidence | Reconciliation and review | Only source-backed accepted deterministic facts may update IR |
| Provenance/evidence | Append-only build record | CI, review, diagnostics | Add facts; do not silently rewrite source or prior evidence |
| Docs and CI | Governance and verification view | Contributors and reviewers | Describe actual behavior; update when workflow changes |

Source spans are part of the trust boundary. An analyzer observation without a
source span can remain syntactic or candidate evidence, but strict reconciliation
must reject it as a semantic update. Absence from a partial Go analysis is not an
implicit deletion; removals must be explicit.

## Semantic flow

1. Parse `.gooo` into a syntax tree while retaining source locations.
2. Lower declarations into stable nodes and the small PROV-inspired core. An
   activity signature derives `used` edges for inputs and `wasGeneratedBy` edges
   for its result; it does not invent domain-specific relations.
3. Normalize names, aliases, insertion order, facts, and canonical identity
   spelling. Presentation metadata may change without changing the semantic
   fingerprint.
4. Project structural Go with stable `//gooo:generated:*` markers and explicit
   handwritten slots. Preserve marker-outside text and slot bodies on regeneration.
5. If Go is analyzed, classify observations before reconciliation. Deterministic,
   source-backed facts may produce a semantic delta; ambiguous facts remain
   candidates; syntactic facts never change the model.
6. Record locality, source spans, hashes, and command results as evidence for the
   reviewer and CI.

## CI-only branch and promotion flow

The checked-in governance mode is `ci_only`. Work branches target `dev`; no
intermediary branch is a current route or promotion source. The only
promotion route is an exact same-repository pull request with `base=main` and
`head=dev`.

The six canonical proof jobs are `gofmt`, `go vet`, `go test`, `go test -race`,
`Semantic conformance`, and `CI policy`. Protected `dev` requires those six
contexts plus `CI guardian shadow`. Protected `main` requires those six plus
`CI guardian`. The Guardian context is app-bound and route-specific, so both
branches require exactly seven contexts.

For a promotion, the proof must bind an open, non-draft, unmerged, mergeable,
clean same-repository PR to the current `main` and `dev` refs. The live refs are
read before and after inspection and must remain unchanged. The topology must
be `ahead` with `ahead > 0`, `behind = 0`, and `merge_base_sha` equal to the
current main SHA. The six proof jobs, exact artifacts, Guardian evidence, and
both seven-context protection snapshots must also pass.

The proof emits a `promotion_authorization` containing the exact base/head
SHAs, `source=dev`, `target=main`, `operation=fast_forward`, and the proof bundle
digest. It is `PASS` only when all predicates above hold; otherwise it is
`FAIL_CLOSED` with a reason code. This authorization is pure evidence: the
proof producer never writes refs or branch protection. After a final exact
reread of refs and protection, an external gate may perform only a normal
compare-and-swap/fast-forward update of `main`; force-push and force-update
operations are not permitted.

The current CI workflow runs format, vet, unit-test, race-test,
semantic-conformance, policy, evidence, proof, and failure-report jobs. The
semantic check is runnable for the billing fixture. Richer relations, a
production LSP, a stable Go analysis CLI, cache conformance, durable evidence
publishing, and self-hosted verifier promotion remain unsupported until each
has an implemented entry point and runnable evidence.
