# Reproducibility is not meaning

This is a deliberately small Gooo meta-code experiment. It keeps two claims
independent:

- `BYTE_REPRODUCIBILITY`: a producer and consumer compare specified artifact
  bytes or their cryptographic digests.
- `MEANINGFUL_RESULT`: a consumer compares an observed result with an
  independent meaning oracle.

The four case values are declared by `computes` programs on the four case
activities in `main.gooo`. The producer parses and lowers that source before
writing a receipt; the production consumer in
`internal/meta/reproducibilitysemanticsconsumer` independently parses and
lowers it again. The wire types live in the schema-only
`internal/meta/reproducibilitysemanticsschema` package, so the consumer has
zero producer-package imports.
The receipt is read-only evidence: repository writes, mutation authority, and
promotion authority are all zero or false.

| Case | Byte claim | Meaning claim | Joint status | Counterexample |
| --- | --- | --- | --- | --- |
| `both-discharged` | DISCHARGED | DISCHARGED | DISCHARGED | no |
| `reproducible-but-wrong` | DISCHARGED | REFUTED | REFUTED | yes |
| `meaningful-but-unreproduced` | REFUTED | DISCHARGED | REFUTED | yes |
| `claims-open` | OPEN | OPEN | OPEN | no |

The fixed receipt coordinates are matrix `4/4`, byte `2/4`, meaning `2/4`,
joint `1/4`, counterexamples `2/4`, open cases `1/4`, source-digest binding
`4/4`, and semantic causality `4/4`. The denominators are part of the
independent judge contract, not inferred from the numerator.

Conformance and subject resolution are separate. The baseline matrix is
`conformance_decision=DISCHARGED` with `conformance_resolution=EXACT`, while
the unresolved `claims-open` evidence keeps the subject at
`subject_decision=OPEN`, `subject_resolution=LOWER_RESOLUTION`, and
`subject_reason=OPEN_EVIDENCE_REMAINS`. A conforming matrix is therefore not
reported as a complete subject result.

Every case persists three transitions: byte, meaning, and joint. Each records
`OPEN -> DISCHARGED`, `OPEN -> REFUTED`, or remaining `OPEN`, with a fixed
`1/1` claim coordinate, stage, step, reason, and evidence digest. The
consumer recomputes all three transitions from the receipt evidence and
rejects transition drift.

Source-digest binding and semantic causality are separate indicators. A
receipt that supplies only the raw source digest is `REFUTED` as
`DIGEST_ONLY_REFUTED`; equal bytes never substitute for the lowered source
meaning.

The CI contract persists exactly two independent intervention cases (fixed
denominator `2`, with no aggregate score): changing a declared meaning value
changes meaning and joint coordinates and their transitions, while adding a
comment changes the raw source digest but leaves the semantic digest, all
coordinates, and all transitions unchanged. The artifact itself has zero
repository writes and zero mutation/promotion authority.

## What the references justify

The experiment adopts the [Reproducible Builds definition](https://reproducible-builds.org/docs/definition/): with fixed source, build environment, and instructions, specified artifacts can be recreated bit-for-bit. That rule is admitted only as `BYTE_REPRODUCIBILITY` evidence. It does not discharge a meaning claim.

It also adopts Nix's [reproducibility check](https://nix.dev/manual/nix/2.34/advanced-topics/diff-hook.html) as a determinism comparison: a second result is checked against an existing store result. Nix's [declarative dependency pinning](https://nix.dev/tutorials/first-steps/reproducible-scripts) supports recording inputs, but neither source establishes that equal bytes mean the right behavior. The experiment therefore rejects that inference.

From Necula's Berkeley-hosted [proof-carrying code paper](https://people.eecs.berkeley.edu/~necula/Papers/sfpol_isss02.pdf), it adopts the producer/consumer split and machine-checkable proof choice. It rejects calling a byte hash a semantic proof when no meaning oracle or semantic specification was checked.

## Falsifiability

The judge must fail closed if a byte digest, meaning digest, source declaration,
case order, stage/step/reason, producer/consumer identity, proof choice, raw
source binding, semantic digest, transition evidence digest, or fixed
denominator is changed. A CI dependency guard also rejects any producer call
or producer-package import in the production consumer package. Thus the experiment can be disproved by changing either
evidence channel without changing the other; that is the point of the two
failure paths.

This does not claim a general build system, Nix execution, semantic equivalence
for arbitrary programs, a safety proof, or authority to change the repository.
