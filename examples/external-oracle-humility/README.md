# External oracle humility

This is a meta-ontology experiment, not a port of gomacro, Racket, or a build
system. It asks a narrower question: how can an external implementation,
language reference, or paper be accepted as comparison evidence without being
accepted as the final authority for Gooo meaning?

The real source under observation is `main.gooo`. Its semantic authority stays
`GOOO_SOURCE_INTENT`. The reference set records three primary sources:

- [gomacro README at commit `cf0d4bf`](https://github.com/cosmos72/gomacro/blob/cf0d4bf32da393dbda97e3572f216731013ffa55/README.md), used to compare an external implementation's effect surface;
- the [Racket Reference syntax model](https://docs.racket-lang.org/reference/syntax-model.html), used to compare a language's binding and expansion boundary;
- the [Reproducible Builds definition](https://reproducible-builds.org/docs/definition/), used to compare replayable, bit-exact evidence.

The adopted principle is `compare -> record -> independently replay`. The
rejected principle is `external agreement -> semantic promotion`. Thus the
positive case is deliberately named `REFERENCE_AGREEMENT_OBSERVED`, never
`PASS`, and has `authority_grant=NONE`, `official_mutations=0`, and
`enforcement_effect=NO_EFFECT`.

The fixed denominator has 12 indicators. It binds the source receipt, three
references, their comparative-only relation, the refusal to grant external
authority, persistent claim transitions, and read-only/no-PASS guardrails.
The case denominator has exactly 3 cases: agreement, known mismatch, and
reference absence. Agreement is an exact observation; mismatch is an exact
known failure; absence is `UNKNOWN` and fails closed. None can rewrite the
source, official semantic denominator, or promotion state.

The receipt has explicit `producer`, `consumer`, `meta_operation`,
`proof_choice`, `stage`, `step`, and `reason` fields. The consumer is an
independent bounded judge: it re-reads the `.gooo` source and receipt and does
not import the compiler parser. Claim transitions persist the distinction:
`OPEN -> AGREEMENT_OBSERVED`, `SOURCE_ONLY -> SOURCE_ONLY`, and
`UNCHANGED -> UNCHANGED` for the official state.

## Adopted and rejected rules

| Primary material | Adopted comparison rule | Rejected authority leap |
| --- | --- | --- |
| gomacro README | External implementations expose useful effect and capability comparisons. | A macro system's feature set or arbitrary I/O becomes a Gooo feature or permission. |
| Racket Reference | Binding, phase, and expansion descriptions are evidence about a possible language boundary. | Racket's binding/expansion semantics define Gooo semantics. |
| Reproducible Builds definition | Same bounded inputs and independent replay support exact evidence receipts. | Bit equality proves semantic correctness or grants source authority. |

This boundary is falsifiable. A change to the source digest, claim IDs, source
declarations, reference availability, reference relation, authority field,
receipt replay, transition state, or denominator changes the corresponding
indicator or lowers the resolution. A future implementation may disprove the
model only by showing that an external reference can change Gooo semantic
intent without a Gooo source change; this experiment intentionally rejects
that promotion path.
