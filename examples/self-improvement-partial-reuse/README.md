# Public partial test reuse

This is the canonical v15 example for semantic impact-based test reuse. It is
intentionally small but real: `CreateReceipt` is the independently testable
orders partition and `SnapshotInventory` is the independently testable
inventory partition. Both use `SharedContract`, so the canonical metadata
declares exactly two dependency edges:

```
shared-contract -> orders
shared-contract -> inventory
```

The `computes` metadata on the `.gooo` activities is the policy table. It binds
source roots, generated symbols, test units, dependency edges, the fixed six
acceptance cases, and the v14 orchestration-to-evidence journey. The Go
implementation lowers this source, derives the semantic subgraphs and actual
edges, computes impact closure, and validates immutable partition receipts.
It does not select behavior from filenames or maintain a parallel case switch.

## Source-owned assessment

The development evaluator now passes graph and receipt observations to
[`policy.gooo`](../../internal/meta/publicpartialreuse/policy.gooo). `Assess`
chooses the outcome, first unresolved cause, ordered execution mask and
required/reusable test counts.
`Refuted()` supplies a shared constant result without a dummy input. Go still
reads receipts, traverses the declared graph and observes execution results.

The policy and case report use their `v2` schemas. The case report retains the
observations and the existing fixed-value record: activity ID, source digest,
projection digest and returned value. Report verification re-evaluates those
observations against the current source and compares the reported outcome.
Missing files and invalid content have separate observations. A broken Gooo
program returns an evaluation error; it cannot become a successful outcome.

The authored assessment suite has 24 rows and compares all ten output fields.
It includes contradictory observations, first-cause ordering, invalid unused
receipts and malformed partition lists. This population is separate from the
six end-to-end cases below. The ordinary example has exactly two partitions.

Before any selective build or test, the existing driver now reads the source's
ordered mask and maps it to the observed partition identities. An unresolved or
refuted assessment exposes no selected work. Mask length and counts must agree;
changing only the Gooo rule can request both declared tests. The original full
baseline runs once on the prior source. The changed-source case reads those same
immutable receipts, executes the changed orders test, and keeps the inventory
receipt bytes and ID. It never establishes a fresh full baseline on the changed source.

The declared `independent-declaration-partitions/v1` profile checks the complete
generation record, then compares owned declaration and semantic projections.
Every binding also retains the package name, imports, shared declarations, all
comment text, exact `go.mod` bytes, test contract and selected command. Cross-owned
references and shared declarations that reference owned symbols are rejected.
The syntactic check can conservatively reject shadowed names.

The shared package reader observes the same three regular files before build,
before test, after test and before reuse sealing. It rejects extra members and
changed bytes. These boundary observations are not an atomic filesystem snapshot.
Complete SDK, transitive dependencies and external effects remain further work. The assessment alone does not establish arbitrary
dependency closure or authorize customer CI to skip tests. It uses the existing
interpreter with no model call.

The host driver measures completed Go commands through the shared operating
system resource reader on Darwin and Linux. `peak_rss_kib` is the largest
completed build/test command observation in that stage, with Darwin bytes
converted to KiB. It does not describe the compiler observer or a simultaneous
sum of processes. The driver uses no GNU `time` options or timing sidecar files;
unavailable resource observations remain errors. This change adds portability,
not a cross-platform performance comparison.

Command execution now uses the same bounded native Go observer as the whole-test
reuse driver. It resolves the executable, checks its embedded Go 1.27.2 identity,
freezes the child environment and invokes the observed absolute path. The test
execution witness retains `execution_tool_inputs`, and its result digest includes
that observation's digest. The shared reader rechecks the command after execution.
The source contract is now `v3` and names this execution tool profile. Receipt
`v3` binds the command, module and partition profile observations; the original
execution retains the complete source, semantic, generated-program and manifest
digests. The aggregate `v2` report records prior/current sources and reused receipt
IDs. Whole-program byte equality is false for the changed-source case. Before selective work, the host observes
the selected command again and presents the current bindings to Gooo. A changed
path, executable, environment or platform makes the old receipt stale. A legacy
`v1` receipt becomes `TOOL_UNOBSERVED`; Gooo returns `UNKNOWN` with the missing
command and environment origins. A `v2` receipt becomes `PARTITION_UNOBSERVED`
because the new partition input scope was not recorded. Contradictory records still take precedence.
The selected absolute command and frozen environment are shared by build and
test, rechecked before and after execution, and observed again before sealing
the reuse result. This records observed boundaries; it does not claim an atomic
filesystem snapshot or complete SDK input closure.

The v15 acceptance denominator is fixed at six cases: two `CLOSED`, two
`UNKNOWN`, and two `REFUTED`. The original v15 source was registered as one
append-only language package member. Its historical corpus increment was 58 to
59 `.gooo` sources and 1011 to 1022 physical `.gooo` lines; those values are not
the current repository totals. The assessment source above is registered
separately in the current corpus.

Selective reuse is explicit opt-in and fail-closed. A receipt is reusable only
when its canonical source and semantic subgraph, generated partition artifact,
compiler/released identity, Go 1.27.2 command, exact test contract, dependency
graph, successful result, and v14 orchestration provenance all match.

The same canonical source also carries the v16 counterexample-driven semantic
resolution repair table. A proven hidden dependency first falls back to the
full project test contract, then produces a deterministic `shared-contract ->
inventory` proposal from the preserved v15 REFUTED record. Only an explicit
human authorization permits a caller-owned immutable graph overlay; the
overlay replay proves complete affected closure and retains selectivity for
an independent unchanged partition. Ambiguous or unsupported proof evidence
is `UNKNOWN`, while tampered counterexamples and unauthorized overlays are
`REFUTED` with REFUTED dominance.
