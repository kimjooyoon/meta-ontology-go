# Source declaration split checks

Use the existing native Go command to inspect a candidate before scheduling a
write. It reads the same source plan and topology constraints as write mode:

```sh
go run ./scripts/source-splitter -root "$ROOT" -metrics "$METRICS" \
  -sha "$SHA" -check -subject path/to/source.go
```

An explicit subject exits unsuccessfully when its declaration cannot fit a
rendered part, the projected directory exceeds its limit, the directory has
mixed topology, or the source no longer needs splitting. The error retains the
subject and original planning reason. Missing subjects, incomplete directory
evidence, and mismatched repository/evidence revisions are also errors. No
source writes occur during this check.

Without `-subject`, `-check` remains an inventory operation: it reports planned
and blocked counts, and a blocked candidate alone does not fail that inventory.
The existing repository-wide CI uses this form. Agents deciding whether to
attempt one split should use the explicit-subject form and its exit status.

The check matches the repository HEAD with the metrics revision and reads the
working source. It does not attest that the working tree is clean, reserve
future output paths, test filesystem writes, or compile the resulting package.
Write mode rechecks the plan and still owns output collision and write errors.
Keep the source and metrics fixed between planning and execution.

This exposes an existing operation's read-only prerequisite check. Connecting
such checks to Gooo's operation selection and successor verification is tracked in
[#1782](https://github.com/kimjooyoon/meta-ontology-go/issues/1782).

## Candidate selection and source origin

The native planner lives in
[`internal/meta/sourcesplit`](../../internal/meta/sourcesplit). The command and
the self-improvement planner share its parsing, rendering, import normalization,
declaration partitioning, and logical topology checks. Metric scanning does not
run those checks. `refactor-metrics -plan ... -check` observes split candidates
in the existing priority order until it has enough independent actions.
A blocked candidate does not consume its independence group.

The existing [Gooo operation contract](../../internal/meta/generation/operation-input-contract.gooo)
declares the chain:

```gooo
activity InspectSplitRendering(FileInput) -> SplitRenderingObservation
activity InspectSplitTopology(SplitRenderingObservation) -> SplitTopologyObservation
activity RequireSplitProjection(SplitTopologyObservation) -> SplitProjectionObligation
```

Lowering supplies `used` and `wasGeneratedBy` facts. A candidate observation
links its indicator, repository HEAD, inspected source-byte digest, complete
metrics-payload digest, and contract source/semantic digests. The plan digest
also includes the observations. These identify the inputs and declared
relations; the read-only native inspector supplies the actual measured result.

| State | Candidate treatment |
| --- | --- |
| `READY` | Eligible for selection; still requires normal write and conformance checks. |
| `BLOCKED` | Retain the rendering/topology cause and prerequisite work; inspect the next candidate. |
| `UNKNOWN` | Retain a missing native consumer, source, or directory observation; inspect the next candidate. |
| `DEFERRED` | Keep successor projection outstanding; inspect the next candidate. |

All raw indicators and their original outcomes remain in the decision ledger.
The two-independent-operation floor is unchanged. If too few candidates remain,
the result is a shortfall with no selected writes, rather than a fixed point.
The executor reobserves prerequisites before materialization, including for
legacy plans, and compares recorded origins when present. Returned split source
bytes must also match the recorded source digest before evidence is accepted.

An operation declaration does not establish that every execution consumer
implements it. The canonical native registry exposes `execution_route`, which
the common executor uses to dispatch its handlers. Observed planning records
`EXECUTION_CONSUMER_UNAVAILABLE` for a registered operation with no such route.
It retains the metric and declared operation without pretending to inspect its
source. For example, Gooo section splitting remains registered, but its
source-repacker has no common-execution adapter yet. The planner can continue to
another independent operation; connecting the adapter remains explicit work.
Tests require every advertised route to have a handler and every handler to
have a canonical binding.

Function extraction uses the same preparation stage as its committing command,
including active recipe precedence, the generic extractor, import/module checks,
and final formatting/capacity checks. Existing Gooo obligations such as
`ProveCalleeEffects` and `ProveRenderedCapacity` retain their native failure
stage, step, reason, classification, diagnostics and next operation. An unknown
callee effect stays unknown while the planner considers the next candidate.
Preparation builds proposed files in memory; it does not install them. Execution
reobserves the candidate and checks the copied source bytes before writing.

Current coverage is preparation for declaration splitting and function
extraction. Readiness still requires normal writing, conformance and verification
at execution; it does not establish a completed repair. A separately projected
split workspace retains a
`SUCCESSOR_PROJECTION_REQUIRED` obligation: logical directory counts cannot
establish physical storage capacity. This change does not implement that
successor consumer, reserve output paths, or make concurrent source edits atomic.
The follow-up must use the existing repository projector and its verifier before
such a deferred candidate can be admitted.

Native performance observations can be requested separately from CI:

```sh
go test ./internal/meta/sourcesplit -run '^$' \
  -bench '^BenchmarkSplitPrerequisite$' -benchmem -count=5
```

The timed region is one candidate's read-only observation. Repository setup,
metrics scanning, HEAD verification, code writes, and tests are outside it.
