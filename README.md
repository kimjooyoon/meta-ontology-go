# meta-ontology-go

**Gooo — Go Of Ontology** is an experimental language for expressing intent,
assembling programs, and keeping the evidence of how they behave. This repository
contains its Go compiler. Source files use the `.gooo` extension.

Think of Gooo as a workshop: declarations provide the plan, a small local model
can suggest which permitted parts to assemble, and the compiler checks the fit
and generates Go. A Go experiment runner builds and executes the resulting
program. Each attempt leaves a receipt connecting intent,
choices, generated code, test results, and unresolved questions.

Business intent lives in Gooo declarations. The compiler lowers them to semantic
IR and projects structural Go. Handwritten Go slots hold implementation logic;
the experimental typed-path route can assemble bounded activity bodies from
conditions, assignments, references, branches, and expressions.

```text
Gooo source + intent + permitted choices + finite expectations
                         │
       optional local model → proposed path → finite tests
                                  ↑              │
                                  └── feedback ──┘
                                         │
                          selected source → Go → execution
                                         │
                         source-bound completeness receipts
```

**Start here:** [direction and current progress, 한국어](docs/language-direction.ko.md)
· [body generation](docs/language/body-codegen.md)
· [Gooo source assembly](docs/source-assembly.md)
· [compose activity bodies](docs/native-body-composition.md)
· [small model integration](docs/three-choice-path-model.md)
· [completeness observations](docs/declared-completeness-receipt.md).

## First run

The checked-in example declares the baseline body, two legal alternatives, five
finite input/output checks and an attempt limit in Gooo. From the repository root,
run it without a model:

```sh
go run ./cmd/gooo body-codegen --json --activity Qualified \
  examples/body-codegen/source-assembly.gooo.fixture
```

The result contains the selected Gooo source in `gooo_source`, generated Go in
`source`, and a report with the route, finite checks, completeness and unresolved
items. Repeating the command with the same source gives the same selection.

To use a compatible local model, add `--path-model /path/to/model.json`. The model
only ranks choices already declared in `assembling`; it does not invent a new
plan or bypass type checking and finite evaluation. When the model is omitted,
the same declared search uses deterministic ordering. The
[worked example](docs/source-assembly.md) explains the Gooo syntax, limits and
model profile. For a complete multi-activity run, including compilation and
execution, see [native body composition](docs/native-body-composition.md).

## What we are developing

The goal is to make program construction an inspectable language operation.
Stable semantic IDs connect a declaration to its generated structure and observed
behavior. Small models guide finite choices inside those boundaries. Omitting an
explicit model path selects the declared deterministic route.

The current integration includes an independently trained **4,096-parameter,
16 KiB whole-candidate judge** that compares eight complete body paths, including
two-operation order. Earlier experiments use a **2,072-parameter shared judge**
for three binary decisions and actual failed-case feedback on later attempts.
Both receive source-derived context, Korean/English intent and candidate
structure. Gooo owns the alternatives and the budget.
Broader natural-language discovery and reusable learned abstractions are the next
language research questions.

Activities can declare `assembling` alongside `computes`: intent, permitted source
paths, finite input/output cases and the search budget live together in Gooo.
`body-codegen`, context export, file execution and retained workers read the same
contract. See [source assembly](docs/source-assembly.md) for model/deterministic use.
Generation also returns the selected Gooo source. `body-realize` replays a saved
selection into a fresh directory, keeping its planning baseline and picked paths
beside the working body so the next generation can use the same alternatives.
`body-compose` connects several checked bodies through source-declared `bind`
edges. One optional model is retained across integer assemblies; Integer,
Boolean and Text activities then execute in a compiled graph with ordered
intermediate input/output observations and finite expectation counts.
Ordinary bodies support up to 16 ordered scalar inputs, including repeated types
and partially bound joins. [Input-port example and runnable guide](docs/native-body-composition.md).
String and Boolean records can be constructed, read and passed through
composition binds as native value structs. Boolean fields become Go `bool` and
stay typed across inputs, outputs and field observations. [Record body examples](docs/native-record-values.md).
Record constructor fields can also declare alternative value expressions and
typed JSON cases. Their assembly reports matching fields as well as complete
cases, retaining partial results under a small attempt budget.
[Field assembly example](docs/record-field-assembly.md).
Local records support [sequential field writes and saved values](docs/record-field-updates.md).
The bounded search [continues after a combination fails type checking](docs/record-candidate-continuation.md).
An optional [source value-flow export](docs/record-value-flow.md) follows those
definitions through copies, branches and early returns. A separate small-model
input version summarizes the relation in fixed arrays; training that version
is the next measured step.

