# One owned native executable: fixed comparison protocol

Commit this protocol and collector before collecting results. Use the original
public 16 KiB order judge and frozen 64 known EN/KO development sources spanning
four arithmetic families, with attempt budgets 1 and 8. The preceding native
baseline supplies exact input/source/plan/generated/result hashes.

For each of the 128 source/budget pairs, alternate one-shot and owned-executor arm
order. Each arm loads a generator and makes two fresh source-decoding/generation
calls. Only the owned-executor arm retains its first compiled executable for the
second call. Generation immediately precedes two compiled-program executions;
no later generation starts before those observations finish. Both arms retain
model candidate preparation on their second call. Original source, complete
search, ranking (prediction duration excluded), generated Go and all eight ordered
native outcomes must match. Original finite failures remain in the denominator.

Planned completion is 512 generations/predictions, 1,024 runtime executions,
384 actual builds and 128 artifact reuses. Count actual records after completion.
Measure decoding, generation, execution, current build, first/second run, current
child CPU and maximum single-child peak RSS separately. A prior build's resources
are history, excluded from the reuse call. Alternating arm order reduces a simple
order confound; this bounded same-machine observation leaves cache, scheduling
and first-launch variation present. It establishes no new training/generalization
result. Test changed expectations, tampering, cancellation and concurrency in the
executor and stream suites separately.

Build from a clean committed compiler checkout with Go 1.27.2:

```sh
go build -trimpath -o /tmp/retained-native-runtime ./examples/retained-native-runtime
/tmp/retained-native-runtime --baseline /path/to/order-judge-native-initial \
  --model /path/to/public/model.json --go-bin /path/to/go1.27.2/bin/go \
  --out /path/to/fresh-observations
```

Every generation and runtime receipt is saved. The journal is flushed after each
completed pair of observations. A failure preserves files and stops the collector;
the final records/manifest are written only after the complete cohort. EOF/close
removes owned temporary executables. The CLI integration is described in
[the stream guide](../../docs/native-body-worker.md#generate-and-immediately-execute-in-one-process).
