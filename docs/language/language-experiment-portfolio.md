# Language experiment portfolio governance

The portfolio is a comparison protocol, not a dashboard. It exists to keep a
promising result from becoming a language-quality claim merely because one
experiment produced a good-looking output.

## Protocol

Each alternative is a real `.gooo` source declaration with its own named
meta-operation. A receipt producer records the source digest, candidate
identity, provenance, fixed coordinate vector, counterexample records, unknown
locations, extension evidence, and read-only effects. A separate evaluator
verifies the receipt digest and emits a report that copies the receipt's vector
and evidence lists. It does not replace them with a total.

The contract fixes the coordinate order and denominator:

| Coordinate | Denominator | What is preserved |
| --- | ---: | --- |
| `source-replay` | 1 | source binding evidence |
| `receipt-independence` | 1 | receipt boundary evidence |
| `counterexample-boundary` | 2 | raw counterexample count |
| `unknown-localization` | 2 | raw unknown locations |
| `extension-evidence` | 1 | extension evidence state |
| `read-only-effects` | 1 | repository effects |

The numerator is an observed integer from the receipt. `OPEN`, `DISCHARGED`, and
`REFUTED` are kept alongside it. The protocol has no aggregate score, rank,
winner, estimated improvement, or weighted average. A refuted coordinate is
useful evidence about a boundary; it is not a penalty that can be hidden by a
different coordinate.

## External principles applied

1. [W3C Testing Policy](https://www.w3.org/policies/testing/) says that a
   normative specification change should have a corresponding test change, or
   an explicit rationale when testing is not practical. The portfolio therefore
   keeps the three source alternatives visible and has CI check every one. A
   missing or untestable case is recorded as evidence rather than silently
   removed.

2. [The TC39 Process](https://tc39.es/process-document/) separates ideation,
   design, testing/validation, and implementation experience. It also expects
   unresolved concerns and non-advancement reasons to be recorded. The
   portfolio mirrors that discipline with `stage`, `step`, `reason`, and the
   three explicit evidence states. An `OPEN` coordinate is not promoted to a
   discharged claim.

3. [ACM CCS Artifact Evaluation](https://sigsac.hosting.acm.org/ccs/CCS2025/call-for-artifacts/)
   evaluates whether artifacts are documented, consistent, complete, and
   exercisable, and treats independent repetition of results as a separate
   question. The receipt/evaluator split, content digests, replay check, and
   explicit non-claims apply this separation locally. The repository is not
   claiming an ACM badge; this is a design translation of the principle.

## Limits and falsification

- The fixture contains only three alternatives and six fixed coordinates. It
  cannot establish open-world completeness or general language behavior.
- The receipt producer records declared evidence; it does not execute the
  meta-operations or prove their semantic equivalence.
- A source or receipt digest proves identity of bytes, not truth of the claim
  described by those bytes.
- Counterexamples and unknown locations are preserved only when their producer
  supplies them. A future producer can be falsified by a missing or mismatched
  count, stale digest, missing provenance field, or coordinate reorder.
- Extension evidence is intentionally local. It demonstrates that an evidence
  slot can remain `OPEN` or be discharged; it does not prove arbitrary future
  extensibility.
- The read-only guardrail is a receipt assertion checked by CI. It is not a
  sandbox or a proof against every possible side effect.

The intended falsifier is concrete: alter one coordinate, reorder the vector,
change a status without its numerator, remove an unknown location, forge a
digest, or add a repository effect. The evaluator must then return
`FAIL_CLOSED`, and the CI counterfactual must observe that result. This makes
the comparison protocol itself testable without turning its evidence into a
winner declaration.
