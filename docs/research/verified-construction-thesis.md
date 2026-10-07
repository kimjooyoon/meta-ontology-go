# Bounded, verifiable construction: thesis and evaluation plan

This is a research plan for Gooo, not a product or performance claim. The goal
is to find out whether explicit construction choices, a replaceable ranking
model, compiler checks, and source-bound evidence can make generated changes
easier to reproduce and verify.

## The problem to measure

Code generation can produce candidates faster than a team can establish what
they change and whether those changes satisfy a task. Gooo explores making that
construction process explicit: a `.gooo` source describes intent and permitted
choices; an optional model may rank those choices; the compiler lowers and
checks the selected path; finite checks and later execution evidence record
what was observed.

The practical question is whether this reduces the resources needed to obtain a
verified result. “Verified” must name the exact checks that ran. Passing a
finite selection suite is not a proof of general correctness, and an unknown
measurement must not be counted as success.

## Responsibility boundaries

| Component | Responsibility | Current limit |
| --- | --- | --- |
| Gooo source and ontology | Declare stable concepts, intent, and permitted construction | Coverage depends on the declared scope; declarations can be incomplete |
| Compiler | Parse, type-check, constrain and lower the declared construction | A successful lowering alone does not establish runtime behavior |
| Decision model | Rank choices already present in the declared search space | Ranking can still select a poor choice; finite checks cover only named cases |
| Verifier and execution | Accept or reject candidates against explicit obligations | Results apply only to the checks, inputs, and environment recorded |
| Provenance and receipts | Bind inputs, choices, generated artifacts, observations, and unresolved claims | A receipt preserves evidence; it does not make weak evidence strong |

The intended boundary is that a model can change the search order without
adding undeclared choices or bypassing compiler checks. Broader claims that
model replacement preserves program behavior require stronger equivalence
obligations than the current finite suites provide.

## What exists in this repository

- Gooo source can declare bounded alternatives and typed body-construction
  paths, with deterministic selection when a model provider is absent.
- A local compact model can rank declared choices in supported routes. It does
  not author arbitrary source outside those choices.
- Generated bodies are checked against the declared profile; supported flows
  record finite checks and source-bound receipts.
- The domain-completeness profile declares its measurement roster, per-axis
  classifier, and aggregate outcome priority. Go adapters still gather the
  evidence and attach the reason and next frontier.
- The receipt model distinguishes `PASS`, `PROGRESS`, `UNKNOWN`, and
  `FAIL_CLOSED`; it keeps unresolved evidence visible rather than folding it
  into a single success score.

This is bounded construction and measurement plumbing. The repository does not
yet demonstrate a general multi-agent development platform, broad natural
language programming, or a comparative advantage in cost or correctness.

## Measure cost to a verified solution

Keep the result as a vector until the units and tradeoffs are justified. For a
frozen task set, report:

1. **Verified completion rate:** tasks reaching the predeclared verification
   boundary divided by tasks attempted. State the exact boundary for every
   result.
2. **Time to verified result:** wall-clock time from the same task snapshot to
   the first accepted candidate, with sample count, median, and tail
   percentiles. Small samples should be described as preliminary.
3. **Compute and model cost:** CPU time, peak RSS, sampled CPU utilization with
   host and process scope, model calls, model latency, input/output tokens where
   available, and provider cost only when measured from an authoritative
   source. Keep child-process measurements separate from whole-run totals.
4. **Search efficiency:** candidates considered, candidates rejected by each
   check, and accepted results per resource unit. Keep “not run” separate from
   “failed”.
5. **Human intervention:** count edits, approvals, and restarts required to
   finish. Unattended progression is the target; report the observed count and
   expose each intervention-required run instead of assuming it is zero because
   a path is called automated.
6. **Reproduction and scope:** replay success on the same inputs, task and
   source identities, finite-case coverage, and every remaining `UNKNOWN`
   dimension.

Do not combine these dimensions into a percentage without a published weighting
rule. “Cost per verified solution” can be a derived measure after cost units,
failure treatment, and verification scope have been fixed. Until then, show the
vector and its denominators.

## A fair first comparison

