# Typed field-origin observations

This change connects the existing typed record-flow engine to ordinary Gooo
computations and exports its graph to a Gooo ancestry program. The question is
whether a field can keep its own origin after copies, writes, helper calls and
conditional returns. Input values are absent from source inspection.

## Original local attempts

All cohorts below were new source/input sets. Completed historical cohorts were
not rerun locally. Failed attempts remain in this directory.

| Cohort | Observed result | Scope |
| --- | --- | --- |
| 12973 | Seven ancestry subtests passed; the final loop diagnostic assertion failed | Integer initializer normalization inserted an internal type before source validation. The new source validator now runs before normalization. The original cohort was retained without a local rerun. |
| 13013, within the same first runner | Producer 2/2; consumer generation stopped at the existing 4,096-call bound | The first Gooo bitset representation used nested helper expansion. No consumer native execution occurred. |
| 13051 / 13057 | Five new source assertions and explicit loop-unresolved result passed; new producer 2/2; actual Gooo ancestry outputs matched 3/3 external comparisons | Record-returning helper, literal local, independent field writes, guarded field replacement and data/control distinction. |
| 13193 / 13201 | Source node-bound and input-identity checks passed; seven new synthetic Gooo inputs matched external comparisons | 180 source writes stop at the 512-node analysis bound without a resolved prefix. Gooo accepts 1/60 nodes and returns `unusable-input` for 0/61 nodes, wrong bits, wrong order and self-edges. |

The source inspection retains scalar and per-field output roots. The first seven
passing assertions also checked original escaped helper bytes. A new optional
pre-typecheck validation hook leaves legacy record-flow generation unchanged.
Only the new source inspection enables it.

The revised column transport includes the bit corresponding to each node number.
Gooo verifies that representation and owns graph traversal and input selection.
The 4,096-call bound is unchanged. The example's 60-node set fits in a signed
integer; larger inspection reports remain available as JSON up to 512 nodes.

Across the two producer programs there are four matched logical cases. The Gooo
consumers have three source-derived and seven synthetic input rows. Their native
receipts remain input-only `OBSERVED` with finite denominator zero; the Go tests
compare the returned strings separately. These are bounded compiler observations,
not a language-wide correctness percentage. Model calls were zero.

Saved runtime elapsed observations were approximately 710ms and 745ms for the
two producer invocations, 587ms for the three-row ancestry invocation and 541ms
for the seven-row bound invocation. They include compilation and replay owned by
the package execution command; they are not inference latency or a speedup claim.

## Source registration

The actual CLI inspection and native `RegisterSyntaxCapability` operation added
the example and ancestry source at denominator versions 59→60 and 60→61.
Each emitted nine artifact roles and nine member files, copied without edits.
Observed operation times were 94ms and 92ms. No user program or model ran during
registration. Snapshot, source, Go driver, compiler and generator digests are
pinned in the retained requests. Existing denominator v60 remains byte-identical
after v61 is generated. The final corpus declares 96 cases: 93 valid and three
invalid, partitioned into 94 language capabilities and two governance observations.

## Files and retained identities

`manifest.json` records source and result hashes. `producer*.tar.gz` contains only
the changed source inputs for that cohort; the base Git commit in its adjacent
JSON identifies unchanged compiler inputs. Registration candidates are compressed
original JSON, with member content unchanged. `*-runtime.json.gz` is the runtime
object extracted from the corresponding original CLI receipt; hashes distinguish
the raw CLI bytes and the extracted representation. The first failed log is a
public path-labelled excerpt; its original private log hash is retained.

This feature covers static source may-dependence of required scalar fields.
Runtime field-sensitive local histories, optional/repeated-field analysis,
loop-carried origins and broader graph consumers remain open.

Static Go 1.27.2 modernization initially proposed `slices.Backward` in the test oracle. That equivalent reverse iteration was applied; the next static source snapshot had empty diff and stderr. No completed source/native cohort was rerun after that style edit.
