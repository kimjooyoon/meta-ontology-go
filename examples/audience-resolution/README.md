# Audience-resolution projection

This is a small meta-ontology experiment, not a dashboard clone. One raw
evidence ledger is projected into three nested knowledge surfaces:

| audience | visible coordinates | required coordinates | base local decision |
| --- | ---: | ---: | --- |
| USER | 4 | 12 | UNKNOWN |
| TOOL_AUTHOR | 8 | 12 | UNKNOWN |
| GOVERNOR | 12 | 12 | PASS |

The global subject decision is `PASS` for the complete ledger. A view carries
that global value only as `global_decision`, with
`inherited_status=INHERITED_NOT_LOCALLY_VERIFIED` when its visible evidence is
insufficient. It never copies the global decision into `local_decision`.
Every omitted coordinate reports its exact stage, step, and reason.

## Semantic authority

The policy is represented as stable Gooo identities in
[`main.gooo`](./main.gooo): `policy/{audience}/{ordinal}/{coordinate}` and
`resolution/{audience}/{resolution}`. The same source declares `OPEN`,
`DISCHARGED`, and `REFUTED` claim-state values plus the
`evidence-to-claim` relation. The producer and consumer both derive these
values through `syntax.ParseFile` → `bidir.Lower` → canonical semantic IR.
The source denominator is the number of canonical IR nodes, currently 56; it
is never a line-count constant.

The raw [`ledger.json`](./ledger.json) contains observations only: source
binding, evidence provenance, prior claim state, observed value, and
counterexample mutation descriptions. It does not contain a decision,
satisfaction bit, claim-after state, expected/observed decision, or
`blocked:true`. Final decisions and claim transitions are derived into the
receipt.

Claim transitions are append-only. Each audience retains the `OPEN` claim
when the evidence is omitted, moves it to `DISCHARGED` only when its visible
evidence is sufficient, and moves it to `REFUTED` for visible contradiction.
The transition records the audience visibility, evidence digest, producer,
consumer, stage, step, and reason; the claim is never deleted.

## Independent consumer

[`cmd/audience-resolution-consumer`](../../cmd/audience-resolution-consumer)
is a separate package and binary. It does not import the producer,
`Evaluate`, `ValidateReceipt`, or `CanonicalContract`. It starts from raw
`.gooo` and raw ledger bytes, reconstructs the canonical IR, recomputes the
audience projections, checks the receipt digest, and audits its own producer
imports. The CI artifact reports the exact producer-import numerator and
denominator, raw-final-field absence, source reconstruction, and all local
views.

## Counterexamples and interventions

The Action job creates real mutated raw ledgers and records their receipts:

- omission removes `receipt.seal`; the global and GOVERNOR local decisions
  become `UNKNOWN`, and the affected `OPEN` claim remains open;
- contradiction changes `ledger.coverage` to `CONTRADICTORY`; the global,
  TOOL_AUTHOR, and GOVERNOR decisions become `REFUTED`, while USER can only
  report what its visible surface supports;
- a semantic audience-policy extension changes the canonical IR digest and
  projection sizes; a comment-only source change preserves both;
- all three mutations are re-run by the independent consumer.

The falsifiable claim is: for every raw ledger variant, no audience may emit
`local_decision=PASS` unless all 12 required raw coordinates are visible and
locally sufficient; a visible contradiction must produce `REFUTED`; and a
semantic policy edit must change the semantic digest/projection while a
comment-only edit must not. These are executable assertions in
`.github/workflows/audience-resolution.yml` and its uploaded artifact.

## Information-flow/view principles

Two official references informed the design:

1. [NIST SP 800-53 Rev. 4 AC-4, Information Flow Enforcement](https://csrc.nist.gov/pubs/sp/800/53/r4/upd2/final).
   Adopted: flows are explicit, policy-bound transfers between subjects and
   objects. Rejected: a single presentation/dashboard is not a sufficient
   flow policy, and this experiment does not claim confidentiality or covert
   channel resistance.
2. [Sabelfeld and Myers, *Language-Based Information-Flow Security*](https://www.cs.cornell.edu/andru/papers/jsac/sm-jsac03.pdf).
   Adopted: a view is a security/knowledge projection whose observations must
   be checked against the information available at that view. Rejected: this
   fixture does not claim a complete noninterference proof; it tests only
   explicit coordinate omission and contradiction witnesses.

Run the witness and independent consumer in Actions; local tests are
intentionally not part of this experiment.
