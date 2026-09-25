# Exact-head transport experiment

This experiment separates three identities that a chained workflow must not
collapse:

- `D`: the producer-declared logical subject and its canonical payload digest;
- `R`: the workflow, run, attempt, job, and orchestration head;
- `T`: the immutable artifact ID, producer run, and archive digest.

The fixed `EHT-8` denominator is declared by the eight `Observe...` activities
in `transport.gooo`. The current CI can verify seven obligations. Authenticated
producer attestation remains `UNKNOWN`, so the receipt is deliberately
`OBSERVED / LOWER_RESOLUTION` at `7/8 = 8750` basis points. This lower
resolution authorizes only the existing read-only, non-executing candidate.
Any known identity or digest mismatch is `FAIL_CLOSED`.

Artifact names are lookup labels only. The consumer selects one artifact from
the exact producer run and attempt, downloads it by immutable artifact ID,
recomputes the archive digest, and then asks the Gooo metaprogram to verify the
producer receipt and logical payload.
