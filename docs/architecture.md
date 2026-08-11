# Architecture

The compiler is organized around explicit semantic boundaries. The `.gooo` view
is authoritative for business intent; the IR is the normalized interchange form;
Go projection is structural output; handwritten slots are the implementation
escape hatch. Evidence records describe what a build observed and do not become
new intent by themselves.

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

## What is not part of the current contract

The architecture leaves room for richer relations, a production LSP, a stable Go
analysis CLI, durable evidence publishing, and broader CI gates. Those are design
directions only until they have a supported entry point and runnable conformance
evidence. The current required CI is documented in
[CONTRIBUTING.md](../CONTRIBUTING.md).
