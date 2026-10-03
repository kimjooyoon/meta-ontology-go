# Repeated body construction with one Gooo process

## Start from source, recipe and expectation files

Use one command to construct a body and immediately execute it twice. No input
JSON lines or helper binary are needed:

```sh
gooo body-path-run \
  --source examples/body-codegen/typed-path-compound.gooo.fixture \
  --activity Combined \
  --path-plan examples/body-codegen/typed-path-compound-plan.json \
  --cases examples/body-codegen/typed-path-runtime-cases.json \
  --out body-run-results --repeat 2
```

Choose a fresh `--out` path; an existing path is rejected. `--activity` is an
explicit source activity name. Both full typed plans and short source recipes
work. Add `--model /explicit/path/model.json` to use a local compatible judge;
omitting it uses deterministic construction. `--go-bin /path/to/go1.27.1/bin/go`
selects the native tool. Repetition is sequential, bounded to 1..16, default 1.
One generator and one native executor serve all repetitions, using the same
request evaluation as `body-path-stream --execute`.

The executor keeps the first successful Go version check for a native Go1.27.1
`cmd/go` executable. Every request hashes the current Go file and checks its
embedded version/platform. The retained check is reused only when tool path,
bytes, child environment and producer context match. Shell wrappers receive a
fresh version process on every request. Current source/parent replay and the two
compiled executions still run each time. A changed environment/tool invalidates
the retained version check; closing or cancelling the executor clears it.

Runtime profile `gooo/typed-path-runtime-v3` records `toolchain_reference` and
the original `source_check`, exact `source_output` and content binding. A reused
version leaves the current `toolchain` process record empty. With both version
and build reused, `runtime_child_resources` counts only two current runtime
children. If the version is reused but a new build is needed, it counts three.
Earlier v1/v2 observations remain readable. Profile and resource-unit changes
are preserved by `completeness-delta`; they do not create numeric improvement
claims across different contracts.

The directory contains exact input copies, `model-retention.json`, and each
`run-N-response.json`, `run-N-generation.json`, `run-N-generated.go` and
`run-N-runtime.json` when that stage exists. A `summary.json` records each
completed observation before it is written to stdout and before the next request.
Failed construction and native observations are preserved. Stdout emits one full
result per request; stderr prints status, finite expectations and artifact reuse.
Inspect generated code directly in `run-N-generated.go`.

The terminal says `finite expectations unobserved (128 declared)` when valid
runtime expectations exist but no outputs were observed. A rejected request
without a runtime observation says `finite expectations unobserved`. Completed
replay keeps `passed/observed`, including a measured `0/128`. If outputs exist
but replay is incomplete, the terminal preserves `observed passed/observed
(replay incomplete)`; a differing declared count is shown separately. The
saved `summary.json` still records the original `passed`, declared `total`,
status and actually started `native_runs`. Read the runtime case observations
and `runtime_replayed` alongside that summary when consuming failures.

`response_ms` includes request decoding, construction and native execution,
excluding initial model setup and saving/output costs. Inner generation/runtime
intervals are parts of that response, so do not add them to it. `native_runs`
counts actually started native children. A completed observation with unmet
finite expectations still exits 0 and retains the numerator/denominator. Rejected
construction or native errors exit 1 after recording the result. Usage errors
exit 2. Output/filesystem errors and cancellation also exit 1; completed files
remain available. Cancellation closes closeable stdout/stderr, joins native
children and removes the owned temporary executable. Embedding requires output
writers with prompt `Close` or cooperative nonblocking writes.

An optional `--options options.json` passes the stream's exact options object,
including caller-supplied CI context, finite ambiguity diagnosis and observation
settings. For example: `{"diagnosis":{"inputs":[2,3],"max_candidates":2}}`.
CI hints require a feedback-capable local model, nonzero `step_attempts` and
explicit `feedback_rounds`, following the stream contract below. The whole-candidate
order judge uses ordinary unseeded bounded search; its current profile rejects
batch/feedback options. A CI-only options object is rejected before construction.
Duplicate, aliased, unknown and trailing option fields are rejected. CI hints are
context for construction; source-bound checks and current execution remain the
measurement. Input files must be nonempty regular UTF-8 files: source <=128 KiB,
plan <=256 KiB, cases <=32 KiB and options <=64 KiB. Runtime cases remain 1..128.
The supported native body is the closed pure `Integer -> Integer` projection.

