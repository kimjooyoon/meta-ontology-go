# Hygienic origin identity

This is a closed philosophical experiment in Gooo about generated-name
identity. It does not add macro syntax or claim to implement a Racket or Rust
expander. The Gooo file declares the ontology and two fixed observations; the
independent judge reads only those source annotations and derives the receipt
from origin identity, scope provenance, and binding resolution.

## Proposition

Generated spelling is not identity. The generated name `tmp` is safe only when
the receipt preserves both the producer origin identity and the scope
provenance that keeps it out of the consumer's `tmp` binding.

| case | spelling | origin identity | scope provenance | resolved identity | result |
| --- | --- | --- | --- | --- | --- |
| captured | `tmp` | consumer binding | consumer call site | consumer binding | `REFUTED` for both preservation claims |
| hygienic | `tmp` | producer expansion 1 | fresh producer expansion 1 | producer expansion 1 | `DISCHARGED` for both preservation claims |

The captured row is an intentional counterexample, not a failing test: it
shows that text equality can coexist with identity failure. The hygienic row
is the safe non-capture contrast.

## Evidence contract

- producer: `gooo://hygienic-origin-identity/producer/name-generator`
- consumer: `gooo://hygienic-origin-identity/consumer/binding-site`
- meta-operation: `generate-name-preserving-origin-and-scope`
- proof choice: `ORIGIN_SCOPE_EQUIVALENCE`
- decision algebra: `PASS` when the two known rows classify exactly; `UNKNOWN`
  when scope evidence is unavailable
- claim statuses are retained as `OPEN`, `DISCHARGED`, or `REFUTED`; an
  unavailable path is separately retained with `stage`, `step`, and `reason`
- the receipt is read-only evidence (`repository_writes=0`, mutation authority
  `false`)

The fixed denominator is four preservation claims over two cases. The baseline
receipt therefore reports 2 discharged, 2 refuted, 0 open, and 4/4 classified
(10,000 basis points of classification coverage). Preservation satisfaction is
2/4 (5,000 basis points), making the counterexample visible instead of hiding
it in a pass rate. The unknown guardrail keeps the same fixed denominator,
adds one explicit unknown path, and carries one separate `OPEN` guardrail
claim.

## Research and design decisions

The experiment adopts the following ideas from official documentation:

1. [Racket's syntax model](https://docs.racket-lang.org/reference/syntax-model.html)
   treats bindings as a function of identifier spelling plus scope sets, and
   [Racket's syntax-object guide](https://docs.racket-lang.org/guide/stx-obj.html)
   keeps lexical context and source location attached to syntax objects. We
   adopt the separation between surface spelling and lexical provenance.
2. [Rust's macro-by-example hygiene
   rules](https://doc.rust-lang.org/stable/reference/macros-by-example.html#hygiene)
   distinguish definition-site and invocation-site resolution, while the
   [procedural-macro reference](https://doc.rust-lang.org/nightly/reference/procedural-macros.html#procedural-macro-hygiene)
   explicitly warns that procedural macro output is unhygienic. We adopt the
   resolution-site distinction and reject any proof that assumes generated
   text is hygienic merely because it is fresh-looking.

We reject copying either language's macro grammar into Gooo. The experimental
unit is the generated artifact's identity tuple, not a new expansion feature.
We also reject treating missing provenance as `REFUTED`: `unknown.gooo` must
produce `UNKNOWN` with a non-empty stage/step/reason tuple.

## Falsification

The claim is falsifiable. Change the hygienic row's `resolves` value to
`consumer-binding`, change its scope to `consumer-call-site`, or change either
row's spelling so the two spellings differ. The independent judge must reject
the expected receipt. Removing the unknown row's stage, step, or reason must
also fail validation. This experiment does not prove real macro expansion,
cross-module imports, or runtime behavior; those are explicit follow-up
questions rather than inferred conclusions.

CI runs the Gooo syntax check, the independent judge, two byte-identical
receipts for each source, and exact fixed-denominator assertions on Go 1.27.
