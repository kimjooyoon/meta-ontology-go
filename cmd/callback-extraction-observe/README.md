# CI callback extraction observation

Expose the existing Gooo-bound callback observer as JSON without enabling an
automatic rewrite. Run only inside GitHub Actions, with an explicit checkout
and the repository's pinned Go toolchain:

```sh
go run ./cmd/callback-extraction-observe \
  -root . \
  -file cmd/language-readiness-witness/predecessor-selection/pagination_test.go \
  -subject func:TestPaginationFixturesExecuteParserAndHTTPClient \
  -timeout 4m > callback-observation.json
```

Upload the JSON even when the command fails. Standard output contains one
observation; diagnostics go to standard error. Exit 2 means invalid arguments,
exit 1 means observation or output failure, and exit 0 means the bounded
observation completed, NOT that semantic admission or rewriting is authorized.
Do not set CI locally to bypass the observer's execution boundary.

The native observer snapshots source into disposable workspaces and invokes
package tests there. It does not modify the input project. The command uses the
existing callback-extraction.observers:v2 Gooo activity and does not create a
second evaluator or change its denominator of two package executions.

Interpret these independently:

- ObservationDecision records observed package events, missing evidence or a
  counterexample. Preserve REFUTED even when semantic admission is UNKNOWN.
- State and operation_admission remain UNKNOWN; apply_permission is FORBIDDEN.
- CompletedTestRuns counts completed runs, not attempts or passing artifacts.
- External dependency binding remains UNBOUND. Matching finite test events
  does not prove arbitrary semantic equivalence.
- WallMS is per-run elapsed time, not an improvement claim or total CI latency.

CI tests cover argument rejection and the local-execution boundary. Existing
extractor CI tests cover original/generated package observation. This adapter
does not add another expensive package replay to every test suite and does not
claim end-to-end CLI success merely from the extractor tests.
