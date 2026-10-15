# Native record-field history observations

The question is whether actual field reads retain distinct origins after copies,
field writes, fixed helper arguments/record returns and an atomic record swap.
The new journal stores source identities and evaluated relationships separately
from values. A Gooo program also consumes its saved flat records.

## Original local cohorts

| Cohort | Observed result | Scope |
| --- | --- | --- |
| 13421 | 4/4 native outcomes and exact two-run journal equality | Direct native projection backend; per-field lineage through copies, helper parameters and local record returns; an atomic swap; equal zero payloads retain different branch-selected input-field roots. Wrong-field reference rejected. |
| 13517 | Actual workspace example 4/4; 60 field writes, 56 reads, 12 frames, 32 parent links | Full execution pipeline and saved-record paging. Eight saved facts passed to the real 32-field Gooo reader matched eight external comparisons. Producer executions/model calls during reads were zero. |
| 13601 | Large loop build failed before user-program execution | Generated scalar callback returns were wrapped as activity record returns, losing the contextual integer type. Original build diagnostics and source are retained. The native projection now excludes nested internal-function returns from activity return provenance. |
| 13627, in the same runner | One matching result and one intended arithmetic fault across two new inputs | Interrupted field assignment remained incomplete; prior completed assignments were preserved; two native runs agreed. |
| 13649, in the same runner | Seven inconsistent saved claims rejected | Saved 13517 record only; stale/wrong-field references, clock/order, missing summary, reason/counter, and changed linked field ID with resealed metadata. Absent historical field journals remain absent. Native/model executions zero. |
| 13811 | New source and 24 inputs matched all integer, boolean and string outputs | 73,728 helper calls per native run; journal stops at 4,096 writes, 65,536 reads and 65,536 frames. Actual computation and exact two-run replay complete while explicit drops remain. |

Completed historical source, test and model cohorts were not rerun locally.
The failed loop cohort remains a failed original attempt; a distinct source and
input set exercises the corrected nested-return path.

## Bounded recording

The large case retained 4,096 writes and 65,536 reads, with 733,376 write drops,
671,936 read drops and 8,216 frame drops. It retained 65,536 frames. Together with
drops, the frame count is 73,752: 24 root invocations and 73,728 helper invocations.
The summary remains `PARTIAL`, including 61,440 unresolved read references after
the write budget. The 24 supplied output records still matched their expectations.

Native writes use nine 32-bit integers per row and reads use five. No field value
or string payload is copied into this journal. Existing argument observations and
other runtime data retain their own formats. The native program's measured peak
RSS was about 18.1MiB in this large cohort; this excludes the compiler, test runner
and the host process retaining decoded records. Native process CPU observations
were about 27.4ms/16.6ms for the two runs. Total execution-command elapsed time was
about 827ms including build and process handling. These are local observations,
without a matched performance comparison.

## Language and source identity

Field reads refer to their latest completed lexical field binding. Completion
order is separate from reservation order because an outer assignment can reserve
its row before an inner helper completes. Copy/argument/known record-return
parents must retain the same record and field identity and an earlier completion.
Constructor fields have separate read windows; their binding slots commit after
the entire record RHS completes. Source scopes use typed Go projection ranges,
with selected Gooo local declarations retained separately.

The two actual Gooo example/reader files were registered through the existing CLI
inspection and `RegisterSyntaxCapability` operation, versions 61→62→63. Each
emitted nine artifact roles and nine unchanged member files. Denominator v62 is
preserved after v63 is generated. The corpus now declares 98 cases: 95 valid and
three invalid, partitioned into 96 language capabilities and two governance cases.
No user program or model executed during registration; the temporary generator
binary was removed after use.

## Retained data and limits

`manifest.json` distinguishes original CLI/result hashes from extracted runtime
objects. Changed-source snapshots identify the base Git tree and file hashes;
source archives exclude observations to avoid recursive copies. Original failed
diagnostics and registration requests/candidates are retained. Large observations
are compressed. Static modernization is recorded separately from native execution.

This covers required single scalar record fields, direct local copies, fixed
record-argument bindings, local record returns and explicit field RHS windows.
Scalar-local intermediates, graph-bound activity inputs, complex arguments and
returns, optional/repeated fields and general whole-program dependency coverage
remain open. Shared RHS and unknown parents keep their own counts. `COMPLETE`
describes the bounded journal within this profile, not language-wide correctness.
