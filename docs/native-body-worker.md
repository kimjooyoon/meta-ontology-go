# Retained native body construction experiment

Build with Go 1.27.1:

```sh
go build -o gooo-body-worker ./cmd/gooo-body-worker
./gooo-body-worker --model /explicit/path/model.json --workers 1 < requests.jsonl
./gooo-body-worker --workers 1 < disconnected-requests.jsonl
```

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
| `document` | Full existing `gooo/body-codegen-typed-path-plan/v1` document |
| `options` | Optional `step_attempts`, `feedback_rounds`, `feedback_unfixed`, `ci`, `diagnosis`, `observation` |

Step attempts are 0..64 (zero means ordinary bounded search); feedback rounds are
0..16. Feedback requires a model, nonzero step and explicit rounds. CI context is
optional `{ "source_sha": "<40 lowercase hex>", "status": "PASS|FAIL|UNKNOWN" }`;
it is caller context, not authority. Disconnected mode accepts no feedback or seed.
Input is bounded to 1 MiB per line, with canonical lowercase ASCII field names,
depth <=64, no duplicate or unknown keys. Oversized lines are consumed without
growing retained line storage and rejected; the following line can proceed.

Every request freshly prepares its typed plan, binds fallback to authoritative
source, creates a private search/feedback session, then performs native emission,
type/replay checks and finite-case evaluation. Source files are never written.
Immutable model arrays are the only construction state shared between requests.
The result's `model_retention.setup_ms` repeats constructor provenance and is
excluded from per-request `timing.total_ms`; `model_load_ms` is zero. Do not add
setup once per result. Existing fresh-process CLI receipts remain unchanged.

Both queues have capacity equal to workers; there is no unbounded reorder cache.
Each native request inherits cancellation and the existing eight-second budget.
Cancellation closes transport streams to unblock reads/writes and joins workers.
Embedding requires cooperative generators and I/O closers that unblock promptly.
Normal EOF does not close caller-owned streams. Malformed/body-mismatched requests
are rejected individually; native failures retain their partial receipt.

This is an explicit experiment. Finite completeness can be partial. Transport
`completed` means verified native construction, not perfect natural-language
understanding. Packed model storage is separate from resident decoded arrays.
Worker memory and startup amortization require measured evidence; this document
does not assert a speedup, all-input correctness, or arbitrary text generation.

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