On Unix, source, plan, cases and options are opened without waiting for a FIFO
writer, then validated using the opened file descriptor. A pathname replaced
between discovery and opening cannot defer that descriptor check until a writer
appears. A symlink to a regular input retains its exact bytes. Other platforms
keep the preliminary path check and the descriptor check after opening. The
nonempty/UTF-8/byte limits above still apply. Missing-file and permission failures
retain the underlying filesystem error; nonregular inputs fail before model preparation
or output directory creation.

## Read the cost of each stage

Add `--timing` to `body-path-run` to save `run-N-timing.json` and
`timing-summary.json`. The latter reports milliseconds by phase for each request.

The terminal prints `unobserved` for a missing phase, including a native pair
where only one run started. An observed zero duration remains `0.000ms`.
The executor lazily retains one 32 KiB hash buffer behind its existing gate and
releases it on close or cancellation. Every current tool/executable check still
reads the complete file; retained digests do not stand in for current bytes.
Reads are bounded to 256 MiB plus one detection byte, and a changed read length
fails the binding. These limits do not establish an atomic filesystem snapshot.
On Unix the hash reader opens with a nonblocking flag and rejects nonregular
files before reading. An executable FIFO without a writer therefore returns
`Go tool must be a regular executable file` instead of waiting for a writer.
Ordinary regular-file symlinks remain supported. Other platforms inspect the
file kind before opening and validate the opened file again. These checks bound
the FIFO case; regular filesystem I/O still depends on the operating system.
Stderr also shows generation, source replay, two native runs and artifact saving.
The recorder retains at most 32 sequential intervals per request and creates no
background workers. Without the flag, no timing sidecars are created.

```sh
gooo body-path-run --source original.gooo --activity Name \
  --path-plan recipe.json --cases cases.json --out measured-body --repeat 2 --timing
gooo body-path-run --verify-timing --out measured-body
```

The read-only verification checks ordered nonnegative intervals, accounting,
the original summary, and SHA256 bindings to the saved request, response,
generation, generated Go and runtime files when those stages exist. It loads no
model and starts no native child. A changed file, incomplete interval record or
mismatched summary returns exit 1. Verification measures consistency of these
saved files; execution evidence is still the original runtime receipt.

`response_ns` ends before saving. `wall.capture_ns` extends through original
artifact saving and file hashing. It excludes initial input/model loading,
timing sidecar and summary writes, stdout and final executor cleanup. Phase
durations plus `unassigned_ns` equal capture time; unassigned time includes
inter-stage bookkeeping. Missing stages are unobserved, rather than zero-cost
operations. Generation's existing inner timing is part of its outer interval.
`executable_prepare` includes cache lookup, executable binding, temporary
workspace preparation and a build when needed. `native_run_N` includes child
startup, waiting, I/O and process accounting. `toolchain_bind` covers current
embedded build-info validation and the retained version-check lookup.

The intervals describe current wall costs. CPU and RSS come from the separate
current-child process records; historical `source_build` and `source_check`
records are not current CPU work. Whole-host CPU utilization remains unobserved.
The stream result, original file summary and runtime receipt profiles keep their
existing schemas. Failed requests retain only the stages actually reached.

## Use streaming requests

Use the installed `gooo` command. Each line in `requests.jsonl` is one request:

```sh
gooo body-path-stream --workers 1 < requests.jsonl > results.jsonl
gooo body-path-stream --model /explicit/path/model.json --workers 1 < requests.jsonl > results.jsonl
gooo body-path-stream --help
```

Omit `--model` for deterministic construction. The optional model path is explicit;
an unreadable or incompatible model stops setup with exit code 1. The command
does not download a model. To try a source recipe from this repository root:

```sh
jq -cn --rawfile source examples/body-codegen/path-recipe.gooo.fixture \
  --slurpfile document examples/body-codegen/path-recipe.json \
  'range(2) as $i | {schema:"gooo/native-body-stream-request/v1",correlation_id:("recipe-"+($i|tostring)),
    source:$source,activity:"Compose",document:$document[0]}' \
  | gooo body-path-stream > results.jsonl

jq '{id:.correlation_id,status,error,preparation:.response.report.body_paths.whole_candidate_preparation}' results.jsonl
jq -s '.[1].response' results.jsonl > generation.json
```

