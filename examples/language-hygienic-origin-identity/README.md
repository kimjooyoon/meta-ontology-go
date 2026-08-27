# Hygienic origin identity

This is a closed philosophical experiment in Gooo. It tests the proposition
that generated-name identity is the tuple `(spelling, origin identity, scope
provenance, resolved identity)`, not spelling alone. It does not copy macro
grammar or claim to implement a Racket or Rust expander.

## Semantic source, not comment evidence

`main.gooo`, `unknown.gooo`, `comment-only.gooo`, and `intervention.gooo` are
parsed by `syntax.ParseFile` and lowered by `bidir.Lower`. The consumer judge
reads only activity `computes` values from the lowered semantic IR:

- producer activities emit `case`, `spelling`, `origin`, and `scope`;
- consumer activities invoke a resolver that derives `resolved identity` or
  reports `provenance=missing`;
- the consumer package has its own wire model and judgment and imports no
  producer or evaluator package.

The `# experiment...` lines are intentionally misleading, non-authoritative
comments. A comment-only edit changes `raw_digest` but preserves the semantic
digest, decision, claims, transitions, and result. The old design was
REFUTED because comments could change PASS to UNKNOWN; this version makes
that impossible by construction and checks it in CI.

## Cases and separated metrics

| case | spelling | producer origin | producer scope | resolver result | role |
| --- | --- | --- | --- | --- | --- |
| captured | `tmp` | consumer binding | consumer call site | consumer binding | negative control |
| hygienic | `tmp` | producer expansion 1 | fresh producer expansion 1 | producer expansion 1 | target |

The captured case intentionally produces two `REFUTED` control claims. They
are not mixed into target preservation. The hygienic target has two separate
preservation claims, both `DISCHARGED` in the baseline.

The fixed CI denominators are:

| metric | observed / expected |
| --- | ---: |
| source cases | 2 / 2 |
| producer imports | 0 / 0 |
| semantic causality | 1 / 1 |
| comment invariance | 1 / 1 |
| control capture | 1 / 1 |
| hygienic non-capture | 1 / 1 |
| target preservation claims | 2 / 2 |

Baseline exact values are 2 discharged and 2 refuted total claims, with
4/4 classified (`10000` bps). Target preservation is 2/2 (`10000` bps).
This is a PASS because the target succeeds, while the captured negative
control remains visibly REFUTED.

## UNKNOWN and transitions

`unknown.gooo` encodes missing provenance in the semantic resolver value, not
in a comment. It yields `UNKNOWN` / `LOWER_RESOLUTION`, keeps the fixed 2-case
and 4-claim denominators, and records one unknown path with non-empty
`stage`, `step`, `reason`, `evidence_digest`, and `provenance`. The hygienic
scope claim and the explicit unknown guardrail claim remain `OPEN`.

Every claim also has an append-only transition from `UNCLASSIFIED` to its
final `OPEN`, `DISCHARGED`, or `REFUTED` status. The receipt retains the
transition evidence and provenance instead of only reporting final status.

## Intervention and reconstruction

`intervention.gooo` changes only the semantic resolver value for the hygienic
case from producer expansion to consumer binding. CI requires:

- semantic digest changes: 1/1 causal intervention;
- decision changes `PASS` -> `REFUTED`;
- target preservation changes from 2/2 (`10000` bps) to 1/2 (`5000` bps);
- receipt digest and claim transitions change;
- a fresh source reconstruction agrees with the intervention report.

The validator reconstructs the report from source before accepting a receipt,
so a coherent reseal of tampered cases is rejected. CI also snapshots
`git status --porcelain=v1` before and after evaluation and requires the
observed repository write delta to be zero.

## Research choices

1. [Racket's syntax model](https://docs.racket-lang.org/reference/syntax-model.html)
   makes binding depend on identifier spelling plus scope sets, and its
   [syntax-object guide](https://docs.racket-lang.org/guide/stx-obj.html)
   keeps lexical context and source location attached. We adopt the separation
   of spelling from lexical provenance and reject raw spelling as identity.
2. [Rust macro-by-example hygiene](https://doc.rust-lang.org/stable/reference/macros-by-example.html#hygiene)
   distinguishes definition-site and invocation-site resolution. We adopt
   that resolution-site distinction. Rust's
   [procedural-macro reference](https://doc.rust-lang.org/nightly/reference/procedural-macros.html#procedural-macro-hygiene)
   describes output as unhygienic; we reject any proof that assumes generated
   text is safe merely because it looks fresh.

## Falsification

The claim is falsifiable. Change the hygienic producer's origin or scope,
change the resolver binding to the consumer binding, or remove the semantic
provenance value. The independent consumer must change the receipt and
decision or preserve UNKNOWN. Change only a comment: raw digest must change,
while semantic digest, decision, claims, and transitions must remain equal.
Remove UNKNOWN evidence coordinates or provenance and CI must reject the
receipt. The experiment does not prove runtime macro expansion, cross-module
imports, or language-level hygiene.
