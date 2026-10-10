# Using condition counterexamples with the existing local model

Observed 2026-10-10 with clean compiler source
`5fba132655ec9145a5fa7bccb5411ace8341bf02`, public decision-runtime
`v0.2.28-experimental`, Go 1.27.2. Compiler binary SHA-256:
`799fc2fc1cd3255857ccff16ba73e247bdf2a9e35ee921278133f6d5f617befb`.
The existing ternary model has 2,759 weight bytes, SHA-256
`dcd8e44591626d421d4961bfef82ec197e947cb1d5d2cd92868d908bc7de4aed`.

We used the same Gooo source, model and seven output cases in both commands.
The source also declares three intermediate condition cases. The first command
ranks once; the second permits two rounds of explicit observed-failure feedback.
Both advance one candidate at a time, with a total budget of eight candidates.

| Observed quantity | Rank once | Two feedback rounds |
| --- | ---: | ---: |
| Candidates evaluated | 5 | 3 |
| Candidates rejected by conditions | 3 | 1 |
| Total local model predictions | 3 | 9 |
| Additional feedback predictions | 0 | 6 |
| Native output cases passed | 7/7 | 7/7 |
| Intermediate conditions passed | 3/3 | 3/3 |
| Sum of model prediction time | 36,542 ns | 173,168 ns |
| Full feedback input sizes | — | 456–493 bytes |
| External provider calls | 0 | 0 |
| CLI wall time (`time -l`) | 0.58 s | 0.01 s |
| Maximum resident set (`time -l`) | 21,839,872 B | 22,020,096 B |

The selected Gooo and Go sources are identical. Exact inputs
`±9007199254740995` both return `9007199254740995`. The Go recount decodes these
values into int64 and checks them without floating-point conversion.

In the feedback arm, the first rejected candidate is mask 2. For the negative
large input, its comparison evaluates to false while the source expects true.
The next three model inputs contain that counterexample, including its original
choice ID and exact integer. After the next candidate, the second feedback round
also includes a final-output mismatch. The complete Korean/English intentions
and source feature headers stay in each input. No context was declined here;
the maximum observed input leaves only 19 bytes below the existing limit.

This is one frozen-model observation, with no training or weight changes. It
compares rank-once execution with the complete feedback mechanism; it does not
isolate the new condition fields from the earlier output-feedback fields.
The extra six predictions accompany two fewer candidate evaluations. These seven
cases are consumed during construction and provide no unseen-input accuracy
estimate. Larger programs and longer intentions need separate measurements.

Commands ran sequentially, baseline first. The wall-time gap may include cache
and process startup effects and is not evidence of a model speedup. Process RSS
includes the compiler/runtime. Whole-machine CPU utilization and model-only peak
RAM were not measured. New saved-runtime evaluation was not part of this study;
the compiler regression suite separately covers saved projection/native replay.

## Recount the original records

`original/` contains deterministic gzip copies of the two complete outputs,
timing logs, source, frozen model and producer build record. Compressed payloads
were decompressed and compared byte for byte with the originals before publishing.

```sh
cd docs/research/condition-feedback-20261010
shasum -a 256 -c SHA256SUMS
GOWORK=off GOTOOLCHAIN=go1.27.2 go run recount.go original > /tmp/condition-feedback-recount.json
cmp recount.json /tmp/condition-feedback-recount.json
```

For a new observation, decompress into a new directory and use the recorded
compiler revision:

```sh
gooo body-codegen --json --path-model model/model.json --path-step-attempts 1 \
  --activity Choose source.gooo
gooo body-codegen --json --path-model model/model.json --path-step-attempts 1 \
  --path-feedback-rounds 2 --activity Choose source.gooo
```

The [earlier paired condition study](../source-conditions-20261010/README.md) uses
a different original compiler and a single local output case. Its recorded
11/11 caller outputs and 1/3→3/3 predicate comparison remain separate observations.
