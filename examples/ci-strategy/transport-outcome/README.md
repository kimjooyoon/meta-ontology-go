# Assess a transport result in Gooo

`outcome.gooo` classifies the write response and subsequent readback from the
[bounded registry transport](../proof-transport/README.md). Its input keeps the
canonical effect ID, the prior Gooo decision, the observed write response and
the separately observed readback. Gooo owns the result and next-operation rules;
Go supplies those values and the CLI consumes `command_failed`.

```sh
go run ./cmd/gooo package generate --go-package transportoutcome --embed-origin \
  --out examples/ci-strategy/transport-outcome \
  examples/ci-strategy/transport-outcome/gooo.workspace.json
go test ./examples/ci-strategy/transport-outcome ./examples/ci-strategy/proof-transport/...
```

| Observations | Assessment |
| --- | --- |
| No write was attempted | Retain the source decision's next operation |
| The write could not be sent | Preserve the failure and refresh before another attempt |
| Successful response and content seen in readback | `EFFECT_OBSERVED` |
| Failed response and content seen in readback | `EFFECT_OBSERVED_WITH_WRITE_FAILURE`; command still fails |
| Complete readback sees the content absent | `EFFECT_NOT_OBSERVED`; revise before another attempt |
| Missing, expired or unresolved readback | `UNRESOLVED_EFFECT`; collect a current observation |
| Effect IDs or phase-presence claims conflict | `INCONSISTENT_OBSERVATION` |

The adapter performs no automatic retry. `effect_observed` refers to content
presence in readback; it does not establish which writer caused it.
`command_failed=false` preserves the response classification and does not imply
effect completion. The original decision, response/error and readback remain in
the report. Generation origin for this assessment library is retained separately
from the registry policy's origin.

Independent expected results cover 11 response/readback combinations, three
HTTP status boundaries and seven inconsistent phase/effect records. Integration
reuses three existing local registry outcomes, one early return and one orphan
readback decision. Generated-source consistency is
checked against the authored workspace and original compiler. These are finite
regressions for one operation, with no general recovery accuracy or speed claim.
The case adds source-level behavior using existing grammar. A full PROV-O
responsibility graph and independent reader implementation remain later work.
