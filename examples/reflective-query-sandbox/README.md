# Reflective query sandbox

This is a deliberately narrow vertical slice: a Gooo program describes the
subjects it may query, the prior state of its claims, and the metric relation
being queried. The producer parses and lowers that exact source, projects the
canonical semantic IR into the existing read-only query graph, and records
what the query API actually returned.

The source IDs are semantic coordinates, not comments or Go-side fixtures:

- `/subject/query`, `/subject/structure`, `/claim-state/open`, and
  `/metric/relation/read-only` are the query subject and query targets;
- every `/claim/<class>/<name>/<proof>/<meta-operation>/<evidence>/<prior>`
  entity is a claim with a formal prior state and evidence attempt;
- activity signatures create the authoritative `used` relations that the
  producer queries. The `ReflectClaims` spelling alone is never evidence.

The source currently contains 12 claims, distributed by its own ID
coordinates into three classes and three proof choices. The producer computes
the denominator, class/proof buckets, source node/fact counts, attempt counts,
and claim-transition count from the parsed IR. No Go constant table supplies
those values.

## Boundary observations

The producer makes four exact read-only queries from source-declared
activities, one real mutation-capable request, and one unknown-target query.
The mutation request calls `semantic.Graph.ApplyGraphPatch` against the
semantic graph with an immutable `id` field. A rejection is recorded only when
that API returns its rejection; there is no `DENIED` constant shortcut. If the
API accepts the request, the attempt is `REFUTED` and its claim transitions
become `REFUTED`.

The query decision and subject resolution are separate fields. A missing
endpoint is `UNKNOWN / LOWER_RESOLUTION / UNKNOWN_TARGET`, with stage
`UNKNOWN`, step `resolve-unknown-subject`; its claim remains `OPEN` with
reason `UNKNOWN_PRESERVED`. A known exact relation is `PASS / EXACT`. A
boundary violation is `REFUTED`.

Each claim gets two append-only events: its source-declared prior state, then
an evidence-dependent event. Only the attempt named by the claim's source
`evidence` coordinate may move it to `DISCHARGED`; unknown evidence retains
`OPEN`; explicit violations move it to `REFUTED`.

The shell captures repository status before and after the producer. Those
actual lines and their changed write set are embedded in the observation and
copied into the receipt. The baseline target is `repository_writes=0` and
`mutation_authority=false`.

## Independent consumer

The consumer binary has its own wire structs and imports no producer package.
It directly performs `syntax.ParseFile -> bidir.Lower -> query.FromSemanticIR`
on the raw `.gooo`, recomputes the source and query digests, calls the same
mutation-capable API boundary, rebuilds attempts and transitions, and only
then emits a receipt. CI checks the producer import budget as `0 / 0`.

The receipt reports a dynamic indicator numerator/denominator, source
reconstruction numerator/denominator, and actual reflective-query
numerator/denominator. Because the unknown-target claim intentionally remains
open, the baseline indicator numerator is one less than the source denominator
while conformance can still be `PASS`: unknown is a lower-resolution subject
result, not a successful claim discharge.

## Interventions

CI also publishes a JSON intervention artifact. A semantic intervention
changes only the metrics activity's declared input from `MetricReadRelation`
to the already-declared `ClaimPriorOpen`; the canonical semantic and query
digests change, the metrics query changes from `PASS` to `UNKNOWN`, and the
metrics claim changes from `DISCHARGED` to `OPEN`. A nonsemantic intervention
changes only a comment; the raw source digest changes while the semantic
digest, graph digest, query decision, and claim state are preserved.

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
the mutation API accepts the immutable request without a `REFUTED` result, if
unknown becomes exact/pass, if a claim discharges without its named attempt,
if transition order/digests are changed, if repository before/after differs,
or if the independent consumer cannot reconstruct the raw source.

This does not claim general Go reflection equivalence, hostile-process safety,
compiler/source completeness beyond the declared claims, or runtime
memory/performance bounds.
