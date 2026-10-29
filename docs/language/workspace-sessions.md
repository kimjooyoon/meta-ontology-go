# Caller-owned workspace sessions

`gooo package session` keeps one executable owner alive across workspace execute
and replay requests. The caller can inspect an execution receipt, save it, and
choose the next replay input before sending another request. The source remains
in `.gooo`; current source files and inputs are loaded for every request.

From the compiler repository root, this complete stream executes an existing
Gooo record program and then releases its native workspace:

```sh
gooo package session <<'JSON'
{"schema":"gooo/workspace-session-request/v1","id":"first","operation":"execute","arguments":["--cases","examples/record-collections/cases.json","examples/record-collections/gooo.workspace.json"]}
{"schema":"gooo/workspace-session-request/v1","id":"finish","operation":"close"}
JSON
```

Each request occupies one line. Responses use
`gooo/workspace-session-response/v1` and contain the same `id`, an `exit_code`,
and an ordinary workspace execution `receipt` when execution was attempted.
Requests are processed serially with no internal work queue. Relative paths use
the session process's working directory. Request lines are bounded at 64 KiB,
IDs at 128 bytes, and argument lists at 48 entries.

To replay, save the returned **receipt object** as `first-receipt.json`, then
send the following request to the same still-running session:

```json
{"schema":"gooo/workspace-session-request/v1","id":"second","operation":"replay","arguments":["--receipt","first-receipt.json","--cases","new-cases.json","gooo.workspace.json"]}
```

The response frame carries correlation information. Replay parent identity binds
the bytes of the caller-saved receipt through `receipt.replayed_from_sha256`;
the result also binds the serialized prior workspace result. Save and inspect
those records in the same way as ordinary `package execute` and `package replay`.
The framing owns JSON output, so `--receipt-compression` is unavailable in this
surface. Ordinary receipt readers still accept their supported saved formats.
`--json` is optional in request arguments and `--timing` writes the existing
per-operation timing observation to stderr.

Current manifest, source/dependency closure, selected bodies and caller inputs
pass through the same CLI loading and workspace reconstruction paths. The native
owner then verifies the current tool/environment and executable identity. Source,
graph, driver or tool changes keep their existing miss/refusal behavior. Host
facts carried as current inputs are read again; the owner retains the executable
rather than the input values or a prior execution outcome.

Each successful execute or replay request performs two fresh native runs.
`receipt.result.runtime.artifact.source_build` and its digest identify the
historical build of a reused artifact. Current `build.started` can be false while
current `runs`, runtime-suite identity and finite outcomes are newly observed.
Read `receipt.decision` and the current finite denominator: a command exit of 0
can carry `PROGRESS`, and input-only `OBSERVED` has no expected-output score.

Send `close` after reading the final response for explicit cleanup. Stdin EOF
finishes the last request before releasing the owner. SIGINT or SIGTERM cancels
active work and releases owned native resources; it also interrupts an idle
stdio input wait. A close frame is processed between requests. The owner keeps
at most one compiled native artifact. Cleanup errors are returned, and only
owned temporary workspaces are removed.

Malformed request lines return a per-request error and leave the session able to
read a following valid line. An oversized line or broken stream terminates the
session and triggers cleanup. A failed expected value stays in the current
receipt and never becomes a reused PASS. Session closure, compilation and test
observations remain separate lifecycle facts.