We want the language to carry the assembly plan and its unfinished obligations
along with the program. A useful improvement should complete more of the declared
behavior, spend fewer attempts or less memory, and keep the reason for each
change inspectable. Those are the comparisons we are building toward across
languages; the measurements below compare Gooo's own experimental variants.

| Part | Public home | What to find there |
| --- | --- | --- |
| Language and compiler | This repository | Gooo source, semantic IR, projection, execution and receipts |
| Local inference | [gooo-decision-runtime](https://github.com/kimjooyoon/gooo-decision-runtime) | Go model loaders, fixed workspaces, typed search and feedback |
| Research | [gooo-neural-decision-experiments](https://github.com/kimjooyoon/gooo-neural-decision-experiments) | Training, comparisons, raw evidence and reproduction tools |
| Model weights | [Whole-candidate judge](https://huggingface.co/asketeddy/gooo-order-judge-tiny-v1) · [shared judge](https://huggingface.co/asketeddy/gooo-shared-judgment-tiny-v1) | Current 16 KiB model and earlier FP32/ternary studies, model cards and evidence bundles |

## Starting and reading a body run — 2026-10-04

Inspect the running executable with `gooo version --build`; add `--json` for
`gooo/build-identity/v1`. It shows the actual build Go, compiler module, SDK
dependency/replacement and source-binding state. Dirty or unbound builds retain
their original metadata and `UNBOUND_LOCAL_SOURCE`. Native requests check the
actual selected Go executable independently. The default looks for native
Go1.27.1 on PATH, then in the compiler's local GOROOT and exact toolchain cache. `--go-bin` takes priority
for an explicit tool. [Setup and output fields](docs/native-body-worker.md#check-the-executable-and-run-with-a-local-native-tool).

Use `gooo body-path-run` for source, recipe and finite-case files; use
`gooo body-path-stream` for repeated NDJSON requests. Omitting `--model` selects
the deterministic route. An explicit model path is validated before construction.
The [native worker guide](docs/native-body-worker.md) gives executable examples,
file limits, compatible model profiles and result fields.

| What happened | Where to start |
| --- | --- |
| Input/model setup error before an output directory | Check regular-file kind, readable path, size, source UTF-8 and model metadata/weights |
| Go-tool setup error after generation | Check the Go 1.27.1 executable; the generated program and original error remain available |
| Finite cases are unobserved | Read the setup/execution cause before interpreting a completion fraction |
| Executed cases have a mismatch | Compare retained input, expected/actual output and replay result |

Unix input readers open without waiting for a FIFO writer and validate the opened
file itself. This includes source/recipe/case/options, structural model metadata
and whole-candidate judge weights; these readers accept regular-file symlinks.
SDK v0.2.21 also applies bounded nonblocking/no-follow reads to operation/path
and joint/shared model metadata and weights. Those profiles retain their existing
non-symlink rule and check descriptor identity and extent before reading.
Windows arm64 evidence for the SDK reader covers compilation.

The source tested and locally installed for this observation is main
[`d1bfd273`](https://github.com/kimjooyoon/meta-ontology-go/commit/d1bfd273ab4e21d0191548b066a27bcb77d7ed86).
Dev #1202 and main #1203 each passed their own six checks and independently
verified proofs before normal merges. CLI and standalone worker share that
source, Go 1.27.1 and SDK v0.2.20. This is a reproducible observation pin.

Fresh installed controls retained **1,024/1,024** finite cases in eight ordinary
KO/EN model/deterministic constructions and **512/512** in four standalone-worker
constructions with two concurrent requests per mode. All generated Go matched
the frozen programs. A separate 64-swap input study had no writer releases or
timeouts: 33 regular completions retained 4,224/4,224; 31 errors occurred before
construction. These repeat existing authored tasks with unchanged weights/fit.
[Original failures, CPU scope, source proofs and installed records](https://github.com/kimjooyoon/gooo-neural-decision-experiments/tree/main/publication/file-input-open-20261004).

Normal model-request first/reused responses were 298.269/25.206417ms for Korean
and 484.009167/27.778291ms for English, including construction and execution.
Standalone-worker CPU accounting includes setup, native children and EOF joining;
whole-host utilization and model-only RAM/latency remain unobserved.

## Dated model studies — 2026-10-03

The latest model study and the deployed integration have separate evidence:

- **Full-input model research:** four freshly initialized shared judges completed
  6,400 local GPU updates. On 512 previously observed development views, the
  original FP32 control completed all 16 supplied expectations in **113/512**
  first choices; the whole-text fragment model completed **368/512 (71.88%)**.
  Extra ranked attempts fell **1,469 → 186**. Both Korean/English views chose
  valid paths in **180/256** pairs, while **68/256** pairs chose the same wrong
  path. All twelve FP32/PTQ/QAT exports are public.
  [Full comparison](https://github.com/kimjooyoon/gooo-neural-decision-experiments/blob/main/docs/full-input-judgment-initial-results-20261003.md).
- **Compact execution:** 96 generations made 298 local predictions and led to
  192 compiled runs. All **2,304 supplied finite expectations** passed; all
  **48 expanded/compact pairs** retained the same unseeded generated behavior.
  Compact prediction medians were **23.7–30.2 µs** and fresh-process codegen
  medians **27.8–29.5 ms** on local arm64 with Go 1.27.1. QAT's codegen median
  rose slightly. [Native study and resource scope](https://github.com/kimjooyoon/gooo-neural-decision-experiments/blob/main/docs/compact-shared-native-results-20261003.md).
- **New full-input execution:** compiler `e461c1d` with SDK v0.2.15 completed
  **400 generations, 816 predictions and 800 compiled runs**. All **9,600 finite
  expectations** and **192 expanded/compact pairs** matched. Bag-original compact
  FP32 had an 8.33 µs prediction median and 10.00 ms whole-codegen median;
  disconnected codegen was 9.80 ms. The sixteen known tasks show fewer candidate
  attempts and similar process latency. [Complete observation and resource scope](https://github.com/kimjooyoon/gooo-neural-decision-experiments/blob/main/docs/full-input-native-results-20261003.md).
- **What the new study exposed:** counting text fragments can lose operation
  order; wording augmentation and ternary conversion also produced regressions.
  On arm64 and Linux, all 18,432 first choices agreed, while **272 complete
  candidate rankings differed** under small numerical changes. The exact
  comparison failed on partial-completion curves. A subsequent versioned rounding
  rule now reproduces every intermediate value and full ranking on those 18,432
  pairs, with unchanged weights and first-path completeness. SDK v0.2.15 now
  carries the rules used in the new native study; remaining model evaluations
  and broader task coverage are next.
  [Paired arithmetic results and retained failure](https://github.com/kimjooyoon/gooo-neural-decision-experiments/blob/main/docs/full-input-separate-arithmetic-results-20261003.md).

This compiler revision uses **Go SDK v0.2.21-experimental**, with V3/V4 feature
contracts, versioned arithmetic, bounded probe sessions and prepared candidate reuse. The
[observation loop](docs/path-observation-loop.md) can choose a distinguishing
input, obtain its result from a declared Gooo reference activity, and use that
additional case during body construction. Verified probe outputs can be reused,
and an explicit option directly selects a unique surviving candidate. A
[small source recipe](docs/source-path-recipes.md) derives the typed base from
Gooo and names the few structural choices to explore, including bodies that
leave their declared input unread. These operations require
no additional model training.
Each dated native study above identifies its
own model and source revision. The
[model guide](docs/three-choice-path-model.md) identifies the compatible bundle.
The earlier [phrasing diagnosis](https://github.com/kimjooyoon/gooo-neural-decision-experiments/blob/main/docs/bilingual-wrapper-audit-results-20261003.md)
explains why the new study keeps the entire instruction while changing its
representation and training wording.

### Which public component should I use?

| Component | Available now | Current development step |
| --- | --- | --- |
| Compiler at this revision | SDK v0.2.21; source recipes, bounded observation/reuse, prepared candidates, file/stream construction and execution | Usability, finite completion and construction cost |
| [SDK v0.2.21](https://github.com/kimjooyoon/gooo-decision-runtime/releases/tag/v0.2.21-experimental) | V3/V4 inference, explicit arithmetic, owned probes, reusable candidates and bounded Unix model-file loading | Broader source and behavior coverage |
| [Hugging Face model](https://huggingface.co/asketeddy/gooo-shared-judgment-tiny-v1) | Original models, twelve full-input exports and dated evidence | Wording, operation order and new-task evaluation |

At SDK revision `59c8d34`, local arm64 and
[Linux CI](https://github.com/kimjooyoon/gooo-decision-runtime/actions/runs/37070241916)
each replayed 18,432 frozen inputs with 36,864 actual predictions. All intermediate
values and complete rankings matched the explicit-arithmetic observations.
This checks the transfer from research code into the library. The native study
above then observed generated Go and execution using the
[registered integration protocol](https://github.com/kimjooyoon/gooo-neural-decision-experiments/blob/daecfea3583614e006c960de263448a1645a9190/docs/full-input-sdk-native-protocol-20261003.md).

For a concrete walk through intent, finite choices, tests and remaining work,
see [the Korean guide](docs/language-direction.ko.md#한-번의-조립을-따라가-보면).

FP32 weights occupy **8,288 bytes**. Ternary weights occupy **446 bytes** on disk
and decode to **2,096 bytes** of tensors plus eight scale bytes; caller scratch
uses **3,200 bytes**. Whole-process RAM includes additional compiler/runtime
work. These experiments use authored finite tasks; broader workflow coverage and
bilingual meaning preservation need further measurement.

We measure completeness as several obligations: declared structure, generation,
finite behavior, reverse links, execution boundaries, and remaining unknowns.
For example, “368/512 complete” counts functions meeting every supplied example
on their first path. Partial example coverage, later search, and unresolved
execution obligations have their own counts, so one percentage stays tied to
the question it answers.
`gooo completeness-delta` compares two observations while retaining changed
scope, removed obligations, failures, and the first unresolved claim.
[Comparison contract and example](docs/completeness-delta.md).

## Ideas we build on

[SKETCH](https://people.csail.mit.edu/asolar/papers/asplos06-final.pdf) provides a
useful precedent for completing a partial program under a specification.
[DeepCoder](https://arxiv.org/abs/1611.01989) shows how learned program properties
can guide a search; this helps frame our small model's role in choosing attempts.
[DreamCoder](https://arxiv.org/abs/2006.08381) connects program search, neural
guidance and the growth of reusable abstractions.
[Laya](https://huggingface.co/convaiinnovations/laya) helped frame our early
experiments around structured decisions; the current tiny models start from our
own initialization and Gooo training data.
[BitNet b1.58](https://arxiv.org/abs/2402.17764) motivates studying ternary weights,
with storage, runtime memory and quality measured separately.
[W3C PROV-O](https://www.w3.org/TR/prov-o/) supplies vocabulary for tracing the
entities, activities and agents behind an artifact. Our combination of these
ideas is developed through the public experiments linked above.

<!-- PUBLIC-TRUST-BADGES:BEGIN -->
### Public trust surface

These badges are generated from the lowered public-trust `.gooo` policy. Workflow badges report workflow results; they do not claim branch-protection or ruleset enforcement.

#### Language / Release

[![Go 1.27.1 toolchain](https://img.shields.io/badge/Go-1.27.1-00ADD8?logo=go&logoColor=white)](https://github.com/kimjooyoon/meta-ontology-go/blob/dev/go.mod)
[![Published release v0.6.0-dev](https://img.shields.io/github/v/release/kimjooyoon/meta-ontology-go?include_prereleases&label=published%20release)](https://github.com/kimjooyoon/meta-ontology-go/releases/tag/v0.6.0-dev)

#### Build / Conformance

[![CI workflow result](https://github.com/kimjooyoon/meta-ontology-go/actions/workflows/ci.yml/badge.svg?branch=dev)](https://github.com/kimjooyoon/meta-ontology-go/actions/workflows/ci.yml)
[![Compiler compatibility evidence](https://github.com/kimjooyoon/meta-ontology-go/actions/workflows/self-improvement-compiler-compatibility.yml/badge.svg?branch=dev)](https://github.com/kimjooyoon/meta-ontology-go/actions/workflows/self-improvement-compiler-compatibility.yml)
[![Experimental release readiness](https://github.com/kimjooyoon/meta-ontology-go/actions/workflows/gooo-release-readiness.yml/badge.svg?branch=dev)](https://github.com/kimjooyoon/meta-ontology-go/actions/workflows/gooo-release-readiness.yml)

#### Security / Supply Chain

[![CodeQL code scanning configured](https://img.shields.io/badge/CodeQL-code%20scanning-2ea44f?logo=github)](https://github.com/kimjooyoon/meta-ontology-go/security/code-scanning)
[![Dependabot weekly updates](https://img.shields.io/badge/Dependabot-weekly%20updates-0366d6?logo=dependabot)](https://github.com/kimjooyoon/meta-ontology-go/blob/dev/.github/dependabot.yml)
[![Dependency review on pull requests](https://github.com/kimjooyoon/meta-ontology-go/actions/workflows/dependency-review.yml/badge.svg?branch=dev)](https://github.com/kimjooyoon/meta-ontology-go/actions/workflows/dependency-review.yml)
[![Private vulnerability reporting enabled](https://img.shields.io/badge/Private%20vulnerability%20reporting-enabled-2ea44f?logo=github)](https://github.com/kimjooyoon/meta-ontology-go/security/advisories/new)

#### Evidence / Project Health

[![Project status: experimental](https://img.shields.io/badge/Project-experimental-orange)](https://github.com/kimjooyoon/meta-ontology-go#project-status)

#### Community

[![MIT licensed](https://img.shields.io/github/license/kimjooyoon/meta-ontology-go)](https://github.com/kimjooyoon/meta-ontology-go/blob/dev/LICENSE)

The complete row ledger, including unavailable and refuted claims, is emitted by the `Public trust surface` workflow.
<!-- PUBLIC-TRUST-BADGES:END -->

The core language covers packages, namespaces, stable entity IDs and activities.
Experimental profiles add entity fields, typed body construction and Gooo-owned
receipt structures. [The language sketch](docs/spec.md) describes the core;
[body generation](docs/language/body-codegen.md) and
[receipt projection](docs/declared-completeness-receipt.md) describe those profiles.

For source/recipe/case files, `gooo body-path-run --source original.gooo
--activity Name --path-plan recipe.json --cases cases.json --out fresh-directory`
constructs and immediately executes the body, saves generated Go and finite
results, and supports `--repeat 1..16` with one retained native artifact. Omit
`--model` for deterministic construction. Add `--timing` for bounded phase costs
and saved-file bindings; `body-path-run --verify-timing --out directory` checks
those records without loading a model. See the [file and stream usage](docs/native-body-worker.md).

For repeated body construction, `gooo body-path-stream` accepts one JSON request
per line and emits each result when ready. Its optional local model stays loaded
between requests; omitting `--model` selects deterministic construction. Start
with the [runnable recipe example](docs/native-body-worker.md).
Add `--execute` and per-request `execution_cases` to observe the generated body
immediately. A repeated matching body reuses one owned executable while each
request checks current source and runs current inputs twice. Results retain
finite failures and distinguish actual new builds from reused build evidence.

## Earlier design illustrations

The following illustrations explain design ideas. Current implemented behavior
and measured results are described in the sections above and their linked studies.

### The deterministic pressure loop

[![Animated explanation of the semantic self-improvement loop: policy-defined base metrics rise from an observed floor; a deterministic selector chooses a focus subset while every baseline remains guarded; 100 heuristic attempts fan out; single-pressure regressions are rejected; source-backed evidence and path proof let deterministic CI requalify all dimensions; the verified ceiling ratchets into epoch 2's immutable floor](docs/assets/metric-pressure-loop/metric-pressure-loop.gif)](docs/assets/metric-pressure-loop/metric-pressure-loop.png)

The [static PNG preview](docs/assets/metric-pressure-loop/metric-pressure-loop.png)
is useful in viewers that do not animate GIFs. The loop is illustrative: each
system's protected policy/SPI declares its own `N` base metrics, `M` cross
pressures, and active `K`; the language guarantees at least **two independent,
non-compensating pressure dimensions**, not one universal set of numbers. The
animation uses `N=6`, `M=4`, and `K=2` only as one concrete policy instance, and
the 100 parallel agents are an illustrative workload. Every agent focuses only
on the selected subset while all `N` baseline metrics remain non-regression
floors. A performance gain that damages completeness, or the reverse, is
rejected. Attempts may use local inference and may PASS, FAIL, or remain
UNKNOWN, but missing evaluator/oracle/evidence is fail-closed: **Deterministic
CI — not inference** checks exact source-backed evidence across the full
declared vector. Proof computes remaining viable paths, and only a qualified
ceiling becomes the next immutable floor with the next metric/SPI and
provenance obligations. Self-improvement here means verified contract,
evaluator, and evidence gains compound; agents do not rewrite the judge or
lower thresholds.

## The ontology in motion

These ten deterministic GIFs are one visual denominator, generated by the shared
renderer in [`docs/assets/ontology-visuals`](docs/assets/ontology-visuals) and
bound to the source concepts listed in [`visual-manifest.json`](docs/assets/ontology-visuals/visual-manifest.json).
They use labels and shapes as well as color, so CLOSED, UNKNOWN, and REFUTED are
not color-only states. The exact checked-in asset count, byte total, and SHA-256
digests are recorded in [`generated-asset-lock.json`](docs/assets/ontology-visuals/generated-asset-lock.json).

<table>
<tr>
<td width="50%"><img src="docs/assets/ontology-visuals/01-intent-ir-lowering.gif" alt="Animated engineering story showing a billing .gooo activity becoming inspectable semantic IR, generated Go source, and a PASS receipt." width="460"><br><strong>1. Source to generated Go</strong><br>A billing declaration becomes generated Go plus a receipt.</td>
<td width="50%"><img src="docs/assets/ontology-visuals/02-authority-boundary.gif" alt="Animated engineering story showing a declaration becoming typed Order, PayOrder, and Receipt semantic nodes with used and wasGeneratedBy edges consumed by a backend." width="460"><br><strong>2. Semantic IR as graph</strong><br>A typed graph is consumed by the backend.</td>
</tr>
<tr>
<td><img src="docs/assets/ontology-visuals/03-munchausen-proof-choice.gif" alt="Animated engineering story showing activity.gooo and entities.gooo reorder into canonical filename order, merge into a billing package API, and emit a PayOrder receipt." width="460"><br><strong>3. Multifile package resolution</strong><br>Unordered source files become a deterministic package API.</td>
<td><img src="docs/assets/ontology-visuals/04-claim-evidence-lifecycle.gif" alt="Animated three-lane engineering handoff where an author writes main.gooo, a compiler creates generated.go, and a reviewer consumes receipt.json across a no-cross-write boundary." width="460"><br><strong>4. Agent handoff by receipt</strong><br>Author, compiler, and reviewer agents exchange caller-owned artifacts.</td>
</tr>
<tr>
<td><img src="docs/assets/ontology-visuals/05-unknown-cause-descent.gif" alt="Animated engineering story showing a missing artifact create stage, step, reason, class, next_operation, and blocked_by fields before evidence re-evaluates the same claim as CLOSED." width="460"><br><strong>5. UNKNOWN to CLOSED resolution</strong><br>Six fields guide a resolver to a new evidence-backed decision.</td>
<td><img src="docs/assets/ontology-visuals/06-precedence-counterexample.gif" alt="Animated engineering story where a counterexample for the same claim enters a precedence stack, selects REFUTED over UNKNOWN, and appends the preserved record." width="460"><br><strong>6. Refutation precedence</strong><br>A known contradiction wins and remains in the ledger.</td>
</tr>
<tr>
<td><img src="docs/assets/ontology-visuals/07-package-resolution.gif" alt="Animated engineering story showing pinned source, contract, and toolchain digests produce identical output twice, then a changed byte creates a mismatch and blocks adoption." width="460"><br><strong>7. Deterministic replay block</strong><br>Byte drift blocks adoption after replay comparison.</td>
<td><img src="docs/assets/ontology-visuals/08-incremental-conformance.gif" alt="Animated engineering story showing a changed subject with six-digest identity routed to one REUSE receipt, with EXECUTE, UNKNOWN, and REFUTED kept as inactive alternatives." width="460"><br><strong>8. Incremental conformance router</strong><br>One identity selects one reused receipt.</td>
</tr>
<tr>
<td><img src="docs/assets/ontology-visuals/09-bootstrap-oracle.gif" alt="Animated engineering story moving from observed metric bug to meta-rule change, exact-head CI, dev adoption, post-adoption receipt, and main eligibility in causal lanes." width="460"><br><strong>9. Self-improvement gate cascade</strong><br>Each receipt unlocks the next engineering gate.</td>
<td><img src="docs/assets/ontology-visuals/10-promotion-lineage.gif" alt="Animated experimental domain projection showing OpenAPI-style service facts and OpenTofu plan facts connecting through a proposed Gooo contract to an UNKNOWN mismatch dossier without infrastructure mutation." width="460"><br><strong>10. Experimental domain projection</strong><br>Proposed cross-domain projection remains UNKNOWN until implemented evidence exists.</td>
</tr>
</table>

Regenerate the complete set with `go run ./docs/assets/ontology-visuals`.

Regenerate and verify the checked-in media with:

```sh
go run ./docs/assets/metric-pressure-loop
go run ./docs/assets/metric-pressure-loop -check
```

## Quick start

Run the repository checks from the project root:

```sh
gofmt -l .
go test ./...
go vet ./...
go run ./cmd/gooo check examples/billing/main.gooo
```

The conformance walkthrough in [docs/conformance.md](docs/conformance.md) uses
the checked-in examples and shows the expected command shapes. It also calls out
which CLI surfaces are not yet supported. The current GitHub Actions workflow is
documented in [CONTRIBUTING.md](CONTRIBUTING.md); it should be treated as the
source of truth for required CI, not as a promise of future compiler features.

## Branch and promotion contract

Work branches target `dev`. Promotion uses a same-repository pull request to
`main`. CI accepts either the exact `dev` head when it is a fast-forward, or a
single-commit snapshot branch named `agent/main-promotion-snapshot-<dev-sha>`
whose tree equals the live `dev` tree and whose sole parent is the live `main`
head. The snapshot form preserves the current tree and main's linear history
when old branch ancestry has diverged. Governance is `ci_only`: review and
approval fields do not authorize a protected-branch promotion.

The six canonical proof jobs are `gofmt`, `go vet`, `go test`, `go test -race`,
`Semantic conformance`, and `CI policy`. GitHub's `main` protection rule now
requires exactly these six contexts; the retired `CI guardian` context was
removed. The `dev` rule has no required status checks. No badge above turns a
workflow result into an enforcement claim.

The repository's default branch is `dev`; checked changes are also promoted to
protected `main` through this process. The separate protection update removed
the retired Guardian check.

For either promotion form, CI emits a digest-bound `promotion_authorization`
with `source=dev`, `target=main`, and `operation=fast_forward`. It passes only
for fresh exact refs and topology (`ahead > 0`, `behind = 0`, `main` as merge
base), the six canonical proof jobs, and a clean, open, non-draft, unmerged
same-repository pull request. Snapshot evidence also binds the live `dev` SHA
and tree, candidate tree, and candidate's sole `main` parent. The proof producer
never mutates refs or protection. After a final exact reread, only a normal
fast-forward update is allowed; force updates are prohibited.

## Project status

The language and its model interfaces are experimental. Runnable commands and
their linked evidence define current support. Production editor integration,
broader analysis, automatic reconciliation of ambiguous observations and durable
provenance publishing remain development areas. Historical protocols retain the
scope and outcomes of the revision they measured.

## Governance

[AGENTS.md](AGENTS.md) defines authority boundaries and agent roles.
[CONTRIBUTING.md](CONTRIBUTING.md) defines branch, PR, review, and CI workflow.
[docs/governance.md](docs/governance.md) records the SSOT boundary, BX laws, line
caps, and evidence policy. [docs/metrics-rfc.md](docs/metrics-rfc.md) defines
the design-only deterministic metric contract. [docs/conformance.md](docs/conformance.md)
is the runnable example index. The [Deterministic CI Evolution Retrospective](docs/deterministic-ci-evolution.md)
is the append-only read-only evidence record.

An experimental source-to-Go activity-body projection is documented at
[docs/language/body-codegen.md](docs/language/body-codegen.md). It keeps body
generation separate from the stable package projection until its syntax and
completeness claims have broader evidence.

The completeness receipt's shared structure is declared in Gooo and generates
the compiler's Go types and JSON Schema. See
[declared completeness receipt](docs/declared-completeness-receipt.md) for the
bounded projection profile, source binding and independent consumption checks.

[W3C PROV-O]: https://www.w3.org/TR/prov-o/
