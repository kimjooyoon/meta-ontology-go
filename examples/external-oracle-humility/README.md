# External oracle humility

This is a meta-ontology experiment, not a port of gomacro, Racket, or a build
system. It tests a narrower boundary: external implementations, documents, and
papers may be comparison evidence, but they cannot become the final authority
for Gooo meaning.

The authority policy is written in the real [main.gooo](./main.gooo) source as
`computes` values. The source declares `GOOO_SOURCE_INTENT`, permits only the
`COMPARATIVE_EVIDENCE` relation, and records three persistent claims:
`SOURCE_ONLY`, `EVIDENCE_ONLY`, and `REPLAYABLE_EVIDENCE`. Both the producer and
the consumer independently run `syntax.ParseFile -> bidir.Lower`; the consumer
has its own wire model and does not import the producer package.

`references.json` is a checked-in `HISTORICAL_FIXTURE` capsule. Each entry has a
URL, revision, locator, content digest, capture time, evidence class, and a
structured proposition. It has no `available` or `agreement` authority inputs.
The capsule's 3/3 result is only a metadata-conformance claim. Because the raw
bytes and a versioned extraction recipe with a recipe digest are not attached,
each reference remains `HISTORICAL_FIXTURE / UNVERIFIED / LOWER_RESOLUTION` and
the semantic agreement claim remains `OPEN`.

The pinned gomacro raw URL is classified as `IMMUTABLE_RAW`; the Racket and
Reproducible Builds documentation URLs are `MUTABLE_DOCUMENTATION`. Actions
separately retrieves all three URLs and emits `CURRENT_EVIDENCE` only for a
real 200 response whose digest matches the capsule. That closes only the
current byte observation. Retrieval failure or digest mismatch is
`OPEN/LOWER_RESOLUTION` at `retrieve`; a missing recipe is
`OPEN/LOWER_RESOLUTION` at `extract`. Current observations never grant semantic
authority.

The fixed denominator exposes these values without hiding the denominator:

| Metric | Fixed denominator |
| --- | ---: |
| source_policy | 1/1 |
| producer_imports | 0/0 |
| historical_fixtures | 3/3 |
| current_byte_observations | x/3, derived by Actions |
| semantic_extraction | 0/3 (no raw bytes + recipe receipt) |
| semantic_agreement | 0/3 |
| semantic_causality | 1/1 |
| nonsemantic_preservation | 1/1 |

The 12-indicator report also binds receipt replay, claim lifecycle evidence,
authority refusal, and the read-only effects snapshot. The subject decision is
`REFERENCE_AGREEMENT_OPEN`, never `PASS`, and always has
`authority_grant=NONE`, `enforcement_effect=NO_EFFECT`, and
`resolution=LOWER_RESOLUTION`. The three conformance cases are checked-in
inputs: an unverified capsule, a proposition/provenance mismatch branch, and a
missing reference branch. Their branch outcomes do not discharge or refute a
subject external claim.

## Adopted and rejected rules

| Primary material | Adopted comparison rule | Rejected authority leap |
| --- | --- | --- |
| [gomacro README at a pinned commit](https://github.com/cosmos72/gomacro/blob/cf0d4bf32da393dbda97e3572f216731013ffa55/README.md) | An external implementation can reveal a useful effect/capability surface for comparison. | Its macro features or arbitrary I/O become a Gooo feature or permission. |
| [Racket Reference: syntax model](https://docs.racket-lang.org/reference/syntax-model.html) | Binding, phase, and expansion descriptions can be compared with a language boundary. | Racket binding or expansion semantics define Gooo semantics. |
| [Reproducible Builds definition](https://reproducible-builds.org/docs/definition/) | Bounded inputs, content hashes, and independent replay make evidence more reproducible. | Bit equality proves semantic correctness or grants source authority. |

The experiment is falsifiable. A changed source policy value must change the
receipt, decision, and claim transition. A comment-only source change must keep
the semantic digest, decision, and claim transitions. A historical proposition
or provenance change must reproduce a conformance mismatch branch without
turning the subject comparison claim into `REFUTED`; an absent reference must
reproduce the absence branch with `LOWER_RESOLUTION`. A changed or unavailable
Actions retrieval must become `OPEN/LOWER_RESOLUTION`, while a successful
retrieval without an extraction recipe must remain semantically `OPEN`. Any
official write or promotion fails the read-only guard.
