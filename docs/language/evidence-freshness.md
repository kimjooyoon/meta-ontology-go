# Evidence freshness as a meta value

This is a bounded philosophical experiment, not a general cache invalidation
system. It asks a narrower question: when a justification was valid for one
combination of facts, which stage first makes that justification unusable after
one part of the combination changes?

## Model

The receipt carries six independent identities:

| axis | stage | changed reason |
| --- | --- | --- |
| `subject` | `SUBJECT_BINDING` | `SUBJECT_CHANGED` |
| `material` | `MATERIAL_CLOSURE` | `MATERIAL_CHANGED` |
| `recipe` | `RECIPE_RESOLUTION` | `RECIPE_CHANGED` |
| `environment` | `ENVIRONMENT_CAPTURE` | `ENVIRONMENT_CHANGED` |
| `runner` | `RUNNER_EXECUTION` | `RUNNER_CHANGED` |
| `verifier` | `VERIFIER_JUDGMENT` | `VERIFIER_CHANGED` |

The decider compares the axes in that order. If more than one differs, the
earliest differing stage wins and `changed_dimensions` retains the observed
dimension. Missing values are `UNKNOWN`, never equal and never zero. A current
epoch beyond `valid_through_epoch` is `STALE` at `VERIFIER_JUDGMENT` with
`TEMPORAL_BOUNDARY_EXPIRED`. `environment_boundary` is a separate temporal
and environmental meta value, so a changed boundary cannot hide behind an
unchanged artifact identity.

The claim transition is explicit:

```text
CLAIM_JUSTIFIED --(exact tuple and current boundary)--> CLAIM_PRESERVED
CLAIM_JUSTIFIED --(changed tuple or expired boundary)--> CLAIM_STALE
CLAIM_JUSTIFIED --(missing or undecidable value)-------> CLAIM_UNKNOWN
```

`CLAIM_PRESERVED` is the only transition that permits an exact current claim.
All three transitions are retained in the read-only report with stage, step,
reason, and the receipt digest.

## Research decisions

### Adopted from official material

- Nix models a derivation as a specification over precisely defined inputs,
  system, builder, and output paths; its reference manual says output paths
  incorporate cryptographic hashes of build inputs and that all sources should
  reside in the store. We adopt the deterministic input-closure idea and make
  `material`, `recipe`, `environment`, and `runner` explicit tuple members.
  See the [Nix Reference Manual: Derivations](https://nix.dev/manual/nix/2.22/language/derivations.html?highlight=derivation).
- SLSA provenance identifies what was produced, how it was produced, by which
  builder, and from which dependencies. Its current requirements separate
  producer consistency, provenance completeness/authenticity/accuracy, and
  consumer verification. We adopt the producer/consumer split and the
  subject/material/recipe/environment/runner vocabulary, while adding the
  verifier as an explicit consumer-side axis.
  See [SLSA Provenance](https://slsa.dev/spec/v1.2/provenance) and
  [SLSA Build Requirements](https://slsa.dev/spec/v1.2/build-requirements).
- in-toto makes supply-chain steps, authorized functionaries, materials,
  products, and layout rules explicit, then verifies signatures, expiry, step
  authorization, and artifact rules. We adopt the stage/step/reason shape and
  the rule that a consumer must check the evidence against expectations.
  See [in-toto Getting Started](https://in-toto.io/docs/getting-started/).

### Rejected for this experiment

- A Nix store path or output byte digest is not a sufficient freshness key here:
  two equal outputs can carry different recipes, runners, verifiers, or time
  boundaries. We therefore do not collapse the six axes into one cache key.
- A producer-generated SLSA-like statement is not treated as self-verifying.
  SLSA distinguishes existence from authenticity and unforgeability; this
  experiment records no signature claim and requires an independent decider.
- in-toto's default allowance for unspecified artifacts is not used as a
  freshness rule. The contract enumerates a finite case set and fails closed on
  missing or unknown values rather than inferring that an absent edge is safe.
- Wall-clock timestamps, scheduler state, and network availability are not
  freshness evidence. Logical epochs and an environment boundary are pinned in
  the contract so the experiment is replayable.

## Fixed observation

The checked-in [Gooo source](../../examples/evidence-freshness/main.gooo) is
the producer subject. The contract has `10/10` satisfied cases: `1/1` fresh,
`7/7` stale, and `2/2` unknown. The six independent axes are exercised at
`6/6`; all `10/10` claim transitions are retained; the temporal boundary is
exercised at `1/1`; and the read-only guardrail is `1/1`. The stale-stage
distribution is subject `1`, material `1`, recipe `1`, environment `1`, runner
`1`, and verifier `2` (verifier identity change plus expiration). Unknown is
subject `1` and verifier `1`.

The producer and consumer are named metadata values in every metric:
`evidence-freshness-producer/v1` and `evidence-freshness-decider/v1`.
Meta-operation names and proof choices are contract-bound, not inferred from
the final count.

## Falsifiability and limits

The model is falsified if changing one isolated axis does not produce its
declared stage/reason, if a missing axis becomes `FRESH`, if an expired epoch is
preserved, or if a digest-valid but differently produced receipt passes through
the independent decider. It does not claim signature authenticity, complete
compiler semantics, generic cache behavior, wall-clock freshness, or mutation
authority.
