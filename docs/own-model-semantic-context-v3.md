# Direct source features for an own small Gooo model

Explicit `semantic_context_intent_v3` structural model metadata selects compiler
context schema `gooo/compiler-typed-path-context/v3`. The compiler first binds
the original Gooo body to its validated fallback, then projects 64 typed source
features and the complete Korean/English intent. The source array describes
variable references/definitions, assignments, operands, branch effects and root
ordering. It excludes literal magnitudes, spelling of names, stable IDs, expected
answers, finite test outcomes and CI. This is a lossy structural representation,
not a complete semantic proof or unrestricted natural-language-to-code model.

The 256/48/8 model dimensions and 1,248-byte workspace remain unchanged. The SDK
projects source facts into fixed arrays with zero heap allocations in its tests.
Its eight labels rank declared typed alternatives. Gooo owns the possible body
fragments, type/scope checks and deterministic finite continuation. Confidence
does not approve a body or veto another candidate.

## Explicit export

```sh
gooo body-context --plan plan.json --activity ChoosePath \
  --feature-version semantic_context_intent_v3 source.gooo
```

Export schema `gooo/compiler-path-input-export/v2` carries the exact source-bound
text and source-feature SHA256 used by actual ranking. The canonical text is
`gooo;sem64=` followed by 128 lowercase hex digits, `;intent: ` and the complete
natural suffix. As before, the suffix after the caller's last literal `intent: `
replaces caller-supplied context. Normal receipts record digests; only explicit
export discloses text. Deliberately choose public/synthetic sources for training.

The command generates/typechecks original and fallback validation projections.
It makes no model predictions, candidate tests, selected emission or repository
writes. The default command/API retain v2. Existing v1/v2 feature semantics and
previous model bundles are unchanged; a v2 weight bundle cannot silently use v3.

## Continuation and feedback

Both complete initial input and feedback input have a 512-byte UTF-8 limit.
Nothing is truncated. Any initial representation decline skips all optional
predictions, seed sampling and feedback, preserving deterministic search and its
best partial body. Duplicate local definitions and input shadowing decline the
optional projection rather than inventing variable-binding facts. The compiler
still rejects invalid or unreachable body nodes before inference.

SDK v0.2.11-experimental also preserves the prior prediction and workspace when
semantic-v3 input validation fails. Actual-kernel tests verify the error contract
and zero-allocation valid inference; previous feature versions keep their
existing failed-output behavior. V0.2.10's early prediction clearing was corrected
without changing valid features, predictions or weights.

After actual partial tests, feedback keeps the source header bit-identical and
appends observed failures/CI hints to the natural channel. Oversized feedback
records the complete attempted byte count/hash, predicts nothing and leaves the
frontier unchanged. A sole remaining path skips unnecessary ranking. Acceptance
is still the observed finite functional result, not model confidence or CI text.

## Evidence and limits

Contract tests exercise all five structural kinds in Korean/English, reversed
fallbacks, exact export/ranking equality, outcome/seed independence, atomic
decline, partial continuation, source-preserving feedback, cancellation, source
mismatch and eight concurrent requests with independently owned workspaces.
SDK source/tests are pinned by its extraction manifest. The contract weights in
these tests are constant logits, not evidence of learned-quality improvement.

New v3 weights require a separately preregistered fresh composition cohort,
matched training and native dogfood. No new trained weights or accuracy claim
comes from this compiler ABI change. Runtime/orchestration remain Go; offline
GPU training/export is separate.
