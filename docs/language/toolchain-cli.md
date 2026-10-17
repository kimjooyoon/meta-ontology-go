# Deterministic toolchain CLI

## Boundary

`TOOLCHAIN-CLI` credits the existing `cmd/gooo` executable only after CI builds
that exact head and a separate metaprogram observes its process contract. The
implementation does not add another router, shell, interpreter, plugin loader,
network client, or repository writer.

The observer accepts one absolute binary outside the repository and maps a
versioned registry to twelve compiled-in argument plans. Registry text cannot
select an arbitrary executable or arbitrary arguments. Each plan runs twice
with a five-second deadline and a 64 KiB output bound.

## Useful properties

- Text and JSON identity are both executable contracts.
- Syntax, semantic, and bidirectional language operations cross the real CLI.
- Exit code, stdout, stderr, and repository tree state are independently
  measured rather than summarized as one green result.
- Each invocation records a positive integer `peak_rss_kib` observation from the
  runner. This value is `RUNNER_SCOPED_NONDETERMINISTIC`: it is visible evidence,
  not replay authority and not a memory or performance improvement claim.
- Unknown observations lower resolution and cannot become readiness credit.

The [gomacro project](https://github.com/cosmos72/gomacro) is used only as a
structural reference for separating a command front door from staged language
evaluation. This project neither imports gomacro nor claims its REPL,
interpreter, macro, or dynamic package capabilities.

No novelty claim is made. The positive element is the combination of a real
binary boundary, fixed meta operations, deterministic replay, and zero-write
evidence in one readiness receipt.

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

## Internal workflow scope

A complete registry evaluation runs 12 cases with a first invocation and replay,
so it makes 24 process-boundary calls. Each call captures a before and after tree.
The transformation-effect workflow writes two reports, then `--check` evaluates
again before comparing the saved report. Across the CLI and format/fix families,
the configured complete path is 2 × 3 × 12 × 2 = 144 invocation attempts and
288 tree snapshots. Early failures can reduce those counts. The five-second
invocation deadline covers the process call; tree snapshots are outside it.
These are source-derived work counts. Actual calls, scan cost and elapsed time
need their own execution records before changing verification or predicting time.

These witness blocks also depend on resolved predecessor inputs. In the original
[#1660 workflow](https://github.com/kimjooyoon/meta-ontology-go/actions/runs/39028706174),
the concept production step was skipped after a blocked predecessor was recorded.
The successful workflow status therefore supplies no execution observation for
these blocks. Read the selected step and actual case results as well as the run status.
