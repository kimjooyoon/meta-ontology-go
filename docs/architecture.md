# Architecture

The compiler is organized around explicit semantic boundaries. The `.gooo` view
is authoritative for business intent; the IR is the normalized interchange form;
Go projection is structural output; handwritten slots are the implementation
escape hatch. Evidence records describe what a build observed and do not become
new intent by themselves.

The proposed history from a handwritten seed to a self-hosted compiler is tracked
in the [self-hosting research note](research/self-hosting.md), owned by the
self-hosting-bootstrap workstream. This document connects that research to the
normative architecture without making the research note, or a future self-hosted
verifier, authoritative by itself.

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

## Bootstrap path to self-hosting

Self-hosting is a staged trust transition, not a single generator feature. The
following sequence is the proposed bootstrap history; a stage is current only
after its evidence is present in protected CI.

| Stage | Implementation shape | Trust boundary | Promotion evidence |
| --- | --- | --- | --- |
| Seed | Handwritten Go kernel plus a small `.gooo` fixture | Go parser, IR, generator, and verifier are trusted; DSL is intent input | Reproducible Go checks and the example semantic check |
| Semantic mirror | `.gooo` describes the compiler ontology, contracts, and verifier vocabulary | Go remains authoritative for execution; `.gooo` is a reviewed semantic mirror | Stable IDs, source spans, canonical IR, and BX tests agree |
| Structural self-host | `.gooo` drives structural Go for compiler components; handwritten slots retain irreducible logic | Generated Go is replaceable output; the seed remains the rollback implementation | Marker/locality checks, deterministic regeneration, and build equivalence |
| Shadow verifier | A `.gooo`-described verifier runs beside the seed verifier | Candidate verifier is evidence only and cannot approve itself | Seed/candidate decision and evidence streams match on pinned fixtures |
| Promoted self-host | The promoted verifier checks the next bootstrap and can rebuild the compiler | Previous verifier and source digest remain the rollback authority | Reproducible bootstrap, BX/provenance/locality gates, and independent approval |

At every stage, `.gooo` remains the SSOT for declared intent and stable IDs;
semantic IR remains the normalized comparison form; generated Go remains derived;
and provenance remains append-only evidence. A bootstrap artifact cannot grant
itself authority merely because it generated or verified itself.

The research note owns experiment history, alternatives, and stage-specific
measurements. This architecture owns the boundary rules that any implementation
must satisfy before those measurements can support promotion.

The comparable evidence shape and paired fixtures are defined in the
[bootstrap evidence bridge](bootstrap-evidence.md). It is the contract between
the Go-hosted seed and a future gooo-hosted candidate; it does not promote the
candidate or replace the research note.

## Current CI baseline

The integration workflow in [`.github/workflows/ci.yml`](../.github/workflows/ci.yml)
currently has separate format, vet, unit-test, race-test, semantic-conformance,
and CI-policy jobs. The semantic job delegates to
[`scripts/semantic-conformance.sh`](../scripts/semantic-conformance.sh); its
explicit deferred path is evidence that the CLI is not yet available, not proof
of self-hosting. The policy job checks branch ownership, Go size caps, and the
integration PR target. These checks protect the seed and candidate boundaries;
they are not yet the self-hosted verifier promotion gate.

## What is not part of the current contract

The architecture leaves room for richer relations, a production LSP, a stable Go
analysis CLI, durable evidence publishing, and self-hosted verifier promotion.
Those are design directions only until they have a supported entry point and
runnable conformance evidence. The current required CI is documented in
[CONTRIBUTING.md](../CONTRIBUTING.md) and the integration workflow. Self-hosting
and verifier promotion are future capabilities until the staged gates above are
wired into protected CI.
