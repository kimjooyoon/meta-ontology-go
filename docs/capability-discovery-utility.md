# Capability Discovery Utility Contract

## Purpose

Capability discovery answers a bounded question: what can this `.gooo` declaration support now, what evidence is missing, and what is the next non-executing operation? It is not a completeness score, an execution request, an authorization decision, or a claim that the language can represent every domain.

The discovery result is useful only when a user can act on it without guessing which boundary was crossed.

## Observable contract

Every discovery receipt binds the same source tuple:

- source digest
- declaration or query digest
- capability catalog digest
- schema digest
- evidence digest
- toolchain identity

The receipt must contain exactly one state:

- `AVAILABLE`: a capability is supported by the bound declaration and evidence.
- `DEFERRED`: the capability is plausible, but a named declaration or evidence stage is missing.
- `UNKNOWN`: the source, evidence, or boundary is incomplete, invalid, or tampered.

Each receipt also contains a `first_missing_stage` when the state is not `AVAILABLE`, and a `next_operation`. The next operation is descriptive only: it cannot execute code, mutate a catalog, authorize access, or silently widen scope.

## Metrics

Metrics describe utility of the discovery path, not semantic completeness of the language.

### Query resolution

`resolved_query_rate = available_or_deferred_queries / valid_query_receipts`

The denominator is every valid receipt in the declared measurement window. Invalid or tampered receipts are excluded from this rate and counted separately; they must never improve the rate by disappearing.

### Boundary fidelity

`boundary_fidelity = receipts_preserving_non_execution_and_non_authorization / valid_query_receipts`

This is a hard guardrail. Any execution or authorization crossing is a contract failure, not a low score to average away.

### Evidence continuity

`evidence_continuity = receipts_with_complete_source_tuple / valid_query_receipts`

A cached result can be reused only when the exact source tuple is unchanged. Cache presence is not semantic evidence and does not increase this metric.

### Next-operation utility

`next_operation_utility = confirmed_useful_outcomes / adopted_next_operations`

A useful outcome must be confirmed by a later observation that carries the prior receipt digest. Self-reported adoption, generated text, or a higher capability count is not confirmation.

### Boundary localization

`first_missing_stage_accuracy` is evaluated only on fixtures with an independently known missing stage. It measures whether discovery identifies the first missing boundary rather than a later symptom. Unknown ground truth remains unknown and is not scored as a failure or success.

## Investment policy

Investment follows the failing contract stage:

- Repeated `UNKNOWN` at source or tuple validation: repair provenance and input integrity before adding capabilities.
- Repeated `DEFERRED` at the same declaration stage: add the smallest declaration or evidence primitive that can move that stage forward.
- `AVAILABLE` with low next-operation utility: improve explanations, examples, and LSP presentation before expanding the catalog.
- Boundary fidelity below 1.0: stop capability expansion and repair the non-execution/non-authorization guard first.
- No confirmed outcomes in the measurement window: keep the investment decision `HOLD_FOR_EVIDENCE`.

Every investment receipt records the observed failure stage, bounded change, expected falsifier, rollback trigger, and exact source/toolchain tuple. It must not convert a fixture conformance result into a whole-language claim.

## Measurement window

A window is valid only when it has:

1. a declared query population and inclusion/exclusion rule;
2. exact source, schema, catalog, evidence, and toolchain identities;
3. counts for `AVAILABLE`, `DEFERRED`, `UNKNOWN`, and invalid receipts;
4. the first missing stage for each non-available receipt;
5. at least one confirmed or refuted outcome where utility is claimed.

If any item is missing, the window is `UNKNOWN` and its first missing stage is preserved. The next safe operation is to repair the window, not to infer improvement.

## Falsifiers and rollback

This contract is falsified when a discovery receipt can be marked `AVAILABLE` without the required source tuple, when an `UNKNOWN` receipt loses its first missing stage, or when the next-operation utility metric rises without a linked later observation. Roll back the capability change and retain the failed receipt when any falsifier is observed.
