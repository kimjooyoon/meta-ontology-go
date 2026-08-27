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

### v2 source-semantic causality

The v2 contract preserves the six v1 coordinate denominators exactly and adds
one seventh coordinate, `source-semantic-causality`, with denominator `3`.
`contract.json` records `contract-v1.json`, its predecessor digest, and the
additive upgrade reason. This explicit predecessor binding prevents a new
axis from silently changing the meaning of the original denominator.

Each candidate must provide the same fixed three-case contract: baseline,
semantic intervention, and non-semantic intervention. The source observer
reads the actual `computes` value from the `.gooo` source. A semantic
intervention must change that value and change at least one contracted receipt
field (`semantic_value`, `decision`, or `claim_transitions`). A non-semantic
intervention may change only whitespace/comments: its raw source digest must
change while the receipt semantic projection and decision stay equal.

The independent causality consumer reports exact case status, stage, step,
reason, coordinate vectors, and claim transitions. It exposes
`causal_cases N/N`, `digest_only_cases`, `hardcoded_fixture_cases`, and
`UNKNOWN` findings without producing a score, aggregate, weighted average,
rank, or winner. A semantic source change with no contracted receipt change is
`REFUTED / DIGEST_ONLY_BINDING`; a comment-only semantic drift is refuted with
`NON_SEMANTIC_SEMANTIC_DRIFT`; missing evidence is fail-closed with its exact
stage/step/reason.

The fixed transition denominator is `9`: three operations multiplied by the
three intervention claims (baseline observation, semantic causality, and
non-semantic invariance). The current evidence yields `REFUTED 3/9`,
`DISCHARGED 6/9`, and `OPEN 0/9`. Baseline source observation claims transition
from `OPEN` to `DISCHARGED`; semantic causality claims transition from `OPEN`
to `REFUTED`; and non-semantic invariance claims transition from `OPEN` to
`DISCHARGED`. A `0/3` coordinate is a count and does not override its derived
`REFUTED` status.

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

- The fixture contains only three alternatives, seven v2 coordinates, and
  three intervention cases per alternative. It cannot establish open-world
  completeness or general language behavior.
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
- The current receipt producer intentionally does not bind its semantic
  receipt fields to the `.gooo` operation value. CI therefore reports three
  `REFUTED / DIGEST_ONLY_BINDING` semantic cases and three
  `hardcoded_fixture_cases`; this is the intended counterexample for the
  upgraded metric, not evidence that the producer is semantically sound.

The intended falsifier is concrete: alter one coordinate, reorder the vector,
change a status without its numerator, remove an unknown location, forge a
digest, or add a repository effect. The evaluator must then return
`FAIL_CLOSED`, and the CI counterfactual must observe that result. This makes
the comparison protocol itself testable without turning its evidence into a
winner declaration.
