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
| `options` | Optional `step_attempts`, `feedback_rounds`, `feedback_unfixed`, `ci` |

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
