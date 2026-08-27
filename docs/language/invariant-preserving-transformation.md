# Invariant-preserving transformation authority

This experiment makes transformation permission conditional on four Gooo meta
values:

```text
precondition ∧ transformation ∧ postcondition ∧ regression_witness
```

The conjunction is an authority predicate, not a generic refactoring command.
The source contract is [the checked-in Gooo value model](../../examples/invariant-transformation/main.gooo).
Its producer emits a receipt containing the four values, their
`producer/consumer/meta_operation/proof_choice/verification_check` bindings,
and a `stage/step/reason` coordinate. The receipt contains separate baseline
and replay input/operation/output/digest observations. Each fixed case is an
actual `computes` value in the Gooo source: an `int64` input, an `add:n`
candidate, an expected value, an invariant relation, a replay recipe/capability,
and an effect policy. Producer and judge parse and execute the candidate and
recipe independently; a replay label alone is never evidence.

## State and effect model

Each claim starts `OPEN` and records one explicit, case-qualified transition to
`DISCHARGED`, `REFUTED`, or `OPEN`. The transition carries proposition/target,
prior-state, evidence, previous-transition, and current-transition digests.
`DISCHARGED` means the required evidence is present and bound. `OPEN` means
evidence is absent; it never becomes authorization by optimism or by a green
producer result.

The producer emits a `PROVISIONAL_NO_EFFECT` receipt. Only the independent
judge's exact, digest-bound authorization is accepted by the separate
post-judgment executor. The executor writes one artifact under `RUNNER_TEMP`;
the judge and consumer read and verify its actual bytes, path, size, digest,
case, subject SHA, authorization digest, and execution provenance. The
`RepositoryMutationAuthorized` field is scoped to this protocol's repository
path authorization; ambient process/OS authority and transient writes remain
`UNKNOWN`.

The four-case denominator is deliberately mixed: one preserved translation,
one semantic violation, one missing replay recipe, and one approved artifact.
Temporary filesystem mutation is reported separately as
`temp_artifact_write_authorized=true`; repository net content is observed by a
raw tracked+untracked `path,digest` artifact immediately before and after each
witness execution. The independent report consumer checks that artifact's
address, byte digest, entry count, sorted entry projection, exact execution ID,
and exact HEAD before deriving `NET_REPOSITORY_CONTENT_STATE_UNCHANGED`.
`repository_actual_or_transient_writes=UNKNOWN` remains separate from that
content observation. `AUTHORIZED` is scoped to the bounded transformation
receipt or temporary artifact emission; it does not grant repository edit or
promotion authority.

The current exact receipt is `4/4` bounded source-derived cases and `10000`
basis points, with `2` independently authorized cases, `1` refuted case, and
`1` open case. It records `16` unique case-qualified claim instances from `4`
templates, `16/16` accepted transitions, `4/4` bounded input-domain
observations, `4` provisional receipts, `2` authorization receipts, `1`
executed effect, `1` independently observed effect, and `1` unknown transient
effect scope. Across the claims, `13` are discharged, `2` refuted, and `1`
open. These are bounded witness counts, not a claim that the violation or
evidence gap is a successful transformation.

## Research basis and limits

The design follows the per-translation viewpoint of [George C. Necula,
“Translation Validation for an Optimizing Compiler,” PLDI 2000](https://dl.acm.org/doi/10.1145/349299.349314): check the concrete output of each transformation and use an explicit simulation-style witness where possible. It also follows [Nik Sultana and Simon Thompson, “Mechanical Verification of Refactorings”](https://kar.kent.ac.uk/23959/), which formalizes refactorings in Isabelle/HOL and makes behavior preservation the correctness condition.

The experiment is smaller and weaker than either line of work. Program
equivalence is undecidable; finite evidence cannot prove arbitrary programs or
all effects. A validator can be incomplete or report false alarms when it
cannot explain a transformation, and validation has runtime cost. This model
therefore makes only a synthetic, bounded claim over four explicitly declared
integer fixtures and the single declared input-domain observation per case; it
does not authorize arbitrary transformations or claim a verified refactoring
engine, complete semantic equivalence, toolchain correctness, repository
mutation, or promotion authority. Changing the `.gooo` expected value or
candidate recipe changes the independently recomputed result to `REFUTED`,
while an unavailable recipe leaves it `OPEN`.

## Intervention separation

The intervention witness is separate from the four-case authority score. It
publishes three non-aggregated fixed denominators, each `1/1`:

* The semantic-expected slice changes `expected=3` to `expected=4`. Its parsed
  projection, receipt, and independent decision change from `AUTHORIZED` to
  `REFUTED` with reason `SEMANTIC_POSTCONDITION_REFUTED`.
* The semantic-operation slice changes the candidate and replay recipe from
  `add:1` to `add:2`, changing the candidate output, postcondition, and
  decision.
* The non-semantic slice appends only whitespace and a comment. Its raw
  `SourceDigest` and receipt digest change, while the parsed/lowered fixture
  projection, replay output and semantic digest, decision, resolution, reason,
  claim transition outcome, and effect remain equal. Repository write count
  remains unobserved (`-1`, `repository_writes_observed=false`) in both
  receipts.

The intervention claims use exact stage/step/reason coordinates and persistent
digests. The effect gate adds six independently adjudicated observations:
unauthorized, refuted, open, stale-SHA, and tampered-authorization variants
create zero artifacts; the valid authorized case creates exactly one observed
artifact. A wrong or stale authorization fails closed. `DeterministicReplay`
is explicitly a producer repeatability check, not independent evidence. The
separate `interventionconsumer` package uses its own wire structs and
functions, parses the raw `.gooo`, lowers it canonically, executes each
candidate and replay recipe, and compares receipts, decisions, coordinates,
transitions, and effects without importing producer or calling `Build`.

The validator expectation contract is bound by a separate digest. It labels
expected outcomes only and cannot supply source case inventory or executable
recipes. CI changes one case outcome together with its claim, denominator, top
decision, and resealed digest, then proves that the source-bound consumer
rejects the coherent-looking artifact.

CI exposes fixed evidence counts: production judge/consumer producer imports
`0/0`, reconstructed intervention cases `3/3`, actual replays `3/3`, one
artifact evidence object observed from actual bytes, effect gates `8/8`, and
coherent tamper rejection `1/1`. Repository content snapshots use a fixed
`1/1` denominator with raw-entry provenance; forged summaries, raw-entry
tampering, content-only changes, and cross-execution bindings fail closed.
The correction denominator is `12/12`.
