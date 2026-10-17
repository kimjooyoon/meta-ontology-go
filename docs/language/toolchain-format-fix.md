# Deterministic toolchain format/fix

## Boundary

`gooo format` emits canonical Gooo source or checks that a source is already
canonical. `gooo fix` emits a versioned `gooo/format-fix-plan/v1` value. It
does not accept `--write`, open a repository writer, or apply its own plan.

The plan binds source bytes, result bytes, semantic fingerprints, one bounded
full-file edit, an explicit decision, and zero direct writes. A separate
metaprogram consumes the same plan in memory, rebuilds the result, and accepts
only an explicit `FIXED_POINT`. Unknown source meaning, plan decisions,
resolutions, registry operations, or top-level decisions fail closed and lower
resolution.

## Executable use cases

The fixed corpus contains twelve cases:

- six positive paths: format text, format JSON, canonical check, changed-plan
  JSON, fixed-point JSON, and plan text
- six guardrails: missing input, malformed format input, malformed fix input,
  unknown format option, unknown fix option, and rejected write authority

Every process case runs twice against the exact CI-built `gooo` binary.
Repository trees are measured before and after all twenty-four invocations.
The metaprogram separately performs one in-memory application and observes two
fixed points with zero direct writes.

## Structural reference

The [gomacro project](https://github.com/cosmos72/gomacro) remains a structural
reference for keeping a command boundary separate from staged language
evaluation. Gooo does not import gomacro and makes no claim to its interpreter,
REPL, macro, or dynamic package capabilities.

No novelty claim is made. The useful property is the measured combination of
canonical output, a data-only repair plan, semantic equality, in-memory
application, fixed-point replay, and a denied write boundary.

## Partial invocation observations

An executor may return captured data together with an error. The unresolved case
retains that `first` or `replay` observation, including any observed repository
writes. An error or failure marker on the first observation stops that case before
a replay request. Partial observations do not satisfy readiness or authorize work.

`invocations` counts observations returned without an executor error. It is not a
count of proven operating-system process starts; an errored call can still retain
partial facts. The existing successful-case targets and resource requirements stay
in effect. Unknown collection and process-start details still need separate work.

`Observation.Failure` describes incomplete invocation observations such as a
process deadline, output limit or unavailable required resource measurement.
An expected nonzero exit code is still an ordinary result: the existing refusal
cases execute and replay, then compare against their declared expected exit and
output. Business decisions such as `DENY`, `never` or `allowed=false` do not set
this invocation marker and do not acquire replay suppression from this change.

The process boundary also retains an executable launch error before classifying
missing resource state. If the after snapshot fails, the completed before digest
and process fields remain available; both invocation and snapshot errors remain
inspectable. A missing after digest means file changes are unknown: a zero write
count in that partial record does not establish that no files changed. When both
snapshots are available, their observed write count survives an invocation error.

## Comparing saved reports

The internal witness can compare two saved reports without running Gooo:

```sh
go run ./cmd/toolchain-format-fix-witness \
  --compare first.json --check replay.json --expected-head "$HEAD_SHA"
```

This mode accepts no execution or output configuration. It validates both raw
reports, the requested head and the unique case-proof binding. The comparison
view excludes only `cases[*].first.peak_rss_kib`, `cases[*].replay.peak_rss_kib`,
the already checked case digests, the report digest and the evidence digest of
the unique REGRESSION proof with `reject-unknown-and-write-boundaries`. Source,
executable, tree and other proof digests remain part of the comparison. Raw files
and their digests are preserved.

Exit zero means those two records compare equal. Equally unresolved records can
compare equal; this supplies no successful execution, authenticity, reuse or
effect authority. Ordinary `--check` with execution configuration still performs
fresh evaluation before comparing. Workflow step selection and skipped status
remain separate observations; this report schema does not carry CI step state.

Each saved report has a 32 MiB input limit. The reader consumes at most the limit
plus one byte before rejecting excess input; it does not truncate and compare a
prefix. The native producer currently has 12 cases, two invocations per case and
two 64 KiB output streams per invocation: 3 MiB of captured output, up to 18 MiB when
JSON string escaping expands every byte sixfold, plus fixed report metadata.
The input limit is a reader contract, not a measured bound on total decoder heap.
Custom reports exceeding the limit are unsupported by this internal command.
