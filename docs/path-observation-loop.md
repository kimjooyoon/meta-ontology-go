# Asking for a useful observation during body construction

`body-codegen --path-observation` extends a finite typed-path contract using a
separate pure Gooo activity declared in the same source. It works with the local
model or deterministic enumeration. SDK v0.2.16 supplies the probe ranking;
the compiler binds the source, evaluates the declared oracle and generates Go.

Think of fitting two parts that look identical from one side. Before choosing,
the compiler finds an angle from which their behavior differs. The declared
reference activity supplies what should be seen from that angle.

## Run the small example

```sh
go run ./cmd/gooo body-codegen --json --activity Probe \
  --path-plan examples/body-codegen/path-observation-plan.json \
  --path-observation examples/body-codegen/path-observation-inputs.json \
  examples/body-codegen/path-observation.gooo.fixture
```

Both `input-2` and `2-input` satisfy the initial case `2 → 0`. The supplied
inputs `[2,3,0]` contain a distinguishing input: `3`. The source activity
`Expected` computes `2-input`, so its observed result adds `3 → -1`. The surviving
candidate set shrinks from two to one. The generated `Probe` returns `2-input`.
The example uses zero model calls and performs no training.

The observation request is a strict JSON object of at most 4 KiB:

```json
{
  "schema": "gooo/path-observation-request/v1",
  "inputs": [2, 3, 0],
  "max_candidates": 2,
  "max_rounds": 2,
  "oracle_activity": "Expected"
}
```

`oracle_activity` may be omitted. The compiler then records a recommended input
and leaves its expected value unresolved. Generation still uses the original
cases. A present oracle must be a different, pure `Integer -> Integer` activity
in the source. It must match every initial case before any new observation is
accepted. The compiler uses its bounded Go AST evaluator; this step starts no
native process or network request.

## Execution order and model context

1. Prepare the typed plan and bind its fallback to the original source activity.
2. Check the declared oracle against the original finite cases.
3. Enumerate bounded candidates in ascending mask order. Keep candidates that
   pass the current cases and evaluate their outputs on the supplied probe inputs.
4. Recommend the input separating the most unordered candidate pairs; break ties
   by the smallest largest equal-output group, then supplied input order.
5. Evaluate that input with the declared oracle, append the observation, and
   repeat within the round budget. Previously tested inputs are never recommended.
6. Run the existing model-guided or deterministic search against the effective
   cases, then perform Gooo emission, type checking and replay.

`--path-model`, `--path-step-attempts`, `--path-feedback-rounds` and caller CI hints
compose with this route. The initial model context retains the original complete
document and intent. Added observations affect candidate evaluation; later model
feedback can include a resulting failure. Probe ranking itself makes no model
predictions. The loop is synchronous and shares the existing eight-second request
deadline. A retained generator can use `TypedPathOptions.Observation`; each
request owns its cases and observations.

The optional `"reuse_probe_outputs": true` request field retains the first
ranking's candidate outputs in an SDK v0.2.17 probe session. Each added oracle
observation then filters those values. Later observation rounds perform zero
new candidate evaluations or model calls. Each request owns its session, including
when a retained generator serves concurrent stream requests.

The omitted/false setting preserves earlier v1 receipts and their repeated
evaluation accounting. With the default resolution setting, both modes run the
existing candidate search after probing. Total code-generation latency and process
memory need their own paired measurements; the SDK's cached-operation microbenchmark
covers a smaller scope.

## Budgets and what the receipt means

### Project a uniquely observed candidate

The optional `"resolve_unique_candidate": true` field uses a candidate directly
when the complete declared enumeration leaves exactly one survivor. It composes
with either fresh evaluation or `reuse_probe_outputs`. A partial enumeration,
empty survivor set or ambiguous set continues the existing bounded search and
records `PARTIAL_ENUMERATION`, `NO_SURVIVOR` or `AMBIGUOUS` respectively.

