# Capability-scoped expansion

Status: an independent, read-only philosophy experiment. This PR does not
promote a language-readiness obligation, change the compiler kernel, or claim a
general macro sandbox.

## Proposition

Compile-time and expansion-time meta code must declare four authority kinds:
file, time, environment, and network. The default is deny. A request is
`ALLOW` only when the exact Gooo source declaration, the capability value, and
the matching evidence value agree. An `ALLOW` receipt is therefore a bounded
proof about one expansion request, not a permission to mutate the repository.

`examples/capability-scoped-expansion/main.gooo` is the same source for the
positive and negative cases. The witness emits `allow-receipt.json` and
`deny-receipt.json` with the same `source_digest`; the negative request changes
only its capability target to an undeclared network target. This is the
experiment's central comparison.

## Gooo-specific relation

This is not an ACL copied into a Go command. The `.gooo` source declares
activities that produce typed capability values. A request carries those
values, evidence is keyed to each value, and the expansion receipt records the
relation between source, value, evidence, stage, and step. The independent
judge recomputes that relation from raw JSON and source bytes.

That relation is the useful Gooo-specific claim: permission is data in the
language's semantic graph, and proof is another typed edge in the graph. The
producer's decision field has no authority by itself. The experiment does not
claim that the current Gooo parser has become a general macro language.

## Stages and epistemic states

The only exact stage in this experiment is `expand`, at step
`authorize-before-expand`. An absent source declaration, stage, step, or
evidence does not become a denial that looks complete. It produces:

```json
{"decision":"UNKNOWN","unknown":{"stage":"expand","step":"bind-capability-evidence","reason":"EVIDENCE_UNOBSERVED"}}
```

Claims have an explicit lifecycle:

| Situation | capability-scope-exact | default-deny | effect-ceiling |
| --- | --- | --- | --- |
| exact allow | `DISCHARGED` | `DISCHARGED` | `DISCHARGED` |
| known undeclared/effectful request | `REFUTED` | `DISCHARGED` | `DISCHARGED` |
| missing stage, step, or evidence | `OPEN` | `OPEN` | `OPEN` |

`UNKNOWN` is never promoted to `ALLOW`, and `REFUTED` is never laundered into
`DISCHARGED`.

## Fixed denominator

Every receipt has exactly 12 indicators. Each row carries a producer, consumer,
meta-operation, and proof choice so the metric cannot silently lose its
semantic owner or proof mode.

| # | Class | Proof choice | Indicator | Producer | Consumer | Meta-operation |
| ---: | --- | --- | --- | --- | --- | --- |
| 1 | DRIVER | FOUNDATION | source shape | `capabilityscopedexpansion.Evaluate` | `ci-capability-expansion-gate` | `expand-capability-scoped-meta-code` |
| 2 | DRIVER | FOUNDATION | Go 1.27 pin | `capabilityscopedexpansion.Evaluate` | `ci-capability-expansion-gate` | `expand-capability-scoped-meta-code` |
| 3 | DRIVER | FOUNDATION | file declaration | `capabilityscopedexpansion.Evaluate` | `ci-capability-expansion-gate` | `expand-capability-scoped-meta-code` |
| 4 | DRIVER | FOUNDATION | time declaration | `capabilityscopedexpansion.Evaluate` | `ci-capability-expansion-gate` | `expand-capability-scoped-meta-code` |
| 5 | DRIVER | FOUNDATION | environment declaration | `capabilityscopedexpansion.Evaluate` | `ci-capability-expansion-gate` | `expand-capability-scoped-meta-code` |
| 6 | DRIVER | FOUNDATION | network declaration | `capabilityscopedexpansion.Evaluate` | `ci-capability-expansion-gate` | `expand-capability-scoped-meta-code` |
| 7 | OUTCOME | COHERENCE | value/evidence relation | `capabilityscopedexpansion.Evaluate` | `ci-capability-expansion-gate` | `expand-capability-scoped-meta-code` |
| 8 | OUTCOME | COHERENCE | source binding | `capabilityscopedexpansion.Evaluate` | `ci-capability-expansion-gate` | `expand-capability-scoped-meta-code` |
| 9 | OUTCOME | COHERENCE | stage/step order | `capabilityscopedexpansion.Evaluate` | `ci-capability-expansion-gate` | `expand-capability-scoped-meta-code` |
| 10 | DRIVER | COHERENCE | receipt seal | `capabilityscopedexpansion.Evaluate` | `ci-capability-expansion-gate` | `expand-capability-scoped-meta-code` |
| 11 | GUARDRAIL | REGRESSION | default deny | `capabilityscopedexpansion.Evaluate` | `ci-capability-expansion-gate` | `expand-capability-scoped-meta-code` |
| 12 | GUARDRAIL | REGRESSION | authority ceiling | `capabilityscopedexpansion.Evaluate` | `ci-capability-expansion-gate` | `expand-capability-scoped-meta-code` |

