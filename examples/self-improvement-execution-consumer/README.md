# Bounded execution consumer

The consumer reads the exact v26 grant artifact and the exact v25 contract
artifact named by that grant. It checks their subject, source artifact digest,
policy resolution, and embedded execution input before compiling the immutable
`Increment` source snapshot.

One reservation artifact is written before the value evaluator runs. Its name
is derived from the grant request digest, and the workflow serializes runs for
the same grant. A later run checks for that reservation and refuses a second
consumption. If a run stops after reserving, the grant remains spent; recovery
requires a new system-derived grant.

The consumer evaluates all five fixed corpus cases as one bounded experiment.
The report distinguishes one consumed grant from five evaluator invocations.
It records expected and actual values, execution and result digests, source and
plan identity, per-case elapsed time, compile time, and total evaluation time.
Repository writes and external effects remain zero. Runtime samples are
diagnostic; without a comparable baseline, `performance_improvement` stays
`UNKNOWN`.
