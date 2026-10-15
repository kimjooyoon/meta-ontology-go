# A decision includes its declared reason

This example judges a stock request. It keeps quantity, available stock and request
reference in a source-declared `Request`, and returns a separate `Decision` record.
The source rejects a non-positive quantity or a request above stock. For the remaining
branch it owns three result candidates: allowed, allowed with a wrong reason, or declined.

The selection case compares the whole result record. Matching the Boolean alone does
not satisfy it. A held-out case and six native caller inputs have separate values.

```sh
go run ./cmd/gooo package execute --json --cases examples/source-decision-reasons/cases.json examples/source-decision-reasons/workspace.json
```

The four input files preserve the bytes of one fresh stock835 CLI invocation.
Its [observation](observation.json) records compiler `bbfcaebb`, no overlays, selection
1/1, held-out 1/1 and native activity outputs 12/12 across six inputs. The six `Decision`
records contain 18 matched field values; this is a narrower denominator within those
native outputs. The two native executions agreed, all caller inputs were disjoint from
the earlier selection/evaluation inputs, and model/provider calls were zero.

The CLI succeeded. The observation wrapper then used an absent metadata key; the
record was recovered from its original saved stdout. The source was not executed again.
The input files were unchanged and the temporary CLI was hashed and removed.

Reasons are explicitly authored source values. The current receipt keeps activity and
field IDs, inputs, outputs and candidate scores. A proof of the fired branch, concurrent
inventory updates and missing or stale external stock data need separate requirements.

See the [Korean guide](../../docs/language/source-decision-reasons.ko.md) and the
[Mangle comparison](../../docs/language/mangle-and-source-decisions.ko.md).
