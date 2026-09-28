# Domain completeness and investment vector

Status: proposal. This document defines a domain-scoped measurement contract.
It does not claim that a metric can determine truth, intelligence, or universal
language completeness.

## 1. Problem statement

"Completeness" is only meaningful relative to an explicit domain, a declared
scope, and an evaluator that can produce evidence for that scope. A passing
code-generation test proves that one generation predicate passed. It does not
prove that a domain is semantically complete, that generated code is correct
for every input, or that an unmeasured boundary is safe.

The repository therefore treats a domain as a bounded evidence program, not as
a scalar score:

```text
DomainProfile = {
  domain_id,
  profile_version,
  scope_manifest_digest,
  required_units,
  excluded_units,
  evaluator_catalog_digest,
  boundary_policy_digest,
  investment_policy_digest
}
```

Every field is content-addressed. A profile change creates a new observation
context; it must not silently reinterpret historical measurements.

## 2. Scope before measurement

`required_units` are the smallest stable claims that the profile intends to
measure. A unit has:

```text
Unit = {
  unit_id,
  kind,                 # grammar, IR, generation, reverse, LSP, security, ...
  question,
  authority_inputs,
  evaluator_id,
  predicate,
  evidence_roles,
  boundary_effect
}
```

`excluded_units` are explicit non-goals. An excluded capability is not counted
as a failure and cannot be converted to `NOT_APPLICABLE` without a profile
record proving that it is excluded for the current snapshot. This prevents a
missing feature from disappearing merely because no evaluator was written.

The first gooo/jev profile should keep the following scope separate:

1. declaration parsing and source identity
2. normalized IR construction
3. deterministic source generation
4. generated-source reverse observation
5. LSP provenance projection
6. non-executing and non-authorizing capability boundaries

Execution, general semantic correctness, workload identity issuance, and
external mutation are separate future profiles. They must not be smuggled into
the first profile through a larger score.

## 3. Evidence vector

The primary result is a vector of independently interpretable observations:

```text
DomainEvidenceVector = {
  scope_coverage,
  evaluator_coverage,
  evidence_binding,
  replay_determinism,
  boundary_coverage,
  counterexample_retention,
  unknown_resolution,
  resource_cost
}
```

Each dimension is a record, never a bare percentage:

```text
EvidenceDimension = {
  dimension_id,
  numerator,
  denominator,
  unit,
  normalization,
  predicate,
  decision,             # PASS | FAIL_CLOSED | UNKNOWN | NOT_APPLICABLE
  evaluation_state,     # EVALUATED | DEFERRED | NOT_RUN | ERROR
  evidence_refs,
  snapshot_digest,
  observed_at
}
```

The denominator is mandatory. A numerator without a declared population is
not a coverage claim. An empty population is a separate state and is never
silently treated as 100 percent.

The dimensions have distinct meanings:

- `scope_coverage`: required units with an evaluator and retained evidence.
- `evaluator_coverage`: required predicates that can be executed by the
  declared evaluator catalog.
- `evidence_binding`: observations whose source, toolchain, policy, run, and
  artifact identities are exact-bound.
- `replay_determinism`: repeated evaluation over one pinned snapshot produces
  the same normalized value, decision, and failure code.
- `boundary_coverage`: declared execution, authorization, and mutation
  boundaries have explicit observations, including negative cases.
- `counterexample_retention`: known contradictions remain queryable rather
  than being overwritten by later positive observations.
- `unknown_resolution`: previously unresolved units that received new
  independent evidence. This measures observability progress, not correctness.
- `resource_cost`: CI time, changed surface, storage, and explicitly declared
  engineering investment for the observation. Cost is not quality.

No dimension is allowed to compensate for another dimension. For example, a
fast generator cannot offset missing reverse observation, and high coverage
cannot offset an unverified authorization boundary.

## 4. States and completeness claim

The vector uses closed states:

```text
UNOBSERVED -> OBSERVED -> EXACT_BOUND -> CLOSED
                         \-> UNKNOWN
                         \-> REFUTED
```

- `UNOBSERVED`: no reproducible observation exists.
- `OBSERVED`: a value exists but one or more authority inputs are incomplete.
- `EXACT_BOUND`: inputs, normalization, predicate, and artifacts are bound to
  one immutable snapshot.
