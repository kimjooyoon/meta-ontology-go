# Evidence freshness experiment

`main.gooo` is a real source subject and also declares the freshness policy in
the language: six axes, earliest-change comparison, `OPEN` prior claim state,
logical epoch/environment boundaries, raw-material handling, comment-insensitive
semantic comparison, claim-ledger behavior, and CI before/after write-set
observation.

The canonical compiler parses and lowers this source before producing a policy
and semantic digest. The independent decider receives the raw source, receipt,
and current context and repeats that reconstruction without importing the
producer package.

The fixed observation separates one `CURRENT_EVIDENCE` case from nine
`SYNTHETIC_COUNTEREXAMPLE` cases:

- presentation-only comment
- semantic value change
- subject, recipe, environment, runner, and verifier changes
- expired boundary
- unavailable source

The material component of the six-axis tuple contains two distinct values:
`raw_digest` and `semantic_digest`. A comment-only intervention makes raw
freshness `STALE` while semantic freshness and the claim decision remain
preserved under `comments_ignored`. A semantic value change makes both stale.

The claim ledger starts at `OPEN`. Current evidence becomes `DISCHARGED`;
stale evidence preserves `OPEN` unless an explicit refutation is present;
unknown evidence preserves `OPEN`. Each entry is linked by a previous digest
and records source, semantic, receipt, stage, step, and reason provenance.

Independence is reported as two different facts. The receipt's
`forbidden_dependency_count` is a non-ratio guardrail and must be `0`; it does
not mean that an independence test was performed. The performed-and-passed
contract is the positive fixed-denominator value
`independence_contract = { numerator: 1, denominator: 1 }`. The validator
recomputes both facts from the receipt/report, and CI emits the same pair.
