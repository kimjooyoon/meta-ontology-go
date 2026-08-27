# Deterministic evidence quorum

This experiment justifies one bounded claim about the real
`examples/billing/main.gooo` source. It is a provenance decision, not a general
vote and not a confidence score.

The contract fixes four cases and a minimum of three independent origin groups.
The required roles are `producer`, `consumer`, and `meta-operation`. A receipt
records all four semantic coordinates: producer, consumer, meta-operation, and
proof choice. The evaluator binds every observation to the source path and
digest, groups observations by `origin_group`, and counts each group once.

The fixed cases are:

1. Three supporting groups discharge the claim.
2. Two replicas from one producer plus one consumer contain three raw
   observations, but only two independent groups; the claim remains `OPEN`.
3. Three supporting groups and one independently contradictory group refute the
   claim, even when the supporting confidence values are higher.
4. Two supporting groups are insufficient and lower the resolution.

Confidence is preserved as descriptive receipt data and is never averaged,
weighted, or used as a tie-breaker. A same-origin replica cannot increase the
quorum. Conflicts fail closed; missing quorum lowers resolution. Each decision
records an `OPEN` to `DISCHARGED`, `OPEN` to `OPEN`, or `OPEN` to `REFUTED`
claim transition with `stage`, `step`, and `reason`.

## Research basis

Adopted from the primary sources:

- Castro and Liskov's PBFT certificate requires messages from different
  replicas; its `f+1` and `2f+1` certificate sizes motivate a minimum and a
  distinct-origin rule. This experiment adopts distinct origin groups and
  fail-closed conflict handling, but does not claim Byzantine consensus.
- The W3C PROV Constraints Recommendation treats provenance validation as
  consistency checking over a recorded history and defines constraints that
  make reasoning safe. This experiment adopts explicit provenance links,
  deterministic validation, and a bounded claim graph.
- W3C PROV-O's Entity/Activity/Agent vocabulary motivates keeping the source,
  producing action, and responsible roles explicit in every receipt.

Rejected for this meta-semantic experiment:

- ordinary vote totals, confidence averages, and trust weights, because they
  can turn duplicated or correlated evidence into false support;
- the PBFT `3f+1` replica formula as a direct rule, because these are
  provenance groups rather than network replicas and no liveness claim is made;
- signatures, quorum certificates, or full compiler correctness, because the
  scope is read-only claim justification with a deterministic source digest.

Sources: [PBFT certificate algorithm](https://www.usenix.org/legacy/events/osdi2000/castro/castro_html/node4.html),
[W3C PROV Constraints](https://www.w3.org/TR/2013/REC-prov-constraints-20130430/),
and [W3C PROV-O](https://www.w3.org/TR/prov-o/).