- `CLOSED`: every required dimension meets its profile predicate and no
  required unit is unresolved.
- `UNKNOWN`: the evaluator cannot establish the result without inventing an
  input or interpretation.
- `REFUTED`: retained evidence contradicts the predicate.

`CLOSED` means closed for this profile and snapshot only. It is not a claim of
universal correctness or language completion. A profile may report a
`closure_reason` and `remaining_exclusions`, but it must not collapse them into
a global quality score.

`UNKNOWN`, `DEFERRED`, and `NOT_RUN` are never PASS. An evaluator error is not
rewritten as a domain failure, and a successful unrelated CI job is not
evidence for an unrun dimension.

## 5. Investment and prioritization

Work is allocated with an explicit investment record:

```text
InvestmentItem = {
  item_id,
  target_domain_id,
  target_dimension_id,
  target_unit_ids,
  hypothesis,
  budget_ci_minutes,
  budget_engineer_hours,
  expected_evidence_gain,
  stop_condition,
  replan_condition,
  candidate_digest,
  result_refs
}
```

`expected_evidence_gain` is a planning value. It must never be reported as an
observed improvement until the evaluator emits a new exact-bound vector.

The default priority order is:

1. repair missing authority inputs and capability-boundary evidence
2. resolve high-impact `UNKNOWN` units with an independent evaluator
3. increase scope or evaluator coverage inside the declared profile
4. improve replay, counterexample retention, and provenance quality
5. optimize resource cost without lowering any existing predicate

The priority order is a policy, not a truth ranking. A profile may override it
only with a digest-bound investment policy and an explicit reason. A metric
regression lowers the affected dimension or produces `REFUTED`; it does not
permit a compensating increase in another dimension.

## 6. Philosophical boundary

This model deliberately rejects a universal answer to "how complete is the
language?". A language can be complete for a declared workflow while being
incomplete for execution, security, or a different domain. The honest unit of
progress is therefore:

```text
new evidence for a named claim under a named boundary and budget
```

The model also rejects information asymmetry as a design goal. Provenance is
valuable because it makes origin, transformation, and uncertainty inspectable;
it is not a shortcut that recovers arbitrary missing information.

## 7. gooo/jev example profile

For the initial gooo/jev profile, the following observations are distinct:

| unit | positive evidence | does not prove |
| --- | --- | --- |
| declaration parsing | accepted syntax and source digest | general semantic validity |
| IR normalization | deterministic IR digest | domain truth |
| source generation | canonical generated text | runtime behavior |
| reverse observation | generated source reparses to the same IR digest | all transformations preserve intent |
| LSP provenance | exact source/IR/evidence references | production-grade LSP support |
| capability boundary | non-executing/non-authorizing validation | secure workload identity issuance |

The profile can reach `CLOSED` only for this declared evidence loop. It cannot
reach that state by adding an execution claim that has no evaluator.

## 8. CI and code-generation integration

CI should publish two independent receipts:

1. `codegen_receipt`: parser, IR, generated source, reverse parse, and exact
   digest comparisons for the tested fixture cohort.
2. `domain_vector_receipt`: the complete evidence vector, profile digest,
   evaluator catalog digest, resource cost, and unresolved units.

The first receipt may be PASS while the second remains `UNKNOWN` or open. This
is expected and should be visible. Cache hits are only resource observations;
they are never semantic or test evidence.

Every receipt must include the exact commit, workflow, run and attempt,
toolchain, profile/catalog digests, artifact roles and digests, and expiry.
Promotion may consume a receipt only after the normal protected CI policy
accepts the exact tuple.

## 9. Versioning and adoption

This RFC is design-only until an implementation provides:

1. a machine-readable profile and dimension registry
2. an evaluator that emits the vector schema above
3. fixtures for PASS, UNKNOWN, DEFERRED, NOT_RUN, and REFUTED
4. replay tests proving stable output for a pinned snapshot
5. a CI artifact binding the vector to the exact source and toolchain

Adoption must move through `UNOBSERVED`, `OBSERVED`, `EXACT_BOUND`, `SHADOW`,
and finally `BLOCKING`. A new metric cannot become blocking merely because its
number is high or because a human considers it important.