For a `RESOLVED` result, `body_paths.resolution` contains the selected mask and
choices, final-ranking and effective-suite hashes, already observed finite score,
and skipped-work flags. `search_started` is false and the search record is empty:
probe observations are not counted as new search attempts. The compiler still
checks the combined typed body, emits Go, compares its finite outputs with the
typed interpreter and independently replays source observations before execution.

A fresh request reads its model path only if search is needed. Consequently a
resolved request may skip even opening that path; `model_load_skipped` records it.
A retained worker has already loaded its model in the constructor. Both forms
record zero predictions for resolved requests. Requested seed sampling and
feedback are explicitly marked skipped. Original model/search controls remain
bound to their source document and configuration. Existing receipt behavior is
preserved when the option is omitted or false.

The required `typed_path_observation_resolution` dimension checks the link between
the finite observation and its direct selection or explicit search continuation.
One survivor describes the declared finite candidate set and supplied cases;
inputs beyond that suite retain their existing scope.

### Finite limits

Bounds are 1..32 supplied inputs, 1..64 observed candidates, 1..8 added
observations, and 128 effective cases. A final ranking after the last observation
means at most nine rankings. Each ranking uses a fixed 16 KiB output matrix;
the plan, accumulated receipts and process memory are additional costs.

| Observation status | Interpretation |
| --- | --- |
| `ONE_SURVIVING_CANDIDATE` | Exactly one candidate in the complete declared finite enumeration passes the current cases |
| `NO_SURVIVING_CANDIDATE` | No observed candidate passes; generation retains the existing partial-result behavior |
| `NO_DISTINGUISHING_INPUT` | Supplied inputs do not separate the remaining candidates |
| `CANDIDATE_BUDGET` | Candidate enumeration is incomplete and no distinguishing input was found in the observed subset |
| `ORACLE_UNAVAILABLE` | A useful input is recorded, with its expected value unresolved |
| `ROUND_BUDGET` / `CASE_BUDGET` | A useful input remains but the allowed observations/cases are exhausted |
| `FAILED` | Source, oracle, budget, cancellation or evaluation prevented completion |

Inspect `ranking.unobserved` in every round: a recommendation can be useful even
when enumeration is partial. Status describes this finite procedure. For example,
one surviving candidate can still fail on an input outside the supplied suite.

`body_paths.observation` preserves the original cases, each complete ranking,
oracle identity and generated-code hash, newly observed labels, effective case
hash/count, evaluation count and stopping status. Original source/document/test
hashes stay separate. `timing.observation_ms` isolates the added work.

In reuse mode each round also has a `reuse` record: revision, first-ranking hash,
reused output count, cached comparisons and cumulative evaluations/comparisons.
`ranking.evaluation_attempts` counts work in that round. Runtime replay builds a
new bounded session from the source, checks its first outputs and follows each
oracle observation. Imported cached outputs do not replace that replay. CLI and
stream probes both require explicit integer values; a JSON `null` is rejected.

The shared receipt adds `typed_path_observation_binding` and computes finite
accuracy over the effective suite. Binding `PASS` means the records are connected;
unresolved behavior remains visible in the observation status. The explicit
`body-execute` route replays probe outputs and oracle observations from source
before executing the selected generated package. Generation's `native_case_results`
are bounded AST observations; actual compiled execution is reported separately.

This mechanism measures completion of an explicit contract. It cannot discover a
missing oracle or establish that the oracle captures a person's entire intent.
It is useful when an existing simple specification can guide construction of a
different implementation, or when a supplied test suite leaves several bodies
indistinguishable.

## Research context

The experiment borrows the question of which observation is worth obtaining from
[LAVOIR](https://arxiv.org/abs/2609.30706). Its implementation uses deterministic
candidate disagreement and an existing oracle, with zero additional training.
[Do System One Decisions Add Up?](https://arxiv.org/abs/2609.33971) motivates checking
the assembled program and preserving ambiguity across decisions. These are
research motivations; their reported results are separate from Gooo's evidence.
