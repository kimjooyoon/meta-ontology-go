# Audience resolution projection

This is a read-only philosophical experiment over one evidence ledger. It is
not a dashboard copy: `USER`, `TOOL_AUTHOR`, and `GOVERNOR` receive different
fixed coordinate sets and fixed resolution labels from the same ledger. The
projection carries one global `decision` and `reason` into every view, so a
lower view cannot turn a governor-level contradiction into `PASS`.

## Fixed contract

The contract has 12 indicators and 12 ledger records. The audience projections
are nested and deliberately omit information:

| Audience | Resolution | Coordinates | Omitted |
| --- | --- | ---: | ---: |
| `USER` | `USER_VISIBLE_COORDINATES` | 4/4 | 8 |
| `TOOL_AUTHOR` | `TOOL_CONTRACT_COORDINATES` | 8/8 | 4 |
| `GOVERNOR` | `GOVERNOR_FULL_LEDGER` | 12/12 | 0 |

The fixed class denominator is `OUTCOME 4`, `DRIVER 5`, `GUARDRAIL 3`.
The proof-choice denominator is `FOUNDATION 6`, `COHERENCE 3`, `REGRESSION 3`.
Each indicator records its producer, consumer, meta-operation, proof choice,
stage, step, reason, and `UNPROVEN -> OBSERVED` claim transition. Failed
coordinates transition to `BLOCKED`.

The authoritative source is
[`main.gooo`](main.gooo). It contains 22 declarations. `ledger.json` binds its
SHA-256 digest and the witness command checks the source bytes before producing
the receipt.

## Counterexamples

Two cases are part of the ledger as explicit blocked witnesses:

1. `counterexample.missing-information` removes the `author.consumer`
   coordinate. The result must be `FAIL_CLOSED / LOWER_RESOLUTION`, even if
   the four USER coordinates remain locally satisfied.
2. `counterexample.decision-contradiction` changes a record decision so that
   the same source has conflicting decisions. The result must be
   `FAIL_CLOSED / INVARIANT_ONLY` for all three audiences.

The independent `ValidateReceipt` checker recomputes the receipt digest and
checks fixed cardinality, canonical coordinate sets, claim transitions,
counterexamples, and shared decision/reason. It is intentionally separate from
the evaluator and cannot make a local view authoritative.

Run the executable witness with:

```sh
go run ./cmd/audience-resolution-witness \
  --contract examples/audience-resolution/contract.json \
  --ledger examples/audience-resolution/ledger.json \
  --source examples/audience-resolution/main.gooo \
  --out /tmp/audience-resolution-receipt.json
```

## Information-flow/view research decisions

This experiment adopts two principles from official material:

- [NIST SP 800-53 Rev. 4, AC-4 Information Flow Enforcement](https://csrc.nist.gov/pubs/sp/800/53/r4/upd2/final)
  treats information flow as an explicit policy between sources and
  destinations, with enforcement at policy points. Adopted here: every
  coordinate names its producer and consumer, the allowed audience sets are
  explicit, and malformed or unauthorized flow fails closed. Rejected here:
  treating access-control success as proof of semantic correctness; this
  experiment does not authorize producers or enforce external permissions.
- [Sabelfeld and Myers, *Language-Based Information-Flow Security*](https://www.cs.cornell.edu/andru/papers/jsac/sm-jsac03.pdf)
  frames confidentiality as a property of what an observer can infer from
  outputs and uses noninterference as the formal lens. Adopted here: a view is
  a deterministic observation function of one ledger and one audience, and
  view decisions are compared for non-contradiction. Rejected here: claiming
  full noninterference or covert-channel freedom; this small experiment only
  proves explicit coordinate omission and decision consistency.

These choices make the claim falsifiable. A missing record, duplicate or
conflicting record, forged receipt digest, non-nested coordinate set, or
audience-specific decision is a concrete failing input rather than an
interpretive disagreement.
