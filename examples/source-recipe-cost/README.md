# Controlled source recipe decode cost

Fixed before collection, 2026-10-03. Build this identical collector in a clean
baseline revision containing the prior decoder and a clean candidate revision
containing projection reuse. Both use Go1.27.1 and public SDK v0.2.20. Report their
actual commit identities; adding this example changes the build's source SHA.

Use the 128 original model-arm request/budget records (64 requests, budgets 1/8),
verified by the frozen records digest. Each process performs one unmeasured
warm-up per request, then eight fresh decodes. Record the exact returned plan
and complete document digests. No model predictions or selected code generation
are performed by this cost collector. Native semantics are verified separately.

Run four sequential paired trials, alternating baseline/candidate process order.
Neither arm runs concurrently with the other. Planned measured count: 1,024 per
process, 8,192 total. Pair records by trial, request, budget and repeat. All plan
and document digests must match. Summarize latency, allocation bytes/count,
paired differences, and slower cases. Startup/read/hash/serialization and memory
counter reads remain outside each decode interval. Heap/runtime work can affect
allocation deltas. This is a fixed workload on one machine, not general throughput.

```sh
go build -o /tmp/source-recipe-cost ./examples/source-recipe-cost
/tmp/source-recipe-cost --baseline /path/to/order-judge-native-initial \
  --repeats 8 --out /path/to/fresh-output
```

Keep every original JSONL record, manifest, process timing and build identity.
The historical native collection did not confirm a decode wall improvement;
this paired follow-up addresses that unresolved cost observation.
