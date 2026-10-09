# Native process observations

Development source records process creation and termination waiting separately.
Each newly executed toolchain/build/native process has an optional `timing`
object alongside its existing total, CPU, output-digest and failure fields:

| Field | Meaning |
| --- | --- |
| `timing.start_ns` | Wall time inside the host's `exec.Cmd.Start` call |
| `timing.wait_ns` | Wall time after Start returns, including Wait and pipe I/O |
| `timing.deadline_remaining_ns` | Remaining caller deadline immediately before Start; omitted when no deadline exists |
| `timing.context_at_start_return` | `ACTIVE`, `CANCELED` or `DEADLINE_EXCEEDED` observed when Start returns |
| `failure_phase` | `START`, `WAIT` or `OUTPUT` for the host operation reporting the failure; omitted on success |

Start and wait times sum exactly to `wall_ns`. A failed Start has zero wait time.
A successful Start always calls Wait to release the child and its pipes.
The existing caller context, two-second native budget, one-second pipe wait
delay and process-group cleanup continue to apply. The clock never restarts
between these phases.

A long Start interval locates a delay in host process creation. A long Wait
interval can include runtime startup, program execution, scheduling and pipe
handling. Finding when generated user code began requires further observation.
The context snapshot describes the host deadline at Start's return, while CPU
time continues to come from the child's process state.

Older receipts can have no `timing` object or failure phase. That means these
details were unobserved; their original totals and results remain meaningful.
Saved replay creates fresh process observations and preserves prior history.
Executor cache copies keep timing objects independent of the cached original.

## Release example failures

The native language release examples save successful output under the existing
`<platform>-<example>-construct.json` and `-replay.json` paths. A failed command
saves its exact combined stdout/stderr bytes under `.failed-output`, including
an empty file if the command produced no bytes. This allows partial JSON and
arbitrary diagnostic bytes to survive the failing CI step. Both command and
file-write errors remain available if persistence also fails. The platform
workflow already uploads its output directory on failure.

On 2026-10-09, two Windows readiness failures at compiler candidate `8951c5f8`
had empty output, zero recorded CPU time and an expired native deadline. Those
older records do not identify the Start/Wait boundary. These fields provide
evidence for that distinction in subsequent executions.
