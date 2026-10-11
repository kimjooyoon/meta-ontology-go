# One native construction using the published relation model

## Fixed question

Can the public source-relative HF weights enter the normal Gooo CLI, rank before
candidate execution, check the source requirements, run the chosen body as Go,
and replay the saved construction without reading a model? This is an integration
observation on one small authored task. It does not estimate general model quality.

## Source and requirements

Freeze and publish the compiler adapter, this protocol and the three new
`examples/body-codegen/source-relative-batch*` files before the original run.
Record the frozen Git commit, source and case file SHA256 values. Use the compiled
CLI from that clean commit and Go1.27.2. The source marks comparison order and
branch layout as two binary choices. Its fallback branch layout is intentionally
inconsistent with the authored output cases.

The task chooses a small batch amount: backlog below8 maps to backlog+3; backlog
at least8 maps to backlog*2. Source outputs use0,3,8,13; Boolean cases use0,8,13.
The outer construction case is9→18. Fixed later evaluation cases are1→4,7→10,
10→20,21→42. These source, outer and later inputs are separate sets except the
intentional source output/condition overlap. The same four later cases are used
for both policies and saved replay. Repeated evaluation is not independent evidence.

## Fixed model

Use the already retained public HF artifact from
`asketeddy/gooo-canonical-decision-tiny-v1`, revision
`be6d95747d3a498620d0a5a6c42b4284c964cba2`, path
`research/source-relative-learning-20261011/result/model-relative.json`.
Its SHA256/fingerprint is
`cae32bdad17ab818498c6099066918cd9c9c591eb118afe1bc6a7ef82dc83cc4`.
This file is also unchanged in the newer documentation revision1b26d624.
Use the local previously verified copy. Do not download, fit, tune, quantize or
select weights. None of the original288 study documents is rerun.

## Original sequence

Use a new output directory and a once-only start marker. Keep the complete stdout,
stderr, exit status and per-process resource observation of each operation.
Perform exactly this sequence, sequentially after local test processes finish:

1. `body-context` on the new source with the fixed model; zero predictions/candidate
   executions, full input arrays and condition relation explanations retained.
2. `body-construct` with no model, fixed source and both case files, four attempts.
3. `body-construct` with the fixed model and identical source/cases/attempts.
4. Replay the model construction's saved original source and construction.json,
   omitting the model, against the same later evaluation file.

The inspection's zero calls, the model session's initial zero candidates/one call,
later advances' zero predictions, retained partial/failure data, generated source,
native output cases and model-free replay calls are separate checks. Preserve
original source/plan/model identities and compare the exported input hashes to
the actual ranking receipt. Count first-candidate success, final authored output
and condition passes, candidate attempts and later native4-case results separately.

If any original command fails, retain its output and stop dependent operations;
do not change the inputs, source, model or thresholds and rerun this observation.
Unit fixture regressions and read-only audits remain repeatable.

## Timing and publication

Record Go/compiler/SDK versions, hardware and `/usr/bin/time -l` process data.
Use `GOMAXPROCS=1`, with default Go build caching. Fixed order, warm caches and
shared hardware make these descriptive timings, not a causal speed comparison.
Compile/execute/process costs are distinct from the model's own ranking interval.
Do not infer an instantaneous whole-host CPU increase from user/system CPU time.

Publish scoped JSON results, input/identity/call-count audit, resource data and
failures. Exclude local filesystem paths and temporary environment logs from the
public bundle. A separate digest manifest identifies published files. The earlier
HF study remains immutable; this is a new native integration observation.