Start with a small, versioned corpus of tasks that can be replayed locally.
Freeze the task text, source snapshot, candidate set, finite checks, toolchain,
resource budget, and acceptance rules. Run deterministic ordering and local
model ranking against the same candidates and checks. If ranking is stochastic,
record the model digest, seed, and repeated runs. Compare paired tasks rather
than unrelated totals.

For each run, retain the task and source digests, compiler revision, model
identity, selected path, generated artifact digest, check results, resource
observations, and unresolved claims. Publish the corpus and runner with the
results so another person can reproduce them. Add a stronger single-agent or
parallel-agent baseline only when the runner can apply the same task snapshot,
verification boundary, and resource accounting to it.

The first study should answer narrow questions:

- Does ranking reach an accepted candidate with fewer attempts under the same
  budget?
- Does it improve time or compute cost without lowering verified completion?
- How often does it rank a candidate that the compiler or finite checks reject?
- What remains unknown after a successful run?
- Does deterministic fallback produce byte-stable results when no provider is
  configured?

Set decision thresholds before looking at aggregate results. Publish negative
and inconclusive results alongside positive ones.

## Grow through a replaceable construction system

The project direction is to make Gooo useful for constructing small, inspectable
programs and language tools. A first public demonstration should follow one
real task from its Gooo contract through selection, native execution, and an
explanation of the remaining work. Developer-facing language comes first:
what can be expressed, what ran, and how another person can reproduce it.

Treat substitution as two different experiments:

| Change | Keep fixed | Observe |
| --- | --- | --- |
| Deterministic ordering, model A, model B | Source/ontology, permitted candidates, compiler, checks, task inputs and budget | Candidate order, selected body, rejected attempts, finite behavior, resource use and unknowns |
| Domain vocabulary or obligations | The identified compiler/kernel revision | Changed semantic identities, resulting candidate space, required checks, unsupported constructs and migration needs |

With model replacement, the definition of valid construction stays fixed.
The selected program and its behavior may change under a bounded search budget;
record that difference and check it. With domain replacement, explicitly record
the changed meaning. Two small [domain tools](../../examples/domain-tools/README.md)
now share report and rendering packages while declaring different observation
types and rules. Their paired deterministic/model runs retain the same candidates
and checks. This demonstrates substitution within the supported record profiles;
broader ontology substitution needs additional executable examples.

Prioritize three small ecosystem tasks: explain a failed assembly, suggest a
bounded repair from its counterexamples, and generate an executable example
from declared types and obligations. Each should produce a Gooo artifact that
can be inspected and replayed. A useful next language feature is one required by
a recorded failure in those tasks. Model training then learns search preferences
from the retained attempts, with dataset identities and held-out task families.

### Keep the evaluation units visible

