# Reassess a pending effect from saved reports

The Gooo `Reassess` activity keeps an earlier command failure while considering
a separate read-only report about the same effect. The existing transport runner
collects that report. The query returns the next source-defined operation.

The native entry point is `transportorigins.ReassessSavedReports(priorRaw, currentRaw)`.
Pass `nil` for the current report first. `observation_needed` then tells an agent
whether to collect another report. For this case, the next operation is
`read-current-effect`: call the existing `prooftransport.Run` with `apply=false`,
serialize its report, and query the pair. The queries themselves read saved bytes.

The result retains both exact report digests, the prior state and failure flag,
and this query program's immutable generation context. An observed effect in
the read report leaves the old command failure intact. An absent effect leads
to `revise-before-new-attempt`; this program performs no write or retry.

Reports with a different operation, inconsistent source content, or an apply
attempt remain unresolved. A supplied empty byte slice is a malformed report;
`nil` explicitly means no current report. The second report describes its own
recorded observation, including its original timestamps. The query does not
attribute an effect to a writer or identify a unique physical invocation.

From the repository root with Go 1.27.3 and a development compiler:

```sh
gooo package generate --go-package transportresolution --embed-origin \
  --out examples/ci-strategy/transport-resolution \
  examples/transport-resolution.workspace.json
go test ./examples/ci-strategy/transport-origins -run 'TestSavedReassessment|TestGeneratedReassessment'
```

The projection's actual compiler is clean revision
`ba353f5af2e94389cae3af05470ae09e83109141`. The workspace has ten source members
and 22 record declarations. Functional tests reuse the production transport
runner and its retained local HTTP fixture; no public registry write is needed.
