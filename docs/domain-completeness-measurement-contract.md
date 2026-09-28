# Domain Completeness Measurement Contract v1

## Purpose

This contract measures the current evidence for a declared domain. It does not
turn a heuristic score into a claim that the domain is complete, correct, or
finished. Every metric is an observation with a denominator, an evidence
binding, a time, and an explicit unknown set.

A domain receipt MUST preserve the distinction between:

- `IN_SCOPE`: the item is part of the declared domain boundary.
- `AVAILABLE`: the item has an observed implementation path and evidence.
- `DEFERRED`: the item is known but requires a later non-executing operation.
- `UNKNOWN`: the observation cannot establish a state.
- `OUT_OF_SCOPE`: the item is intentionally excluded by the declaration.

`DEFERRED`, `UNKNOWN`, and `OUT_OF_SCOPE` MUST NOT be counted as successful
implementation evidence.

## Domain declaration

A domain declaration is an immutable, versioned set `D` of capability or
scenario identifiers. The receipt MUST include:

- `domain_id` and `domain_version`;
- ordered declared identifiers and their declaration digest;
- the source and schema/toolchain identities used to observe them;
- the policy that determines the in-scope and out-of-scope sets.

Changing the declaration, catalog, schema, or policy creates a new measurement
context. Measurements from different contexts MUST NOT be silently compared.

## Evidence tuple

Each observation MUST bind the following exact identities:

- `source_digest`;
- `declaration_digest`;
- `catalog_digest`;
- `schema_digest`;
- `evidence_digest`;
- `toolchain_identity`;
- observation timestamp and correlation id.

A cache hit, a successful job allocation, or the presence of an artifact is not
semantic or test evidence. Evidence is valid only when the referenced artifact
or observation can be verified against the exact tuple.

## Metrics

### 1. Scope coverage

For an in-scope set `S` and an observed set `O`, report:

```
scope_coverage = |{x in S: x has an explicit observation}| / |S|
```

The receipt MUST report the numerator, denominator, ordered identifiers, and
unobserved identifiers. Coverage is not implementation completeness: an
`UNKNOWN` or `DEFERRED` observation contributes to the numerator only as an
observed state, never as successful implementation evidence.

Also report:

```
available_coverage = |{x in S: state(x) = AVAILABLE}| / |S|
```

### 2. Evidence completeness

For each observation, evaluate the required evidence fields individually:

```
evidence_completeness = valid_required_fields / required_fields
```

The receipt MUST list missing or invalid fields. A high value cannot upgrade an
`UNKNOWN` or `DEFERRED` capability to `AVAILABLE`.

### 3. Provenance continuity

A provenance chain is continuous only when every stage points to the exact
previous digest and the first missing stage is preserved. Report:

- linked stages;
- first missing stage index or name;
- tamper checks performed;
- source, declaration, generation, reverse-observation, and verifier digests;
- the next non-executing operation, if any.

A broken or incomplete chain is a measured result, not a successful fallback.

### 4. Domain utility

Utility is measured from versioned scenarios, not inferred from language size.
For a scenario set `T`, report:

```
utility_yield = scenarios_with_verified_terminal_receipt / |T|
```

Each scenario must retain its input declaration, expected boundary, exact
receipt, and failure/unknown reason. The metric is meaningful only for the
specified scenario set and cannot be generalized to the whole language.

### 5. Investment efficiency

Investment decisions MUST record the marginal change in the preceding metrics
alongside the actual resources used:

- engineering effort;
- CI duration and attempt count;
- runtime or storage cost;
- newly covered scenarios;
- newly resolved unknown stages;
- regressions and reversals.

No single efficiency ratio is a completion score. When uncertainty or evidence
quality is low, prefer a small reversible experiment over a broad claim.

## Investment policy

Select the next domain slice using a written decision record containing:

1. user or system value of the slice;
2. current available, deferred, unknown, and out-of-scope counts;
3. provenance and security risk;
4. estimated effort and CI budget;
5. expected measurable change and how it will be falsified;
6. rollback or abandonment condition.

A slice may be promoted only when its exact-head CI evidence and receipts are
available. A lower metric after a declaration or scenario change is a new
measurement, not a failure to be hidden by averaging.

## Receipt invariants

A valid receipt MUST satisfy all of the following:

- execution and authorization are explicitly separate from observation;
- `UNKNOWN` preserves the first missing stage and cause;
- `DEFERRED` preserves the next non-executing operation;
- all digests validate against their bound content;
- ordered capability identifiers are preserved;
- cache presence is never used as semantic evidence;
- the exact source, toolchain, contract, and CI run are recorded;
- the receipt can be independently rejected when any bound input is tampered.

These invariants make the metric durable even when the observed value rises or
falls across revisions.
