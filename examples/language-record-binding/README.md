# Source-bound record transport

This example runs a real Gooo activity graph with structured data instead of
encoding evidence or authority as an integer.

```sh
gooo run --json --entry Capture --record-input examples/language-record-binding/input.json examples/language-record-binding/main.gooo
```

The source declares 6 required string fields, 3 activities, and 2 bindings.
The native executor applies `record.forward:v1` 3 times and delivers 2 opaque
results. Each detached result carries the compiled source and semantic digests,
producer identity, root input digest, and immediate predecessor result digest.

The sample input deliberately has no evidence and stays UNKNOWN in every
result. Transport PASS means the declared dataflow executed, not that the
claim is true. Semantic admission is UNASSESSED and every operation authority
flag is false. Digest agreement proves content identity, not truth or consent.

Version 1 supports required, single string fields only, matching the string
field form used by the existing callback-extraction contract. Other types,
optional/many fields, undeclared fields, missing fields, ambiguous JSON, unknown
operations, cycles, and direct input to a bound activity fail closed.
The existing integer `--input` path is unchanged.

Native API and CLI regression tests use this exact source and input in GitHub
Actions. No local test execution is required. This example is a transport
capability, not a completed callback-admission or external-utility claim.