`generation.json` has the same shape as `gooo body-codegen --json`; its `source`
contains generated Go. Pass it to [body-execute](source-path-recipes.md)
with the matching original source, recipe and independent cases. The
[public two-request model example](https://github.com/kimjooyoon/gooo-neural-decision-experiments/tree/main/examples/whole-candidate-order)
includes all these files and a small model. A recipe describes permitted choices;
the model ranks those choices and the compiler checks the resulting body.

The standalone `gooo-body-worker` remains available with the same protocol and
options (`go build -o gooo-body-worker ./cmd/gooo-body-worker`, Go 1.27.1).
Both entry points use one command implementation.

The optional local model loads once. Standard error emits one
`gooo/retained-path-model/v1` setup record with metadata/weight hashes and decoded
tensor bytes. Standard output emits a JSON result as soon as each request finishes,
without waiting for input EOF. Worker counts are 1..8, default 1. Results can finish
out of order; join by sequence plus correlation ID. There is no network provider.

Each input line is one object with these fields:

| Field | Value |
|---|---|
| `schema` | `gooo/native-body-stream-request/v1` |
| `correlation_id` | 1..128 UTF-8 bytes, no control characters |
| `source` | Original Gooo source string, at most 128 KiB |
| `activity` | Named source Integer -> Integer activity |
| `document` | Full `gooo/body-codegen-typed-path-plan/v1` document or short `gooo/source-typed-path-recipe/v1` recipe |
| `options` | Optional `step_attempts`, `feedback_rounds`, `feedback_unfixed`, `ci`, `diagnosis`, `observation` |
| `execution_cases` | With `--execute`, a required `gooo/body-runtime-cases/v1` object containing 1..128 current integer input/expected pairs |

Step attempts are 0..64 (zero means ordinary bounded search); feedback rounds are
0..16. Feedback requires a model, nonzero step and explicit rounds. CI context is
optional `{ "source_sha": "<40 lowercase hex>", "status": "PASS|FAIL|UNKNOWN" }`;
it is caller context, not authority. Disconnected mode accepts no feedback or seed.
Input is bounded to 1 MiB per line, with canonical lowercase ASCII field names,
depth <=64, no duplicate or unknown keys. Oversized lines are consumed without
growing retained line storage and rejected; the following line can proceed.

Every request expands its recipe, binds fallback to authoritative source, creates
a private search/feedback session, then performs native emission, type/replay
checks and finite-case evaluation. Source files are never written. The whole-body
order model retains immutable weights and at most one prepared candidate plan.
An identical plan reuses that preparation; each request still predicts and checks
its current cases. Changed plans replace the retained plan. Parallel misses may
prepare independently. `body_paths.whole_candidate_preparation` reports the plan
digest, candidate count, acquisition duration and `reused`. Recipe expansion and
source binding still happen for every request. Other model families retain weights.
The result's `model_retention.setup_ms` repeats constructor provenance and is
excluded from per-request `timing.total_ms`; `model_load_ms` is zero. Do not add
setup once per result. Existing fresh-process CLI receipts remain unchanged.

Both queues have capacity equal to workers; there is no unbounded reorder cache.
Each native request inherits cancellation and the existing eight-second budget.
Cancellation closes transport streams to unblock reads/writes and joins workers.
Embedding requires cooperative generators and I/O closers that unblock promptly.
Normal EOF does not close caller-owned streams. Malformed/body-mismatched requests
are rejected individually; native failures retain their partial receipt.

Exit code 0 means the stream reached EOF (or help was shown). Inspect every result's
`status`: individual rejected requests do not make the process fail. Invalid
arguments exit 2; setup, input/output and cancellation failures exit 1. With
multiple workers, match results by `sequence` and `correlation_id`, rather than
assuming output order. Setup metadata and diagnostics go to stderr; results go
to stdout so they can be piped to another program.

This is an explicit experiment. Finite completeness can be partial. Transport
`completed` means verified native construction, not perfect natural-language
understanding. Packed model storage is separate from resident decoded arrays.
Worker memory and startup amortization require measured evidence; this document
does not assert a speedup, all-input correctness, or arbitrary text generation.

## Generate and immediately execute in one process

Add `--execute` and put the independent suite in each request:

```sh
jq -cn --rawfile source examples/body-codegen/typed-path-compound.gooo.fixture \
  --slurpfile document examples/body-codegen/typed-path-compound-plan.json \
  --slurpfile cases examples/body-codegen/typed-path-runtime-cases.json \
  'range(2) as $i | {schema:"gooo/native-body-stream-request/v1",correlation_id:("run-"+($i|tostring)),
    source:$source,activity:"Combined",document:$document[0],execution_cases:$cases[0]}' \
  | gooo body-path-stream --execute --go-bin /path/to/go1.27.1/bin/go > results.jsonl

jq '{id:.correlation_id,status,error,reused:.execution.observation.artifact.reused,
  cases:.execution.observation.cases}' results.jsonl
```

An explicit `--model` works with the same option. Its absence uses deterministic
construction. Generation finishes before this request starts native execution;
each result includes the original generation and a new `execution` receipt.
Responses arrive before input EOF. A failed native setup/build/run returns
`execution_failed` and preserves the generated body and partial runtime record.
`completed` means the runtime observation finished; read its finite numerator and
denominator to see unmet expectations. Individual failures still leave the stream
available for the next request. Without `--execute`, `execution_cases` is rejected.

One stream owns at most one temporary executable/workspace. Matching generated
Go, driver contract, tool bytes/location/version, platform, producer and child
environment reuse that artifact. Every call freshly replays the source, plan and
parent receipt, hashes the executable before each run, and executes the current
input array twice. Changed expectations produce new observations. Changed keys
or artifact bytes require a new build. Close/EOF removes the workspace; cancellation
joins active children and removes owned files. Linux/macOS cancellation targets
the child process group; other platforms use the direct-child mechanism.

Construction can use 1..8 workers; native observations share one cancellable gate.
They execute serially, with queue wait recorded in `artifact.wait_ns`. The 60-second
runtime deadline includes waiting; each compiled-program run has a two-second
limit. Mixed projections can replace the slot frequently. A source-bound original
build appears in `artifact.source_build`; `observation.build` records only work
done by the current call. Reuse has three current children (version plus two runs)
and excludes prior build cost. Owned execution uses runtime receipt profile v2;
the standalone `body-execute` v1 behavior remains available.

The runtime compiles the closed, pure `Integer -> Integer` projection and a stdlib
array driver. Its Go tool is pinned to 1.27.1. Source checks constrain the emitted
body; the caller still chooses the local Go executable and inherits host permissions.

## Optional finite ambiguity diagnosis

SDK 0.2.8 adds `options.diagnosis`, for example
`{"inputs":[2,3],"max_candidates":2}`. It requires 1..32 integer inputs and
1..64 candidate observations. After bounded selection, the same request deadline
governs deterministic candidate compilation and probe evaluation. Diagnosis makes
zero model predictions, keeps completed observations on interruption, and records
its option hash, candidate budget and extra time separately.

The receipt's `diagnosis` counts candidates indistinguishable by declared case
outputs and reports the first supplied input whose alternative output differs.
The two witness outputs are observations, with expected results unset. Agreement
on every probe remains unresolved bounded evidence. These alternative values are
typed-arena observations; native emission and ordinary case checks concern the
selected body. Alternatives are not independently native-executed by this API.
Source files, selection and finite acceptance are unchanged by diagnosis.

The fresh CLI accepts `--path-diagnosis diagnosis.json` alongside `--path-plan`.
The file schema is `gooo/path-diagnosis-request/v1`, with `inputs` and
`max_candidates`. The examples in `examples/body-codegen/path-diagnosis-*`
show a passed sparse test with two possible subtraction bodies and a distinguishing
input. This is a runnable construction fixture, not natural-language accuracy.

## Optional observations before selection

An observation request can set `resolve_unique_candidate: true`. A complete
enumeration with one survivor then proceeds directly to checked projection.
`body_paths.resolution` records the choice and skipped work; `search_started`
remains false. The retained constructor may already have loaded the model, while
that request performs zero predictions. Requests with partial or ambiguous
observations continue their normal search. Each request owns this state.

SDK v0.2.16 and the compiler's `options.observation` support a bounded observation
loop before candidate selection, with fields `inputs`, `max_candidates`,
`max_rounds` and optional `oracle_activity`. The oracle is a separate pure activity
in the same request source. The compiler appends its observed labels to a private
copy of the cases, then starts model-guided or deterministic search. See
[the observation loop](path-observation-loop.md) for the corresponding CLI,
recorded identities, partial states and independent execution replay.
