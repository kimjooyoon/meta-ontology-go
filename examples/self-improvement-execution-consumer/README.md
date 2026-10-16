# Bounded execution consumer

Think of a grant as a numbered work order: the person writing the order, the
edition of the program and the tools doing the work each have their own origin.
The consumer keeps those identities in its report.

## Source-owned continuation

Nine executable activities in
[`policy/source.gooo`](../../internal/meta/executioncontinuation/policy/source.gooo)
choose subject agreement, grant eligibility, permitted waiting, target-run
binding, observation stages, artifact/workflow roles and the next observation
for missing inputs. Their Go projections come from the existing body generator.
Regenerate them with `go generate` in that policy directory.

The successful v26 grant job automatically dispatches this consumer. The Go
bridge uses GitHub API version `2026-03-10` and binds the returned
`workflow_run_id` to the observed repository, workflow, branch, commit and
original attempt. It sends distinct subject and producer-workflow commits.
GitHub documents the response in its
[workflow dispatch API](https://docs.github.com/en/rest/actions/workflows#create-a-workflow-dispatch-event).

| Recorded state | Observation |
| --- | --- |
| `DISPATCH_UNKNOWN` | No confirmed dispatch outcome; causal fields identify the missing input |
| `DISPATCH_SENT` | Accepted dispatch, with target identity still unresolved |
| `TARGET_RUN_OBSERVED` | Exact returned target run exists with the expected workflow origin |
| `SOURCE_BOUND` | Consumer independently verified grant, contract and exact subject file |
| `EXECUTION_OBSERVED` | Value evaluation was invoked; case results record what happened |

The parent waits for creation. The child can then wait for its parent's successful
completion. Source policy rejects a parent waiting for completion when the child
needs that same completion. An ambiguous dispatch outcome requests observation;
it supplies no blind resubmission instruction.

## Go execution and origins

The Go consumer reads the immutable v26 grant artifact and the exact v25 artifact
named by that grant. It independently verifies the original authorization pair,
contract resolution, grant policy and embedded execution input. It retrieves
the registered `Increment` file at the explicit subject commit and compares its
bytes with the authorized snapshot. The consumer uses its own workflow code.

The command has `prepare`, `reserve`, `consume` and producer-side `dispatch`
modes. The workflow invokes them directly in Go. The evaluator reuses one
compiled Gooo program across the five value calls and verifies each opaque
produced result against that program. Input validation separately recompiles
the snapshot to recheck semantic identity. Reports preserve subject blob/file hashes, both
producer origins, consumer origin and each value result's provenance.

## One-use scope

One reservation artifact is written before the value evaluator runs. Its name
is derived from the grant request digest, and the workflow serializes runs for
the same grant. A later run searches consumer history across workflow commits,
then checks the reservation's immutable archive and typed contents. Unavailable,
expired or unbound history stays unresolved. A local exclusive claim and opaque
invocation handle also reject repeated execution in the same invocation.

The reported scope is `SERIALIZED_CONSUMER_WITH_RETAINED_ARTIFACT_HISTORY`.
Artifacts are retained for 14 days. Continued protection depends on retained,
readable reservation history and this consumer's serialized workflow. If a run
stops after reserving, the grant remains spent within this scope; recovery needs
a fresh system-derived grant.

## What is measured

The consumer evaluates all five fixed corpus cases as one bounded experiment.
The report distinguishes one consumed grant from five evaluator invocations.
It records expected and actual values, execution and result digests, source and
plan identity, per-case elapsed time, compile time, and total evaluation time.
Repository writes and external effects remain zero. Runtime samples are
diagnostic; without a comparable baseline, `performance_improvement` stays
`UNKNOWN`.

Compile and evaluation timers cover the evaluator phase. Source retrieval,
policy and input validation, command startup and queue time have separate costs;
the report supplies no end-to-end speedup measurement.

HTTP reads have a 30-second request timeout and a four-minute command deadline.
History and artifact inventories each have an eight-page bound. Archives are
limited to 16 MiB and 128 entries, with 8 MiB per selected member. Missing
population counts keep the inventory unresolved.

The [observation record](../../docs/observations/consumer-continuation/observations.json)
separates local protocol cases, source registration and actual workflow records.
Local protocol checks do not run the positive five-value consumption experiment;
the new original CI test owns that observation. Runtime improvement and broad
language completeness remain open measurements.
