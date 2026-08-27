# Capability-scoped expansion correction

This PR is an independent philosophy experiment, not a language-readiness or
general macro-sandbox claim. It corrects the first version by making the
semantic `.gooo` values and live provider observations authoritative.

## Authority boundary

`examples/capability-scoped-expansion/main.gooo` uses the existing Gooo
`computes "..."` value-program form. The producer and consumer both run:

```text
syntax.ParseFile -> bidir.Lower -> semantic.IR -> canonical semantic digest
```

The lowered activity value programs define, as language values:

- `capability.policy`: default `DENY`, authorization mode, effect ceiling, and
  prior claim state;
- `capability.declare`: value id, kind, operation, target, policy id, prior
  claim state, and evidence class;
- `capability.operation`: expansion stage, authorization step, and prior claim
  state;
- `capability.case`: request values and effect probes, without an expected
  decision field.

Comments, activity names, and source substrings are not inputs to the
authority decision. The expected result is reconstructed from the lowered
policy/case values and the raw provider wire.

## Safe vertical slice

The provider performs only two live observations:

1. reads a CI-created pinned file in a temporary directory and records its
   content digest as `CURRENT_EVIDENCE`;
2. reads a deterministic logical input `logical-clock:0` as
   `CURRENT_EVIDENCE`.

Environment and network are deliberately not contacted. Environment is
`UNKNOWN`; the network declaration is `HISTORICAL_FIXTURE`. Thus the fixed
capability denominator is four declarations, but only `2/4` are
`CURRENT_EVIDENCE`.

The sandbox provider exposes real request methods for repository write,
mutation, and promotion. Each request returns `DENIED`, records the before and
after sandbox digests, and performs no write. If an enforcement result is not
observed, the producer and consumer emit
`CAPABILITY_ENFORCEMENT_NOT_IMPLEMENTED / LOWER_RESOLUTION`, never a constant
denial.

## Fixed denominator and observed result

The source has eight semantic cases. No case stores an expected decision; the
producer and consumer independently derive it from policy, request values,
and raw observations.

| fixed cases | ALLOW | DENY | UNKNOWN |
| ---: | ---: | ---: | ---: |
| 8 | 1 | 6 | 1 |

The expected CI artifact reports:

| metric | value |
| --- | ---: |
| capability requests | 9 |
| authorized | 2 |
| denied | 6 |
| UNKNOWN | 1 |
| CURRENT_EVIDENCE | 2/4 |
| HISTORICAL_FIXTURE declarations | 1 |
| enforcement observations | 3/3 |
| blocked write probes | 1 |
| blocked mutation probes | 1 |
| actual repository writes | 0 |
| actual mutation authority | false |
| actual promotion authority | false |

## Claims and interventions

Every semantic capability starts with prior state `OPEN`. Each receipt appends
claim transitions containing stage, step, reason, evidence digest, and
provenance:

- successful live observation: `OPEN -> DISCHARGED`;
- missing live observation: `OPEN -> OPEN`;
- explicit target/policy violation: `OPEN -> REFUTED`.

The artifact contains two interventions:

- semantic policy intervention changes `authorization=exact-current` to
  `authorization=deny-all`; the allow case changes from `ALLOW` with a
  `DISCHARGED` scope claim to `DENY` with `REFUTED`;
- comment-only intervention changes the raw source digest but preserves the
  canonical semantic digest, decision, and claim transition.

## Independent consumer

`internal/meta/capabilityscopedexpansion/verify` imports only `syntax` and
`bidir` from the language boundary. It receives raw `.gooo`, raw provider
observations, and raw receipt JSON. It reconstructs the semantic policy,
declarations, cases, evidence availability, effect enforcement, claim
transitions, and receipt digest itself. It does not import the producer. CI
also checks that producer-package imports are absent from the consumer source.

The artifact records source reconstruction `1/1` and producer imports `0/1`.

## Research basis

The design adopts a phase boundary from the [Racket Reference evaluation
model](https://docs.racket-lang.org/reference/eval-model.html#%28part._phases%29),
but rejects treating phase separation as permission to discard observable
effects. It adopts the explicit compile-time resource warning from the [Rust
Reference procedural macro](https://doc.rust-lang.org/reference/procedural-macros.html)
section, but rejects inheriting ambient compiler resources. It adopts narrow
authority and confused-deputy avoidance from [Miller, Yee, and Shapiro,
Capability Myths Demolished](https://srl.cs.jhu.edu/pubs/SRL2003-02.pdf), but
rejects treating a JSON receipt as a complete object-capability system.

## Remaining falsifiability

An independent implementation can still refute this slice by accepting an
undeclared target, treating missing environment evidence as current, accepting
a missing provider observation, changing a comment-only semantic digest,
disagreeing with the raw before/after enforcement result, or importing the
producer into the consumer. The slice does not yet provide OS-level
confinement, revocation, delegation, cryptographic authority, real network
access, or wall-clock observation. Those limitations remain explicit rather
than being counted as current evidence.
