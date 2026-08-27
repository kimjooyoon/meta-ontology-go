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
activities, one real mutation-capable request, one repository-net observation,
and one unknown-target query. The mutation request calls
`semantic.Graph.ApplyGraphPatch` against the semantic graph using field,
payload, intent, and locality derived from the source. The baseline `id` field
produces a typed immutable-field rejection only when that API returns it; there
is no `DENIED` constant shortcut. The artifact captures semantic/graph digests
immediately before the call, the original IR after the call, and the returned
graph separately. If the API accepts the request, the attempt is `REFUTED`,
mutation authority is true, and related claim transitions become `REFUTED`.

The query decision and subject resolution are separate fields. A missing
endpoint is `UNKNOWN / LOWER_RESOLUTION / UNKNOWN_TARGET`, with stage
`UNKNOWN`, step `resolve-unknown-subject`; its claim remains `OPEN` with reason
`UNKNOWN_PRESERVED`. An exact relation is `PASS / EXACT`. A query API error is
an observation failure (`UNKNOWN / LOWER_RESOLUTION`) and does not become a
semantic contradiction. A boundary violation is `REFUTED`.

Each claim gets two append-only events: its source-declared prior state, then
an evidence-dependent event. Every claim has an explicit predicate evaluator
and observed material digest. Only its named attempt and predicate may move it
to `DISCHARGED`; missing or API-error evidence retains `OPEN`; explicit
contradictions move it to `REFUTED`.

The shell captures tracked and untracked `git status --porcelain` lines before
and after the producer. Those raw snapshots and only their net difference are
embedded in the observation and copied into the receipt. The baseline target
is `net_repository_changes=[]` and `mutation_authority=false`; this does not
claim a write count or absence of transient writes.

## Independent consumer

The consumer binary has its own wire structs and imports no producer package.
It directly performs `syntax.ParseFile -> bidir.Lower -> query.FromSemanticIR`
on the raw `.gooo`, recomputes source and query digests, calls the same typed
mutation boundary, rebuilds attempts and predicates, verifies the receipt
material digest and full observation digest, and only then emits a receipt.
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
to `intervened-name`. The actual API outcome changes from typed
`REJECTED`/authority false to `ACCEPTED`/authority true; the returned graph
digest changes while the original graph digest remains stable, and the
mutation claim changes from `DISCHARGED` to `REFUTED`. A nonsemantic
intervention changes only a comment; the raw source digest changes while the
semantic/graph digests, mutation outcome, authority, and claim state are
preserved.

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
transition order or digests change, if repository before/after differs, or if
the independent consumer cannot reconstruct the raw source.

This does not claim general Go reflection equivalence, hostile-process safety,
compiler/source completeness beyond the declared claims, or runtime
memory/performance bounds.
