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

The design-only [deterministic metrics RFC](metrics-rfc.md) defines how repeated
reasoning may become exact predicates. It does not activate new blocking checks;
the contracts in this document and the live CI policy remain authoritative.
The [code-semantic coupling contract](code-semantic-coupling.md) adds the
docs-owned receipt and typed-path design; its two proposed metrics remain
`DESIGN_ONLY`, `UNOBSERVED`, and `NO_EFFECT`.

## 2. CI-only branch and promotion contract

The checked-in governance mode is `ci_only`. Reviewer identity, approval actors,
and last-push approval fields are not CI proof inputs. The six machine checks and
the exact source/ref evidence below decide promotion.

Work branches use `agent/* -> dev`. Promotion is a same-repository pull request
with `base=main` and either `head=dev` or a narrowly bound snapshot branch named
`agent/main-promotion-snapshot-<dev-sha>`. A snapshot must have the exact live
`dev` tree and one parent equal to the live `main` SHA. It carries no independent
source authority; it makes the exact `dev` tree a fast-forward child of `main`
when older history diverged.

The six canonical proof jobs are:

```text
gofmt | go vet | go test | go test -race | Semantic conformance | CI policy
```

The live `main` protection rule requires exactly these six contexts; `dev` has
no required status contexts. The retired `CI guardian` context, its duplicate
workflow, and its app-bound evidence path are absent from the active CI route.

An exact promotion proof requires all of the following:

1. The PR is open, non-draft, unmerged, mergeable, clean, and binds the same
   repository's `main` base to either `dev` or the exact snapshot branch name
   derived from the live `dev` SHA.
2. The live `dev` and `main` refs are reread during inspection. A direct `dev`
   candidate must itself be a fast-forward. A snapshot candidate must have the
   exact live `dev` tree and a sole parent equal to live `main`. In both forms,
   topology is `ahead`, with `ahead > 0`, `behind = 0`, and `merge_base_sha`
   equal to live `main`.
3. The six canonical jobs and their current immutable artifact pass. The same
   CI proof does not request a second observer to approve its branch-policy view.
4. The proof contains a digest-bound `promotion_authorization` with
   `operation=fast_forward`, `source=dev`, `target=main`, exact base/head SHAs,
   the source and candidate tree digests, and `proof_digest` equal to the proof
   bundle digest.

The authorization is pure, non-mutating evidence. It is `PASS` only when every
predicate holds and is otherwise `FAIL_CLOSED` with a reason code. The proof
producer does not write refs or branch protection. GitHub's native required
status checks enforce configured contexts; CI does not duplicate them by
reading a privileged branch-protection snapshot. Once installed on the default
branch, the automatic promotion executor opens or reuses the exact `main` PR,
dispatches CI with the PR number and observed head/base SHAs, and merges through
GitHub's PR API with the expected head SHA only after the proof and required
checks pass. The original successful `dev` run waits for one uniquely identified
new dispatch run, bounded to 40 minutes. A dependent job consumes that run's
exact current attempt and proof, then checks the live PR, refs, source tree and
topology again. Missing, ambiguous, failed or stale observations stop promotion.
The flow does not require a subsequent dispatched-CI completion callback.
Branch protection stays active; force-push and force-update
operations are never permitted.

The bootstrap fixtures and [bootstrap evidence bridge](bootstrap-evidence.md)
record non-promoting evidence shapes only. Self-hosting and a self-hosted
verifier are not current supported authorities, and this contract does not rely
on a separate research note.

## 3. Provenance policy

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

## 4. BX laws

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

## 5. Generated boundaries

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

## 6. Agent roles and separation of duties

The system evaluates the exact source revision, scope, stable IDs, provenance,
BX laws, marker integrity, and available evidence. Review identities and a
separate reviewer/approver lane do not affect the CI result. Documentation and
example ownership remains scoped to `docs/**`, `examples/**`, and the root
governance Markdown files.

## 7. Branch, PR, and CI workflow

Use `agent/<area>` branches, one semantic concern per PR. A documentation change
uses `agent/docs`. Inspect status before editing, keep unrelated work unstaged,
push with tracking, and open a draft PR unless the owner requests ready-for-review.
The PR body should state:

- the authority view changed and the SSOT boundary it respects;
- affected IDs or example contracts;
- generated/provenance implications;
- checks run and any known environmental blocker;
- unsupported features deliberately not claimed.

The current workflow runs these canonical jobs:

```text
format:   gofmt -l .
vet:      go vet ./...
test:     go test ./...
race:     go test -race ./...
semantic: go run ./cmd/gooo check examples/billing/main.gooo
policy:   go run ./scripts/verify
```

Evidence, proof, and failure-report jobs consume those six results. The semantic
check is a runnable billing-fixture check. Static analysis, LSP behavior, cache
conformance, generated-output snapshots, and automatic durable provenance
publishing are not current guarantees.

The six jobs are both the proof core and the complete required-context set for
`dev` and `main`. A read-only CI observer records their runtime spread in the
`gate_bottleneck` field. Missing, stale, duplicate, or sub-second intervals are
reported as `UNKNOWN`; the metric is descriptive and does not gate promotion.

The [bootstrap evidence bridge](bootstrap-evidence.md) defines non-promoting
fixture states such as `deferred`, `not-run`, and
`promotion_eligible: false`. Those states never authorize a branch promotion or
make a self-hosted verifier a supported authority.

## 8. Review line caps

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

## 9. Evidence checklist

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
