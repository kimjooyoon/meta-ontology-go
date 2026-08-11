# Project governance contract

This document is the compact contract for changing the semantic compiler. It
defines source-of-truth (SSOT) boundaries, provenance policy, bidirectional (BX)
laws, agent roles, review caps, and evidence requirements. It is intentionally
more conservative than the future design: an internal type or comment is not a
supported feature until a user-facing entry point and runnable evidence exist.

## 1. SSOT boundaries

| Concern | Single source of truth | What other views may do |
| --- | --- | --- |
| Business intent | `.gooo` declarations and explicit assertions | Lower, project, query, and document it |
| Semantic identity | Stable URI-like IDs | Display names and aliases may refer to it |
| Normalized meaning | Semantic IR for the current compilation | Project it and compare semantic fingerprints |
| Irreducible logic | Handwritten Go slots | Call it from generated structure; preserve it |
| Structural Go | Generated regions | Compile and inspect it; never edit it directly |
| Observations | Parser/analyzer evidence with source spans | Classify as syntactic, candidate, or deterministic |
| Build history | Append-only provenance and verification records | Explain, review, and gate the build |
| Policy | Ontology vocabulary, verifier semantics, and CI workflow | Enforce it; do not weaken it locally |

Three boundaries prevent accidental authority inversion:

1. A display rename cannot change an ID or merge two namespaces.
2. A generated file cannot become a new intent source. Structural changes go to
   the DSL or generator owner and are regenerated.
3. A Go observation cannot become semantic truth without an explicit, accepted,
   source-backed delta. Ambiguous observations remain candidates.

The semantic IR is an interchange representation, not a second business SSOT.
Its normalized form makes projections comparable; it does not authorize a tool to
invent domain meaning. Provenance is evidence, not a write-back channel.

## 2. Provenance policy

Every semantic delta must answer four questions:

- what stable subject/predicate/object triple changed;
- which source view produced the observation;
- which source span or equivalent evidence locates it;
- why the fact is deterministic rather than merely plausible.

Strict reconciliation rejects a semantic addition, removal, or change without a
source span. A syntactic fact may be retained for diagnostics. A candidate fact
may be retained for review. Neither changes the semantic model. A deterministic
fact with the same triple shadows a candidate; it does not erase the historical
observation from the evidence record.

Analyzer snapshots are often partial. Therefore absence is not deletion. A
removal must be represented explicitly in the fact delta, and reconciliation must
be transactional: if one fact conflicts, the model remains unchanged and the
conflict is reported with its kind and source evidence.

## 3. BX laws

Let `s` be a parser-neutral DSL document, `m` a semantic model, `Get(s)` the
lowering function, and `Put(s, m)` the representable write-back. `≈` means
semantic equivalence after normalization; it does not mean byte-for-byte text
identity.

### Get-Put

```text
Put(s, Get(s)) ≈ s
```

Reading a source view and writing it back must not create a new semantic fact,
change a stable identity, or drop an unrelated declaration. Formatting and
ordering may be normalized.

### Put-Get

```text
Get(Put(s, m')) ≈ m'
```

For an accepted and representable semantic update `m'`, the next source read must
show that update. Unrepresentable or unproven updates are rejected rather than
silently approximated.

### Semantic round-trip

```text
s ──Get──> m ──project/lift──> m'
m' ≈ m
```

Generating a structural Go projection and lifting its facts must not create a new
semantic relation. Generated facts are already represented; only an accepted
source-backed delta may change the model.

### Locality

An implementation-only edit, or a change to one semantic region, must not rewrite
unrelated semantic nodes, generated regions, marker-outside handwritten text, or
handwritten slot bodies. Stable IDs define the comparison boundary.

### Normalization and provenance

Normalization is deterministic and idempotent: repeated normalization has the
same semantic fingerprint. Every accepted semantic delta carries provenance, and
failed reconciliation is transactional. These are guard laws for the four
round-trip laws above.

## 4. Generated boundaries

Generated Go uses stable markers such as:

```go
//gooo:generated:start id="billing://activity/pay-order" kind="activity"
//gooo:slot:start id="billing://activity/pay-order/implementation"
//gooo:slot:end id="billing://activity/pay-order/implementation"
//gooo:generated:end id="billing://activity/pay-order"
```

The generator owns text between generated markers. The implementation slot is
the only intentional handwritten region inside that boundary. Regeneration must
preserve slot bodies, reject malformed or duplicate markers, and keep unrelated
text stable. Generated output is not a place to fix a source-model problem.

## 5. Agent roles and separation of duties

- **Builder:** changes the assigned authority view, keeps the diff within scope,
  and supplies tests or runnable examples.
- **Guardian:** verifies semantic scope, stable IDs, provenance, BX laws, marker
  integrity, and evidence freshness; they do not implement the feature.
- **Approver:** makes the final acceptance decision after the Guardian's review
  and required CI checks.
- **Docs/example Builder:** owns only `docs/**`, `examples/**`, and the root
  governance Markdown files for documentation work. Core package source belongs
  to its implementing Builder.

No single agent should implement a change, weaken its verifier, and approve it.
When a worktree contains another Builder's uncommitted source, stage only the
explicit paths owned by the current task.

## 6. Branch, PR, and CI workflow

Use `agent/<area>` branches, one semantic concern per PR. A documentation change
uses `agent/docs`. Inspect status before editing, keep unrelated work unstaged,
push with tracking, and open a draft PR unless the owner requests ready-for-review.
The PR body should state:

- the authority view changed and the SSOT boundary it respects;
- affected IDs or example contracts;
- generated/provenance implications;
- checks run and any known environmental blocker;
- unsupported features deliberately not claimed.

The current GitHub Actions workflow runs:

```text
Go test job: gofmt -l .; go test ./...
Go vet job:  go vet ./...
Semantic job: go run ./cmd/gooo check examples/billing/main.gooo
```

The semantic job waits for the Go job. The workflow does not currently enforce
race tests, static analysis, generated-output snapshots, cache conformance, LSP
behavior, or durable provenance publishing. Do not describe those as CI gates
until the workflow changes and has passed evidence.

## 7. Review line caps

These are soft review policy, not language or compiler limits, and are not
machine-enforced today:

- ordinary source and Markdown lines: 120 columns;
- one handwritten implementation slot: 40 non-blank lines;
- one normal PR: 400 changed lines, excluding generated output;
- generated output: no manual line cap; its cap is marker integrity and
  deterministic regeneration.

URLs, tables, generated markers, and mechanically formatted output may exceed the
soft column cap. If a change exceeds a review cap, explain the exception and split
the evidence by authority boundary where possible.

## 8. Evidence checklist

A change is ready for review when the author can point to:

- a minimal diff scoped to the owning paths;
- the relevant source example or regression test;
- stable IDs and any semantic delta;
- source spans/provenance for accepted observations;
- generated-region and locality evidence when projection changes;
- local commands and their results;
- a clear statement of what remains unsupported.

The example commands in [conformance.md](conformance.md) are the smallest
repeatable evidence for the checked-in DSL fixtures.
