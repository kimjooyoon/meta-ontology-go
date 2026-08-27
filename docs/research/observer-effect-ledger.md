# Observer-effect ledger: principles and limits

## Experiment claim

An observation is not equivalent to an effect-free observation. The observer
may read a repository, environment, or clock and may emit output. The ledger
therefore has two disjoint sections:

* `observation` records the values and before/after digests that were measured.
* `effects` records whether the measurement changed a declared domain, including
  the output channel used to publish the evidence.

The experiment's fixed denominator is 12 indicators. The denominator remains 12
for the clean observation, the intentional repository-write counterexample, and
the UNKNOWN case. This makes a missing or unresolved result visible as lower
resolution rather than letting the producer improve its percentage by deleting
the hard row.

Every indicator and receipt carries `producer`, `consumer`, `meta_operation`,
and `proof_choice`. The producer is the observer, the consumer is the
independent judge, and the proof choice identifies whether the evidence is a
FOUNDATION, COHERENCE, or REGRESSION obligation. UNKNOWN evidence also carries
an explicit `stage`, `step`, and `reason`. A persistent claim record advances
from `CLAIMED` to `SUPPORTED`, `REFUTED`, or `UNKNOWN` with a sequence and prior
evidence digest.

## Relation to effect systems

Koka is a useful design reference because it tracks side effects in function
types and distinguishes pure and effectful computations. Its effect handlers
also make the meaning of an operation explicit at the handler boundary:
[Koka documentation](https://koka-lang.github.io/koka/doc/book.html).

The OCaml manual provides a second reference point: effect handlers describe
computations that perform user-defined effects, and unhandled effects are
forwarded to an outer handler. The manual also calls the original OCaml 5
interface experimental, which is a reminder that a model of effects has its
own compatibility boundary: [OCaml effect handlers](https://ocaml.org/manual/5.5/effects.html).

This repository does not claim that Go 1.27 has a static effect type system.
Instead, the `.gooo` source declares the semantic vocabulary and the Go
observer records a small runtime effect algebra. The independent judge checks
that the declared observations and effects remain separate.

## Relation to hermetic builds

Bazel defines hermetic builds around isolation from host changes and source
identity, while noting that tools and dependencies must be made explicit:
[Bazel hermeticity](https://bazel.build/concepts/hermeticity). Nix similarly
uses declared dependency references and a sandbox, but its reproducible-build
guidance warns that timestamps and other nondeterminism can still leak:
[NixOS reproducible builds](https://reproducible.nixos.org/).

The experiment adopts the useful boundary, not the strongest possible claim:
the clean run measures the checked-out root, a declared environment tuple, and
an injected logical clock; artifacts are written outside the repository. A
replay and an independent judge then verify the same evidence. This is a
hermetic-style observation boundary, not proof that the host kernel, filesystem
metadata, network, wall clock, or CI service had no influence.

## Falsification and limits

The claim is falsified if a clean run reports a changed repository digest, a
non-zero repository-write count, an undeclared output effect, a receipt or
ledger replay mismatch, or a mutation/promotion authority bit set to true. The
intentional `violate` mode is a concrete counterexample: it writes a marker in
the supplied root, and the independent judge must accept only the
`FAIL_CLOSED` subject decision with 11/12 satisfied indicators.

The experiment cannot observe effects outside its declared scope. In
particular, it does not prove that the operating system clock was untouched; it
uses `SOURCE_DATE_EPOCH` or the fixed value `0` as an injected logical clock.
It does not prove that a malicious process could not mutate the root between
the two scans. It also counts its own three external artifact writes by
declaration, so the ledger is evidence about a bounded protocol, not a claim
of metaphysical zero effect.
