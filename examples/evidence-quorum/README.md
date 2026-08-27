# Deterministic evidence quorum

This is a small Gooo meta-semantics experiment. It asks whether the claim

> the bounded evidence activity is justified by independent evidence

can be discharged without averaging confidence scores. The answer is a
deterministic quorum over structured provenance, required evidence predicates,
roles, conflict policy, and a fixed case denominator.

The actual Gooo subject is [`main.gooo`](main.gooo). The semantic policy is
authored in [`policy.gooo`](policy.gooo), including `threshold=3`, the three
required predicates, `prior_claim_state=OPEN`, six cases, and the valid
contradiction predicate. Canonical parsing and lowering produce the policy
semantic digest; Go supplies only wire schemas and invariants.

## Closed three-channel example

The dedicated CI workflow executes three different channels:

1. `cmd/gooo run --json --entry ProduceEvidence` produces the actual
   `gooo/source-execution-receipt/v1` receipt. A source-channel wrapper keeps
   that receipt intact and adds its executable digest, sorted dependency list
   and digest, subject raw/semantic digests, and observation digest.
2. `cmd/evidence-quorum-reconstructor` parses and lowers raw `main.gooo` on a
   separate executable path without importing the producer or canonical
   contract.
3. `cmd/evidence-quorum-artifact-observer` observes the independently generated
   Go artifact and its manifest.

The consumer at `internal/meta/evidencequorumconsumer` receives only raw
channel receipts and raw `.gooo` bytes. It computes the provenance lineage as
`executable_digest + dependency_digest`; identical lineages collapse into one
origin group even when a receipt claims a different label. The baseline has
exactly **3 current receipts / 3 distinct provenance groups / 0 current
replicas**, so the fixed quorum is **3/3**.

Four resealed cases are explicitly marked `SYNTHETIC_COUNTEREXAMPLE`: one
duplicate, one valid contradiction, one invalid contradiction, and one
unknown. They are not current evidence. Across the six case inputs CI sees
**21 raw receipt appearances, 3 unique current receipts, 4 unique synthetic
receipts, 3 provenance groups, and 4 collapsed replicas**. The valid
contradiction alone yields `REFUTED`; an invalid contradiction yields
`OPEN + FAIL_CLOSED`; unknown yields `OPEN` with `UNKNOWN/LOWER_RESOLUTION`.

Every claim record is append-only from `OPEN` and includes a previous state
digest, evidence digests, structured provenance, stage, step, and reason.
`conformance_decision` (whether the case behaves according to the experiment)
is separate from `subject_decision` (what the claim itself says).

## Interventions and limits

CI emits `interventions.json` next to the report:

- changing the policy threshold from 3 to 4 changes the quorum result and the
  policy semantic digest;
- adding only comments to the policy preserves the semantic digest, observation
  digest, and quorum result;
- both before/after observations remain read-only (`repository_writes=0`,
  `mutation_authority=false`).

This does not claim full Byzantine consensus, trust in an identity, compiler
semantic correctness, or confidence-weighted probability. The adopted
principles are the provenance vocabulary of [W3C PROV-O](https://www.w3.org/TR/prov-o/)
and its consistency constraints
([PROV-Constraints](https://www.w3.org/TR/2013/REC-prov-constraints-20130430/));
the rejected principle is treating repeated/correlated observations as new
votes. The Byzantine reference is the replica-distinct certificate idea in
[Castro and Liskov's PBFT description](https://www.usenix.org/legacy/events/osdi2000/castro/castro_html/node4.html),
which informs independence but is not claimed as a full consensus protocol.

Run `scripts/evidence-quorum/main.sh` only in the dedicated GitHub Actions
workflow. Local tests are intentionally not part of this experiment.
