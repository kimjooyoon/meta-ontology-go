# Evidence freshness as a meta value

This is a bounded philosophical experiment, not a general cache invalidation
system. It asks when a claim's justification stops supporting the current
claim, and which declared stage makes that boundary visible.

## Language-owned policy

The checked-in [Gooo source](../../examples/evidence-freshness/main.gooo)
contains formal `freshness` declarations for:

| declaration | value |
| --- | --- |
| `axes` | `subject material recipe environment runner verifier` |
| `comparison_policy` | `earliest_changed` |
| `prior_claim_state` | `OPEN` |
| `boundary_policy` | `logical_epoch_environment` |
| `raw_material_policy` | `raw_material_digest` |
| `semantic_policy` | `comments_ignored` |
| `claim_ledger_policy` | `open_discharge_or_preserve` |
| `effect_policy` | `read_only_ci_before_after` |

The canonical compiler parses the file, lowers ordinary declarations through
the repository bidirectional lowerer, validates the semantic IR, and compiles
these policy values. Go supplies the versioned policy schema, permitted stage
types, and invariants; it does not repeat a ten-row decision table.

## Six-axis tuple and two material digests

The tuple keeps six axes separate:

| axis | stage | changed reason |
| --- | --- | --- |
| `subject` | `SUBJECT_BINDING` | `SUBJECT_CHANGED` |
| `material` | `MATERIAL_CLOSURE` | raw or semantic material change |
| `recipe` | `RECIPE_RESOLUTION` | `RECIPE_CHANGED` |
| `environment` | `ENVIRONMENT_CAPTURE` | `ENVIRONMENT_CHANGED` |
| `runner` | `RUNNER_EXECUTION` | `RUNNER_CHANGED` |
| `verifier` | `VERIFIER_JUDGMENT` | `VERIFIER_CHANGED` |

The `material` member is a pair, not one collapsed cache key:

```text
material = { raw_digest, semantic_digest }
```

`raw_digest` hashes the supplied bytes. `semantic_digest` is the stable hash
of canonical semantic IR. Since comments are not semantic IR, a comment-only
change produces raw `STALE`, semantic `FRESH`, and `PASS` with reason
`RAW_MATERIAL_CHANGED_SEMANTIC_PRESERVED`. A semantic value change produces
raw `STALE`, semantic `STALE`, and a fail-closed claim observation at
`MATERIAL_CLOSURE`.

The decider compares axes in the source-declared order. When multiple tuple
members differ, the earliest declared stage wins. A missing source is not
treated as equal: it is `UNKNOWN` / `LOWER_RESOLUTION` at
`SUBJECT_BINDING/reconstruct-source/SOURCE_UNAVAILABLE`.

## Claim state and provenance

The `CLAIM_JUSTIFIED -> CLAIM_PRESERVED`, `CLAIM_STALE`, or `CLAIM_UNKNOWN`
transition is a freshness observation only. It is separate from the canonical
claim ledger:

```text
OPEN --(current evidence)--> DISCHARGED
OPEN --(stale evidence)----> OPEN
OPEN --(unknown evidence)--> OPEN
OPEN --(explicit refutation)-> REFUTED
```

Every ledger entry contains the prior digest, receipt digest, raw source
digest, semantic digest, stage/step/reason provenance, and its own digest.
The report retains the append-only chain and chain digest. No stale result is
silently converted into `REFUTED`.

## Research decisions

### Adopted from official material

- Nix models a derivation over precisely defined inputs, system, builder, and
  output paths whose identity incorporates build inputs. We adopt explicit
  material, recipe, environment, and runner boundaries, but retain separate
  raw and semantic digests. See the [Nix Reference Manual: Derivations](https://nix.dev/manual/nix/2.22/language/derivations.html?highlight=derivation).
- SLSA provenance separates the produced subject, dependencies, builder, and
  consumer verification concerns. We adopt the producer/consumer split and
  make verifier judgment an explicit tuple axis. See [SLSA Provenance](https://slsa.dev/spec/v1.2/provenance) and [SLSA Build Requirements](https://slsa.dev/spec/v1.2/build-requirements).
- in-toto makes steps, functionaries, materials, products, expiry, and rules
  explicit. We adopt the stage/step/reason and expiry shape, while refusing
  unspecified artifacts as an implicit freshness pass. See [in-toto Getting Started](https://in-toto.io/docs/getting-started/).

### Rejected for this experiment

- One store path or output digest is not a sufficient claim-freshness key: it
  can hide recipe, runner, verifier, semantic, or temporal changes.
- Producer-generated provenance is not self-verifying. The decider imports no
  producer package and reconstructs source/raw/semantic identity itself.
- Wall-clock scheduler state, network reachability, generic cache eviction,
  signature authenticity, and full compiler semantic correctness are outside
  this bounded claim.

## Fixed observation and falsifiability

The exact-head/current-context observation is one `CURRENT_EVIDENCE` case. The
other nine are explicitly `SYNTHETIC_COUNTEREXAMPLE` cases: comment-only,
semantic value, five non-material tuple mutations, expired boundary, and
source unavailable.

The fixed results are:

- cases `10`; current `1`; synthetic `9`
- raw freshness: fresh `7`, stale `2`, unknown `1`
- semantic freshness: fresh `8`, stale `1`, unknown `1`
- claim freshness: fresh `2`, stale `7`, unknown `1`
- source reconstruction `9/9`; source unavailable `1/1`
- freshness transitions `10/10`; claim-ledger entries `10/10`
- claim ledger: discharged `2`, open-preserved `8`, refuted `0`
- coupling axes `6/6`; read-only before/after `0 -> 0`
- forbidden producer dependency count `0`; independence contract `1/1`

The model is falsified if a comment-only intervention changes semantic
freshness or claim decision, if a semantic value change remains semantically
fresh, if source unavailability becomes fresh, if an expired epoch is
preserved, if the decider imports the producer, if the ledger chain breaks, or
if CI's before/after write-set differs. The expected case labels classify the
experiment; the observed decision is produced by the independent decider.