The fixed conformance denominator is eight cases: one exact allow, four
undeclared-capability denials, two effect-ceiling denials, and one UNKNOWN
missing-evidence case. The expected result is `8/8`, with `ALLOW=1`, `DENY=6`,
and `UNKNOWN=1`.

## Permission and effect boundary

The fixed suite requests 32 capability values. Exactly 4 are authorized by
the exact allow case, 24 are denied, and 4 remain UNKNOWN because evidence is
missing. The blocked effect probes include one repository-write request and
one mutation-authority request. Observed repository writes remain `0`,
mutation authority remains `false`, promotion authority remains `false`, and
the toolchain is `go1.27.0`.

The network capability is a pinned symbolic target and is never contacted. The
time capability is a deterministic logical clock and is not a wall-clock
read. These choices make the receipt reproducible; they do not demonstrate OS
enforcement.

## Research: what is adopted and rejected

The following are primary sources, used as design evidence rather than as
claims that Gooo implements those systems.

1. [Racket Reference: Evaluation Model, Phases and Separate Compilation](https://docs.racket-lang.org/reference/eval-model.html#%28part._phases%29)
   explains the separation of execution-time phase 0 from expansion-time
   phase 1, and distinguishes internal effects from externally observable I/O.
   Adopted: make expansion stage and step explicit, and keep effect evidence
   visible. Rejected: treating a phase label or a separate-compilation
   guarantee as permission to discard external effects; Gooo requires an
   explicit capability and receipt instead.
2. [Rust Reference: Procedural Macros](https://doc.rust-lang.org/reference/procedural-macros.html)
   defines procedural macros as compile-time syntax-to-syntax functions and
   explicitly says they have the compiler's standard I/O and file access,
   with the same security concerns as build scripts. Adopted: keep the
   expansion boundary separate from runtime values. Rejected: inheriting the
   compiler's ambient resources; this experiment makes file, time,
   environment, and network authority explicit and default-deny.
3. [Miller, Yee, and Shapiro, Capability Myths Demolished](https://srl.cs.jhu.edu/pubs/SRL2003-02.pdf)
   compares capability models and identifies least privilege and confused
   deputy avoidance as practical advantages. Adopted: treat authority as a
   narrowly scoped value that must be presented to the operation. Rejected:
   claiming that a JSON receipt is a complete object-capability system; this
   PR has no OS-level confinement, revocation, delegation, or cryptographic
   authority.

## What remains falsifiable

The experiment can be refuted if an independent implementation accepts an
undeclared target, accepts missing evidence as `ALLOW`, changes the source
digest between the allow and deny receipts, observes a repository write, or
disagrees with the fixed eight-case result. It can also be shown insufficient
if the bounded source-marker reader is mistaken for a complete Gooo parser, if
the judge is not actually independent, or if a future implementation needs
real wall-clock, network, or OS sandbox semantics. Those are open engineering
questions, not hidden credit in this receipt.
