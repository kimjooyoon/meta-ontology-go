# Repeated source-bound generation and immediate execution

This bounded Go collector exercises `TypedPathGenerator` with the original public
16 KiB order judge. It uses the 64 already observed `new-template` and
`new-constants` requests from the [initial native experiment](https://github.com/kimjooyoon/gooo-neural-decision-experiments/tree/main/publication/order-judge-native-20261003).
The frozen records and manifest digests are checked before use.

## Protocol fixed before collection

For each request and each attempt budget (1 and 8), compare two fresh generators
with two calls on one retained generator. Alternate mode order between pairs.
That is 512 generations, 512 actual model predictions and 1,024 compiled program
executions when the collection completes. These are planned counts until a
complete manifest and all observations have been produced.

Each call decodes the recipe again, binds the original Gooo source, predicts once,
evaluates the selected candidates and emits Go. Save that exact result, then
immediately call `body-execute` for two independent compiled executions on eight
declared cases. Only then start the next generation. The original budget-one
functional failures remain in the observations. The full search, ranking (except
prediction duration), generated Go and ordered native case outcomes must match
the original records. Every result contains a fresh preparation receipt.

Measure `Generate` wall time and process allocation deltas; memory-counter reads
are outside that interval. Constructor and recipe decoding durations are separate.
Serialization and native execution are also outside generation timing. The parent
process allocation counters can include runtime activity. Native child cost is
recorded separately. A retained pair's first call includes candidate preparation;
its second call reuses it. Both perform fresh source binding and model prediction.
This compares API ownership patterns in one revision. The original CLI timings
are historical context; this collector does not establish a CLI speedup.

### Source recipe projection follow-up (2026-10-03)

The next collection keeps the same 512-generation protocol, frozen requests,
weights and native comparisons. Recipe expansion reuses the original checked
projection within that request instead of generating it again for fallback
binding. Source/activity/projection identity and fallback equivalence are still
checked. Generation, training export and independent replay bind fresh source.
There is no source cache across requests.

Compare the recorded recipe decoding interval with the preceding collection
`order-prepared-native-20261003`. Those runs occur at different times; report
their observed costs without treating them as a randomized cross-revision speed
estimate. Full semantic and finite native outcomes must remain identical. This
protocol note is committed before the follow-up starts.

## Run

Build both binaries from the same clean compiler checkout with Go 1.27.2 and
the public SDK v0.2.20-experimental, without replacements. Extract the initial
native observations and keep the original model metadata and weights together.
Choose a fresh output directory. Set paths appropriate to your machine:

```sh
go build -trimpath -o /tmp/gooo ./cmd/gooo
go build -trimpath -o /tmp/order-prepared-native ./examples/order-prepared-native
/tmp/order-prepared-native \
  --compiler /tmp/gooo --go-bin /path/to/go1.27.2/bin/go \
  --baseline /path/to/order-judge-native-initial \
  --model /path/to/model.json --out /path/to/fresh-observations
```

`progress.jsonl` is flushed after every completed generation/execution pair.
`records.json` and `manifest.json` are written only after the full cohort completes.
Failures stop collection and preserve the outputs already written. No training,
remote inference, compiler installation or repository writes are performed.