`body-compose` now reports `input_separation` after replaying the source and
running the native graph twice. It compares actual input tuples at every
assembling activity with declared selection/holdout inputs and recorded probes.
A root case is disjoint only when every assembling activity receives a disjoint
tuple. Duplicate root inputs contribute one case; all expectations supplied for
that input must agree with execution. See the [runtime measurement
contract](../native-body-composition.md#input-separation).

Workspace execution and replay now include earlier source fills and external
fill plans in this measurement. Their training, holdout and probe tuples join
the known-input sets, with activity, source, plan and input-set identities.
[The whole-workspace example](../../examples/workspace-input-observations/README.md)
includes a new root that becomes an already observed downstream value. It is
counted as overlapping. Earlier published receipts retain the scope recorded by
their original compiler revision.

This measures new inputs relative to the recorded construction observations.
Model-training exposure remains unknown. In the same way, native case success,
grammar coverage and declared-obligation coverage each retain their own units.
Add task-level completion only after freezing the task's acceptance boundary;
add external agent baselines when they share that boundary and accounting.

### Build artifacts that others can carry forward

Keep source, compiler revision, model and dataset identities, applicable license
references, generated artifacts and execution receipts linked. Missing records
stay explicit. Provenance establishes a traceable chain; ownership and license
claims require their own supporting records. This makes the research easier to
reproduce, extend and evaluate independently.

Start outreach material with the runnable example, then the paired comparison,
then the language design. Record external reproduction and useful tasks as they
occur. Product packaging and commercial options can develop from observed use;
no acquisition price, adoption figure or hypothetical benchmark is treated as a
project result.

## Research context

Gooo builds on questions studied in related systems, and should be compared to
them carefully rather than described as having invented their individual
techniques:

- [Rosette](https://github.com/emina/rosette) embeds solver-aided programming
  and synthesis/verification workflows in a host language.
- [DreamCoder](https://arxiv.org/abs/2006.08381) combines a growing program
  language with learned guidance for program search.
- [W3C PROV-O](https://www.w3.org/TR/prov-o/) supplies a vocabulary for
  representing provenance across entities, activities and agents.
- [Halide](https://people.csail.mit.edu/jrk/halide12/) separates algorithm
  specification from scheduling decisions. A later [schedule-search
  study](https://ai.meta.com/research/publications/learning-to-optimize-halide-with-tree-search-and-random-programs/)
  explores tree search with a learned cost model.

Gooo's testable question is about the composition of source-declared semantic
space, replaceable decision ranking, compiler-owned constraints, explicit
verification boundaries, and provenance-rich completeness evidence. Whether
that composition is useful or distinct in practice remains to be demonstrated
by matched experiments and external use.

## Development sequence

### Public growth milestones

Keep each milestone tied to something another developer can inspect or run:

| Milestone | Published artifact | Evidence to retain |
| --- | --- | --- |
| Follow one construction | A short source-to-execution example with its expected cases | Source, selected path, generated body, outputs, and unresolved dimensions |
| Compare decision policies | The same frozen tasks with deterministic and model-ranked ordering | Paired budgets, all attempts, model identity, measured zero results, and unknown observations |
| Try another domain | A versioned ontology and explicit verification obligations | Which declarations change, which compiler rules remain shared, and unsupported constructions |
| Reproduce outside this workspace | Setup instructions and an independent run report | Environment, deviations, failed steps, and any human intervention |
| Learn from external use | A documented task supplied by another developer | The task's scope, observed outcome, and reproducible failure cases |

The first demo can be small enough to follow in a few minutes. Show the same
source with a model and with deterministic ordering, then inspect a failed or
incomplete result as well as a passing one. Treat the model like a suggested
route through a workshop: the available parts and assembly rules remain visible
when the suggestion changes.

Public technical notes should link to the corresponding revision, runnable
example, and result artifacts. Repository counts, parameter counts, and stars
provide context; external reproduction and useful completed tasks establish
whether the construction process helps people. The proposed agent comparisons
and commercial directions are hypotheses to revisit as that evidence grows.

### October 8 checkpoint and next deliverable

The first milestones now have small runnable artifacts. Use these as the entry
point for the next contributor or user:

| Starting point | Evidence already published | Next question |
| --- | --- | --- |
| [Diagnostic starter](../language/project-starters.md#start-a-reusable-diagnostic-tool) | A two-package Gooo tool, optional own-model assembly, and saved-program replay on new inputs | Can another developer adapt its rules and cases to a useful task? |
| [Two domain tools](../../examples/domain-tools/README.md) | Assembly and documentation rules share a report interface; deterministic/model pairs retain their attempts and finite outputs | Which missing expression or reusable operation appears when a third domain is added? |
| [Construction explainer](../../examples/construction-observation/README.md) | Gooo interprets retained construction attempts and returns a next operation | Can an explicit proposal be consumed to construct and check the next candidate? |

The diagnostic starter is available on `dev` from commit
`25df294860e1cfdff8e419bc2cf25390a52043b7`. The separately published
[v0.6.5-dev release](https://github.com/kimjooyoon/meta-ontology-go/releases/tag/v0.6.5-dev)
predates that template. Pin the source revision when reproducing a study.
Maintainer runs, including runs of downloaded binaries and module installations,
establish installation observations; external reproduction and adoption still
need independent reports.

The next construction experiment should connect the existing explainer to a
bounded repair proposal. Give it a retained failed attempt, the remaining
source-declared alternatives, and a fixed attempt budget. Have Gooo describe the
next permitted operation; let the own compact model rank eligible choices where
the current adapter supports them. The Go runner applies an explicit choice,
runs the original obligations, and saves the new attempt. Replay uses the saved
construction without new inference. The [connected record-choice
experiment](../../examples/assembly-policy/README.md) now executes the Gooo policy
between scored candidates. Its checkpoint and continuation route carries earlier
observations into the remaining original ranking with zero new model calls.
This revision makes an explicit follow-up invocation possible; the command caller
still starts that invocation. Other assembly profiles and automatic scheduling
need their own supported routes.

Start with one reproducible failure and include a case where the declared space
has no acceptable solution. Keep that unresolved result visible. Compare the
deterministic and model-ranked paths on the same space and budget. Report matched
cases and fields, newly observed inputs, rejected attempts, model calls, elapsed
time, and human interventions with their denominators. Growing the grammar is a
separate language change and needs a new comparison against the old space.

Use this tool-building work to choose language features. The diagnostic tool's
repeated partial-completion condition motivated [pure activity
calls](../../examples/pure-activity-calls/README.md): a fixed Gooo helper can now
be reused inside a branch, with typed arguments, its own local bindings, bounded
call relationships and native projection. The same mechanism lets a Gooo assembly
policy reuse a continuation predicate. [Workspace calls](../../examples/package-body-calls/README.md)
now resolve local-package functions and imported helpers through the calling
file's aliases. Their records map original package/activity identities to lowered
names, and replay reconstructs the same closure. The next useful step is to apply
these libraries to another real tool and keep the newly exposed gaps. Model training should
follow such concrete construction tasks, retaining the source, attempted
decisions, counterexamples and outcomes.

The [package assembly-policy tool](../../examples/package-assembly-policy/README.md)
now applies that reuse to the construction controller itself. Its Gooo policy
imports a shared continuation predicate, reads actual candidate observations and
controls construction of a different workspace. Policy source packages accompany
the saved result for reconstruction. A paired own-model run and a deliberate early
stop retain both complete and partial finite outcomes. This extends reusable
metaprogramming within the current record-choice profile. [Called-body
construction](../../examples/called-body-construction/README.md) now constructs
source-declared helper bodies in dependency order before projecting their callers.
The attempts remain independently inspectable. Automatic scheduling and retaining
dependent caller constructions after a helper changes remain further work.

Fresh called-body executions now retain the actual arguments delivered to each
constructed helper. The [paired follow-up](called-input-observation-20261008/summary.json)
distinguishes input overlap from finite output success, and replays an older
saved construction with fresh argument observations and zero new inference.
The model reached the selected helper in two attempts, compared with four for
deterministic ordering; the measured whole commands took 0.36 s and 0.34 s
respectively. Candidate count and elapsed time remain separate outcomes. Calls
made while scoring another assembling body still need their own input-history
observations before nested construction can claim disjoint runtime inputs.

Publish a short Korean and English walkthrough of the connected tool, its
failed case and its replay. An independent developer should be able to follow it
without reconstructing the research repository history. Their reproducible
failure or useful task becomes input to the next language iteration. Commercial
packaging can be evaluated from those observed uses; acquisition estimates and
proposed user counts remain outside the technical success criteria.

### Language development principles

The contextual-hole example is a concrete instance of this sequence: a retained
failure on `input + hole` exposed a missing candidate, and the optional
`integer-hole-residual/v1` grammar now derives proposals from the surrounding
typed body. Its [runnable composition](../native-body-composition.md#contextual-hole-construction)
keeps the same source-owned expectations and checks the resulting native outputs.
Report this as a grammar change, not a model-ranking advantage: the candidate
space has changed. Keep the old grammar as a baseline, publish nonlinear failures,
and measure model substitution separately with the grammar fixed.

1. Keep language semantics, candidate permissions, decision ranking, and
   verification obligations separately inspectable in Gooo and its receipts.
2. Expand the reproducible task corpus only when each task has a stable source,
   explicit checks, and a known evidence boundary.
3. Compare deterministic and model-ranked routes under paired budgets; improve
   the compiler or language when failures expose missing expressiveness rather
   than hiding those failures in a model score.
4. Add external baselines and adopters when they can be measured with the same
   verification contract.
5. Make public claims only from versioned artifacts and observed results.

The project is an open research effort. The near-term goal is a trustworthy
language and evidence trail; product and commercial directions remain options,
not promises.
