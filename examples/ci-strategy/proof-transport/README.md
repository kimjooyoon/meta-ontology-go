# A Gooo plan with a bounded external operation

This case continues the [original recovery proposal](../recovery-origin/README.md).
Gooo declares the exact subject, existing attachment, effect identity, input
requirements and resource limits. Its `TransportNext` activity selects the next
operation from typed observations. Go performs HTTP, hashes the bytes it reads
and supplies a fresh observation after an attempted effect.

The supported effect is one PUT of the original attachment manifest when its
subject and blobs are already present and a complete native referrers response
shows it absent. The original manifest bytes are retained in `testdata/manifest.json`.
This deliberately small operation leaves image creation, blob upload, signing
and legacy referrer-index updates to distinct plans.

## Use the source and its generated library

From the repository root:

```sh
go run ./cmd/gooo package generate --go-package prooftransport \
  --export-stage Plan=gooo://ci-strategy/transport-plan \
  --out examples/ci-strategy/proof-transport \
  examples/ci-strategy/proof-transport.workspace.json
go test ./examples/ci-strategy/proof-transport/...
go run ./examples/ci-strategy/proof-transport/cmd
```

The last command reads the public registry and emits JSON. A calling execution
environment can request `-apply` and supply `GOOO_REGISTRY_BEARER_TOKEN` with the
required repository permission. The source decision still determines whether a
write is needed. `Run` accepts the runtime's token directly; the example CLI can
obtain an anonymous read token. Tokens are absent from the report.

`Plan(operation)` exposes the Gooo declaration through the shared generated
record type. `Evaluate(operation, observation)` composes that plan with the Gooo
decision. `Run` reads its own plan instead of accepting a caller-modified scope.
The prior plan, new plan and canonical effect keep distinct identities. Stage
comments, field tags and source digests preserve the generation correspondence.
The report is typed execution data; a PROV-O graph export remains separate work.

| Source result | Next operation |
| --- | --- |
| `INVALID_PROPOSAL` | Select the supported operation |
| `REFUTED_MATERIAL` | Restore the original manifest bytes |
| `UNRESOLVED_MATERIAL` | Collect the required subject and blobs |
| `REFRESH_OBSERVATION` | Obtain a complete, current observation |
| `ALREADY_PRESENT` | Retain the current observation |
| `READY_TO_RESTORE` | PUT the pinned manifest once, when apply is requested |
| `REVISE_DISCOVERY` | Use a separate plan for discovery repair |

A partial page may establish that the exact attachment is present. Inferring
absence requires a complete list. The reader retains all Link headers, rejects
missing/null lists and missing descriptor digests, and accepts an omitted
optional index media type. These boundaries follow the
[OCI index](https://github.com/opencontainers/image-spec/blob/v1.1.1/image-index.md)
and [distribution](https://github.com/opencontainers/distribution-spec/blob/v1.1.1/spec.md)
specifications. The reader uses only the descriptor fields needed by this case;
it is not a full OCI schema validator.

## Limits and separate observations

The Gooo plan currently sets a 30-second execution deadline, 32 HTTP request
attempts, 65,536 response bytes per request, 262,144 response bytes in total and
a 60-second observation age. The bounded reader may consume one extra byte to
detect an exceeded limit. GET redirects have a two-hop limit, retain the same
budgets and drop the registry token on another host. PUT is attempted at most
once per call, without automatic retries.

Concurrent apply calls serialize within one process and observe the registry
again after acquiring the slot. Different processes can still send the same
content-addressed PUT; the implementation does not provide distributed exactly-once
execution. The read and write are also separate registry operations, so another
writer may change availability between them.

The write response and subsequent readback are both retained. An HTTP failure
can accompany an applied effect; an expired readback remains unresolved. A
failed write response keeps a nonzero CLI exit even when readback sees the
attachment. A read-only decision can contain a proposed write while
`write.attempted` stays false. Consumers should inspect both fields.

The library's request/byte counters cover its registry observations and write.
CLI anonymous authentication is one separately bounded prerequisite request,
outside those counters but inside the parent deadline. Counters describe HTTP
client attempts and body bytes read, rather than network delivery, headers or
uploaded bytes. Signature verification belongs to a separate verifier; hashes
here establish correspondence with the pinned material.

## Fixed regression denominators

Tests use retained public bytes and a local HTTP transport fixture. Expected
states are written separately from the Gooo implementation.

| Scope | Cases |
| --- | --- |
| Typed source policy observations | 20 variations plus one unsupported operation |
| Write response and independent readback | 3 combinations, including a failed response with an applied effect |
| Same-process concurrency | 3 apply calls over one retained subject |
| Referrer media type, descriptor presence and multiple Link headers | 6 variations |
| Cancellation | Before admission and after an applied effect |
| Regeneration | One exact workspace, its source bindings and selected stage |

Additional checks cover read-only calls, altered material, redirect credentials,
request/byte exhaustion and absent/null discovery lists. This is one historical
operation with finite counterexamples. General recovery accuracy, latency gains,
distributed coordination and arbitrary registry compatibility remain unmeasured.

Initial case cost is 129 authored Gooo lines, 334 handwritten production Go
lines for HTTP/execution/CLI mechanics and a 373-line generated projection.
The source-registry entry and independent tests are counted separately. A later
policy change's marginal Go cost is still unmeasured for this adapter.

A development-source read on 2026-10-27 observed the original subject, manifest,
payload and configuration at their expected content addresses. Discovery used
the registry's legacy referrer tag. Gooo returned `ALREADY_PRESENT`: 8 library
HTTP attempts, 13,857 response body bytes, one preceding anonymous-token request
and zero writes. That observation established then-current availability. The
separate signature-verification command had timed out, so it supplies no success
claim. The observed source was subsequently saved in commit `9f30c071`; its
execution binary was not retained. Current tests and later source revisions keep
their own identities instead of relabeling that earlier observation.
