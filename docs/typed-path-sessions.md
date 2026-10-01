# Incremental typed paths

The optional local model judges bounded Korean or English intentions for Gooo
structural alternatives. A session ranks those alternatives once, then evaluates
new candidate bodies in batches without repeating earlier masks. It does not
train, rewrite source files or call an external provider. Optional explicit
feedback can rerank remaining paths from observed failures. Changing the plan,
cases, seed or model requires a new session.

```sh
gooo body-codegen --json --path-plan plan.json --path-model model.json \
  --path-step-attempts 8 --activity ConditionalAssign source.gooo
```

Omit `--path-model` for deterministic declared-fallback ordering. Omit
`--path-step-attempts` for the existing search. The document retains its total
1..64 candidate budget and 1..128 finite cases. Each step evaluates 1..64 new
candidates. Model ranking happens after binding the fallback to authoritative
Gooo source and before candidate tests. The compiler emits and verifies the
final selected body; the JSON receipt includes initialization and batch progress.
This CLI returns one final JSON result, not streaming native Go emissions.

The receipt separates:

- Typed lowering completeness from finite functional completeness. A body can
  lower completely while satisfying only six of seven declared cases.
- Cumulative evaluated and type-rejected candidates from unattempted masks.
- Initial model calls from zero new model predictions during each advance.
- The eight-byte scheduled-mask bitset for 64 alternatives from frontier,
  prepared body, model, receipt and whole-process memory.

Progress hashes link observations; they do not authorize edits or establish
ontology facts. Source binding, stable activity identity, deterministic replay,
native type checking and parity with the typed interpreter remain required.
Partial outcomes are retained with their explicit finite denominator.

The Go SDK `v0.2.5-experimental` exposes `NewSession`, `Observe`, `Advance`,
`SearchBatches`, `Reconsider` and `SearchFeedbackBatches`. Its core session keeps one best body, a frontier and a bitset,
without previous attempt logs. The native compatibility helper retains up to
64 attempts and their batch observations for its final receipt. The standalone
research CLI can stream these observations across the larger declared finite
space; native documents continue to cap total attempts at 64.

Concurrent calls on the same SDK session fail immediately with `ErrSessionBusy`.
Candidate evaluation checks context cancellation and requeues an unfinished
candidate without recording a partial case prefix. Generic output writers can
still block outside the evaluator: process harnesses must bound and drain
output and enforce a process deadline. This is not a guarantee of all-input
correctness, general language accuracy or arbitrary I/O deadlock freedom.

## Optional observed-feedback judgment

```sh
gooo body-codegen --json --path-plan plan.json --path-model model.json \
  --path-step-attempts 8 --path-feedback-rounds 2 \
  --path-feedback-ci hint.json --activity ConditionalAssign source.gooo
```

The local model and an explicit batch size are required for feedback. The total
candidate budget stays 1..64; 1..16 rounds can reconsider only after new partial
candidate batches with remaining paths. It neither repeats evaluated masks nor
changes the original intentions, test expectations or selected best body.
The same frozen model receives bounded original Korean/English intent plus the
observed attempted count, finite pass count, first mismatch and optional CI status.
Model inputs must fit 512 bytes; the original intention is never truncated.
If failure context makes a valid original intention exceed that bound, the
receipt records `context_declined`, input/intent hashes and the attempted byte
count. This consumes a bounded feedback round with zero new predictions and
leaves the remaining ranking unchanged. Candidate construction continues and
can emit its verified best partial body. Same-batch retries are rejected.
Cancellation, invalid model/ABI, source-binding errors and other prediction
failures still stop generation; this exception applies only to representation
declines detected before any feedback prediction.

`hint.json` is optional and has exactly `source_sha` (40 lowercase hexadecimal
characters) and `status` (`PASS`, `FAIL` or `UNKNOWN`). The file is bounded to
512 bytes. This caller claim is recorded as context; this API does not fetch or
authenticate CI and the hint never grants authority to modify ontology or merge.

`feedback_judgments` records input hashes, prediction counts, model hashes and
links to preceding progress and feedback. Extra progress observations immediately
after each reconsideration include cumulative calls, including interrupted calls.
`predictions_this_advance` stays zero: candidate evaluation itself does not infer.
The same 8-second native deadline bounds binding, loading, search and emission.
Interrupted searches retain receipts but do not emit an unverified Go body.

If exactly one declared candidate remains, reconsideration cannot change the
next path. A hashed `ranking_unnecessary` receipt retains the original failure,
model, plan, cases, progress and caller CI bindings with zero new predictions.
The bounded round is consumed and same-batch retries are rejected. The remaining
candidate still passes normal type and finite-case evaluation. Deadline, model
and context validation remain enforced. This saves a redundant invocation; it
does not establish functional completion or a measured wall-time improvement.

Omitting the feedback flag preserves rank-once batching. Omitting the model and
feedback preserves deterministic fallback ordering. Callers choose their model;
receipt wording does not infer its training history. The earlier frozen-model pilot added 102
feedback predictions with identical final finite completeness across 16 pairs;
this mechanism is an experiment in continued construction, not a demonstrated
accuracy gain. Those results precede this native SDK integration; new native
measurements must be recorded separately.
