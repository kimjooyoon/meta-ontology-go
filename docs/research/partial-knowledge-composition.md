# Partial-knowledge composition calculus

This experiment studies whether several meta-operation observations can be
composed without confusing direct unknown knowledge, an unresolved upstream
dependency, and a known invariant. The five source recipes in
`examples/partial-knowledge-composition/main.gooo` are the fixed corpus. A
separate CI provider creates raw evidence from the source, workspace snapshots,
and capability observation; the source itself never self-attests an observed
result.

## Formal basis and adopted rules

The abstract-interpretation basis is P. Cousot and R. Cousot, “Abstract
interpretation: a unified lattice model for static analysis of programs by
construction or approximation of fixpoints,” POPL 1977. Its ordered,
approximation-oriented view supports a precision loss without inventing a
more precise fact. This experiment adopts that non-invention rule: composition
may preserve an unresolved fact or invariant, but may not promote it to
success.

The three-valued basis is the strong Kleene treatment described in the
Stanford Encyclopedia account of Peirce’s three-valued logic. Unknown is not
silently converted to true or false by composition. The information-order
perspective is cross-checked with the Cambridge treatment of truth versus
information in logic programming. The sources are:

- [Cousot and Cousot, POPL 1977](https://cs.nyu.edu/~pcousot/COUSOTpapers/POPL77.shtml)
- [Peirce’s three-valued logic, Stanford Encyclopedia of Philosophy](https://plato.stanford.edu/archives/sum2018/entries/peirce-logic/three-valued-logic.html)
- [Truth versus information in logic programming, Cambridge Core](https://www.cambridge.org/core/journals/theory-and-practice-of-logic-programming/article/truth-versus-information-in-logic-programming/FCE4AEFF496594C839719E2EC2A0DEDE)

Adopted principles:

1. The source declares a required proposition, its observation recipe,
   dependency recipe, invariant capability, producer/consumer, meta-operation,
   and proof choice. The CI provider owns observed values and availability.
2. The producer and independent consumer each run
   `syntax.ParseFile -> bidir.Lower -> ir.Validate` and validate the raw
   receipt. The consumer independently rechecks the workspace evidence rather
   than copying an expected case table.
3. A dependency block must point to an explicit upstream claim with proposition
   digest, state, lifecycle stage/step/reason, evidence digest, source/semantic
   digests, and target address. It blocks only for upstream `OPEN` or
   `UNKNOWN`.
4. Claims are distinct by normalized proposition/predicate and target address,
   not by arbitrary `composition/<case>` labels. The fixed denominator is five
   source-derived predicates.
5. Tracked, untracked, and status snapshots are compared before and after the
   CI observation. A zero-write observation does not establish promotion
   authority; absent permission evidence remains `UNKNOWN` /
   `LOWER_RESOLUTION`.

Rejected principles:

- Generic optional/error propagation is rejected because it describes value
  transport rather than knowledge resolution.
- “Any non-error is success” is rejected because a known invariant does not
  prove the target operation.
- Replacing a dependency block with a direct unknown is rejected because it
  erases the upstream edge and prevents later resolution.
- `CALCULUS_PROVEN` is not a subject-success claim. It is limited to rule
  conformance; subject resolution, evidence coverage, and authority resolution
  are separate.

## Calculus and source-derived evidence

| Evidence composition | Derived state | Outcome | Claim transition |
|---|---|---|---|
| both observations are available and equal to required values | `EXACT` | `PASS` / `EXACT` | `OPEN -> DISCHARGED` |
| an observation is unavailable without a dependency | `DIRECT_UNKNOWN` | `UNKNOWN` / `LOWER_RESOLUTION` | `OPEN -> OPEN` |
| an observation is unavailable and its upstream claim is open/unknown | `DEPENDENCY_BLOCKED` | `UNKNOWN` / `LOWER_RESOLUTION` | `OPEN -> OPEN` |
| an observation is available with a known invariant capability | `INVARIANT_ONLY` | `HOLD` / `INVARIANT_ONLY` | `OPEN -> OPEN` |
| direct unknown and dependency block are both present | `MIXED_UNRESOLVED` | `UNKNOWN` / `LOWER_RESOLUTION` | `OPEN -> OPEN` |

The five source recipes yield one case for each row: one exact case, four
non-exact cases, four non-exact claims not promoted, one discharged claim, and
four open claims. Every case and claim records producer, consumer,
meta-operation, proof choice, stage, step, reason, normalized proposition,
proposition digest, target operation/output, raw source digest, semantic digest,
raw evidence digest, and evidence provenance.

## Falsifiability and interventions

The semantic A/B uses a real source variant: it changes the
`direct-unknown.left.observation_recipe` bytes from `missing` to `exact`, then
parses and lowers the variant. It must change the semantic IR digest and raw
evidence digest, and move the direct case from `UNKNOWN` /
`LOWER_RESOLUTION`, `OPEN -> OPEN` to `PASS` / `EXACT`, `OPEN -> DISCHARGED`.
The comment-only A/B changes only source bytes. It must change raw source and
raw evidence digests while preserving semantic IR, semantic projection,
case decisions/resolutions, and claim `from`/`to` transitions. Raw provenance
digests remain visible outside that semantic projection.

The exact-head Action wrapper reports source cases `5/5`, distinct predicates
`5/5`, semantic causality `1/1`, nonsemantic preservation `1/1`, open claims
preserved `4/4`, and the independent consumer import predicate as `1/1` with
an actual import count of `0`. It also publishes the actual repository-write
observation and the unresolved promotion capability. Any changed source
recipe, lowering result, dependency lifecycle, claim proposition/target,
digest, denominator, or evidence snapshot must fail the independent check.
