# Reflective query sandbox

This is a deliberately narrow vertical slice: a Gooo program describes the
subjects it may query, the prior state of its claims, the metric relation, and
the mutation contract. The producer parses and lowers that exact source,
projects the canonical semantic IR into the existing read-only query graph,
and records what each API actually returned.

The source IDs are semantic coordinates, not comments or Go-side fixtures:

- `/subject/query`, `/subject/structure`, `/claim-state/open`, and
  `/metric/relation/read-only` are the query subject and query targets;
- every `/claim/<class>/<name>/<predicate>/<proof>/<meta-operation>/<evidence>/<prior>`
  entity is a claim with a formal predicate, prior state, and evidence attempt;
- `/mutation/request`, `/mutation/field`, `/mutation/payload`,
  `/mutation/intent`, and `/mutation/locality` are the source-declared typed
  mutation contract;
- activity signatures create the authoritative `used` relations that the
  producer queries. The `ReflectClaims` spelling alone is never evidence.

The source currently contains 12 claims, distributed by its own ID
coordinates into three classes and three proof choices. The producer computes
the denominator, class/proof buckets, source node/fact counts, attempt counts,
and claim-transition count from the parsed IR. No Go constant table supplies
those values.

## Boundary observations

The producer makes four exact read-only queries from source-declared
activities, one real mutation-capable request, one net repository-status observation,
and one unknown-target query. The mutation request calls
`semantic.Graph.ApplyGraphPatch` against the semantic graph using field,
payload, intent, and locality derived from the source. The baseline `id` field
produces a typed immutable-field rejection only when that API returns it; there
is no `DENIED` constant shortcut. The artifact captures semantic/graph digests
immediately before the call, the original IR after the call, and the returned
graph separately. A typed rejection discharges only the scoped
`immutable_id_patch_accepted=false` fact. A successful detached `name` patch
proves detached graph-patch capability, not repository or global mutation
authority; overall authority remains `UNKNOWN`.

The query decision and subject resolution are separate fields. A missing
endpoint is `UNKNOWN / LOWER_RESOLUTION / UNKNOWN_TARGET`, with stage
`UNKNOWN`, step `resolve-unknown-subject`; its claim remains `OPEN` with reason
`UNKNOWN_PRESERVED`. An exact relation is `PASS / EXACT`. A query API error is
an observation failure (`UNKNOWN / LOWER_RESOLUTION`) and does not become a
semantic contradiction. A boundary violation is `REFUTED`.

Subject identity is also split: `format` reports `FORMAT_VALID` independently
from `checkout`, which reports `CHECKOUT_BOUND` only when producer-shell
evidence records the actual checkout HEAD and the independent consumer matches
that evidence to its own checkout. A valid but stale SHA is therefore
`FORMAT_VALID` plus `SUBJECT_SHA_CHECKOUT_MISMATCH`, not a bound subject.

Each claim gets two append-only events: its source-declared prior state, then
an evidence-dependent event. Every claim has an explicit predicate evaluator
and observed material digest. Only its named attempt and predicate may move it
to `DISCHARGED`; missing or API-error evidence retains `OPEN`; explicit
contradictions move it to `REFUTED`.

The shell captures tracked and untracked `git status --porcelain` lines before
and after the producer. Those raw snapshots and only their net difference are
embedded in the observation and copied into the receipt. The repository
observation is `UNOBSERVED` without both evidence files,
`net_repository_status_unchanged` for equal normalized snapshots, and
`net_repository_status_changed` for an observed difference. Equal porcelain
snapshots do not prove zero transient writes.

## Independent consumer

The consumer binary has its own wire structs and imports no producer package.
It directly performs `syntax.ParseFile -> bidir.Lower -> query.FromSemanticIR`
on the raw `.gooo`, recomputes source and query digests, calls the same typed
mutation boundary, rebuilds attempts and predicates, verifies the provisional
observation digest and complete transition chain, and only then emits a sealed
receipt. The producer cannot seal its own material; the consumer adds a
separate attestation digest.
CI checks raw `go list -deps` evidence as
`forbidden_imports_observed=0 <= maximum_allowed=0` and reports the successful
independence coordinate as `1 / 1`.

The receipt reports dynamic indicator numerator/denominator, source
reconstruction numerator/denominator, and reflective-query
numerator/denominator. The unknown-target claim intentionally remains open:
unknown is a lower-resolution subject result, not a successful discharge.

## Interventions

CI also publishes a JSON intervention artifact. A semantic intervention
changes the source-declared mutation field from `id` to `name` and its payload
to `intervened-name`. The actual API outcome changes from typed `REJECTED` to
`ACCEPTED`; detached graph-patch capability changes from `NOT_OBSERVED` to
`OBSERVED` while overall authority remains `UNKNOWN`, and the scoped mutation claim changes
from `DISCHARGED` to `REFUTED`. A nonsemantic intervention changes only a
comment; the raw source digest changes while the semantic/graph digests,
mutation outcome, scoped capability, and claim state are preserved.
CI also fixes a three-case regression matrix: stale-but-well-formed subject
SHA, changed repository before/after evidence, and missing repository evidence.

## Research decisions

- [Go `reflect` documentation](https://pkg.go.dev/reflect) separates runtime
  views from mutation through explicit settable values. Adopted: detached
  identity/value observations and an explicit mutation boundary. Rejected:
  arbitrary `Set`/`SetMapIndex`/`MakeFunc`-style runtime mutation.
- [OCaml 5.2 effect-handler manual](https://ocaml.org/manual/5.2/effects.html)
  makes named effects visible at an enclosing handler boundary and fails on
  unhandled effects. Adopted: named meta-operations and visible unknown/
  unhandled outcomes. Rejected: continuations, scheduling, and execution of
  runtime effects.
- [Koka language book](https://koka-lang.github.io/koka/doc/book.html)
  motivates explicit effect typing and handlers. Adopted: effect-shaped
  receipts. Rejected: inferred effect polymorphism or an effect algebra not
  implemented by this Gooo slice.

## Falsification

The artifact is falsified if a read query changes semantic or graph digest, if
the source-declared immutable mutation is accepted without a `REFUTED` result,
if an error is mislabeled as an exact rejection, if the original graph changes
or the returned graph is omitted from mutation evidence, if unknown becomes
exact/pass, if a claim discharges without its named predicate/attempt, if
transition order or digests change, if repository before/after differs, if a
missing or stale subject binding becomes `CHECKOUT_BOUND`, if repository
evidence is missing or changed but is labeled unchanged, or if the independent
consumer cannot reconstruct the raw source.

This does not claim general Go reflection equivalence, hostile-process safety,
compiler/source completeness beyond the declared claims, or runtime
memory/performance bounds.
