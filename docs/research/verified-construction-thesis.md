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

## Research context

Gooo builds on questions studied in related systems, and should be compared to
them carefully rather than described as having invented their individual
techniques:

- [Rosette](https://github.com/emina/rosette) embeds solver-aided programming
  and synthesis/verification workflows in a host language.
- [DreamCoder](https://arxiv.org/abs/2006.08381) combines a growing program
  language with learned guidance for program search.
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
