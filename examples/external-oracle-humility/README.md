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
URL, revision, locator, content digest, capture time, and a structured
proposition. It has no `available` or `agreement` authority inputs. The judge
derives agreement by comparing those propositions with the policy predicate.
Actions separately retrieves all three URLs and emits `CURRENT_EVIDENCE` only
for a 200 response whose digest matches the capsule. Missing or changed content
stays `OPEN` with `LOWER_RESOLUTION`; current observations never grant semantic
authority.

The fixed denominator exposes these values without hiding the denominator:

| Metric | Fixed denominator |
| --- | ---: |
| source_policy | 1/1 |
| producer_imports | 0/0 |
| current_reference_observations | x/3, derived by Actions |
| historical_fixtures | 3/3 |
| semantic_causality | 1/1 |
| nonsemantic_preservation | 1/1 |

The 12-indicator report also binds receipt replay, claim lifecycle evidence,
authority refusal, and the read-only effects snapshot. The positive decision is
`REFERENCE_AGREEMENT_OBSERVED`, never `PASS`, and always has
`authority_grant=NONE` and `enforcement_effect=NO_EFFECT`. The three conformance
cases are checked-in inputs: agreement, a proposition mismatch, and a missing
reference. Subject resolution and conformance are reported separately.

## Adopted and rejected rules

| Primary material | Adopted comparison rule | Rejected authority leap |
| --- | --- | --- |
| [gomacro README at a pinned commit](https://github.com/cosmos72/gomacro/blob/cf0d4bf32da393dbda97e3572f216731013ffa55/README.md) | An external implementation can reveal a useful effect/capability surface for comparison. | Its macro features or arbitrary I/O become a Gooo feature or permission. |
| [Racket Reference: syntax model](https://docs.racket-lang.org/reference/syntax-model.html) | Binding, phase, and expansion descriptions can be compared with a language boundary. | Racket binding or expansion semantics define Gooo semantics. |
| [Reproducible Builds definition](https://reproducible-builds.org/docs/definition/) | Bounded inputs, content hashes, and independent replay make evidence more reproducible. | Bit equality proves semantic correctness or grants source authority. |

The experiment is falsifiable. A changed source policy value must change the
receipt, decision, and claim transition. A comment-only source change must keep
the semantic digest and decision. A changed or absent capsule proposition must
become `DISAGREES` or `UNKNOWN`; a changed or unavailable Actions retrieval must
become `OPEN/LOWER_RESOLUTION`. Any official write or promotion fails the
read-only guard.
