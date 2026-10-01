# Own small models linked to Gooo

An explicit local structural model with feature version
`split_context_intent_ngrams_v2` receives compiler-built context automatically via
`body-codegen --path-plan plan.json --path-model model.json`. The Go compiler uses
SDK `v0.2.9-experimental`. Disconnected and earlier-feature inputs retain their
existing deterministic and legacy behavior respectively.

The compiler first validates every typed alternative and binds the declared
fallback body to authoritative Gooo source. Only then does it build an owned,
fallback-normalized fact snapshot for each decision: result type, decision kind,
target node, one-hop expression/operator/local facts, branch indices, execution
order and legal options. It preserves the entire natural intention after the
last literal `intent: ` separator in the closed feature ABI. A caller's old prefix
is replaced with compiler facts. Display names, test expectations and test outcomes
are absent from initial model input. This is a bounded fact view, not a full
semantic IR parser or independent authority over natural-language intent.

The 512-byte buffer never supplies truncated input. An empty natural intention
or an oversized combined representation records
`DECLINED_TO_DETERMINISTIC`, the original/natural intent hashes and complete
attempted byte count. The compiler skips the optional model ranking, sampling seed
and feedback for that request, records those skips, and continues finite TDD with
normal fallback order, candidate budget, cancellation and native verification.
Invalid models, types and source bindings still fail before prediction.

`body_paths.model_context` links the original plan, ranked plan, source semantic
digest, stable activity ID, model metadata and each encoded input digest. Context
preparation time is separate from model loading, bounded search and final emission.
Retained workers share immutable weights and create fresh per-request contexts.
Initial predictions precede candidate tests; candidate evaluation makes no new
model calls. Explicit feedback can add observed finite failures and caller CI
hints, with its existing representation limits and receipts.

Tests use synthetic weights to check ABI, ownership, all five path families,
Korean/English suffix preservation, source binding, overflow continuation and
concurrent retained calls. These tests do not establish trained language quality.
Finite functional completeness is passed/declared cases, separate from complete
compiler lowering. Training and actual-model studies live in the public research
repository, with immutable model pins and preserved negative outcomes.
