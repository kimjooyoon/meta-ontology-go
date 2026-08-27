# Reproducibility is not meaning

This is a deliberately small Gooo meta-code experiment. It keeps two claims
independent:

- `BYTE_REPRODUCIBILITY`: a producer and consumer compare specified artifact
  bytes or their cryptographic digests.
- `MEANINGFUL_RESULT`: a consumer compares an observed result with an
  independent meaning oracle.

The producer writes a receipt; the consumer is an independent judge that
recomputes both channels from the receipt and the `.gooo` source. The receipt
is read-only evidence: repository writes, mutation authority, and promotion
authority are all zero or false.

| Case | Byte claim | Meaning claim | Joint status | Counterexample |
| --- | --- | --- | --- | --- |
| `both-discharged` | DISCHARGED | DISCHARGED | DISCHARGED | no |
| `reproducible-but-wrong` | DISCHARGED | REFUTED | REFUTED | yes |
| `meaningful-but-unreproduced` | REFUTED | DISCHARGED | REFUTED | yes |
| `claims-open` | OPEN | OPEN | OPEN | no |

The fixed receipt coordinates are matrix `4/4`, byte `2/4`, meaning `2/4`,
joint `1/4`, counterexamples `2/4`, and open cases `1/4`. The denominators are
part of the independent judge contract, not inferred from the numerator.

## What the references justify

The experiment adopts the [Reproducible Builds definition](https://reproducible-builds.org/docs/definition/): with fixed source, build environment, and instructions, specified artifacts can be recreated bit-for-bit. That rule is admitted only as `BYTE_REPRODUCIBILITY` evidence. It does not discharge a meaning claim.

It also adopts Nix's [reproducibility check](https://nix.dev/manual/nix/2.34/advanced-topics/diff-hook.html) as a determinism comparison: a second result is checked against an existing store result. Nix's [declarative dependency pinning](https://nix.dev/tutorials/first-steps/reproducible-scripts) supports recording inputs, but neither source establishes that equal bytes mean the right behavior. The experiment therefore rejects that inference.

From Necula's Berkeley-hosted [proof-carrying code paper](https://people.eecs.berkeley.edu/~necula/Papers/sfpol_isss02.pdf), it adopts the producer/consumer split and machine-checkable proof choice. It rejects calling a byte hash a semantic proof when no meaning oracle or semantic specification was checked.

## Falsifiability

The judge must fail closed if a byte digest, meaning digest, source declaration,
case order, stage/step/reason, producer/consumer identity, proof choice, or
fixed denominator is changed. Thus the experiment can be disproved by changing
either evidence channel without changing the other; that is the point of the
two failure paths.

This does not claim a general build system, Nix execution, semantic equivalence
for arbitrary programs, a safety proof, or authority to change the repository.
