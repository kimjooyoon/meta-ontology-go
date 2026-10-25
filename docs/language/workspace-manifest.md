# Gooo workspace package graph

## Construct imported functions using caller feedback

The development command `gooo package construct --construction-cases feedback.json
--cases evaluation.json --attempts 8 gooo.workspace.json` keeps source assembly
choices available after package lowering. It can reconsider imported helpers
whose local examples pass but whose combined caller result differs. The
[runnable example](../../examples/package-caller-construction/README.md) includes
native faults, partial budgets, exact integers and an optional compact model.

Use `--json` to save the full `gooo/workspace-caller-construction-receipt/v1`
record. `package construct --receipt saved.json --cases evaluation.json
gooo.workspace.json` binds the current package image and original construction
cases, then re-executes every saved attempt without inference. Evaluation is
performed after selection. Its caller input overlap counts stay separate from
source-local examples and model-training exposure. `COMPLETE_FINITE` requires
complete local and caller construction observations and matching evaluation
expectations without observed activity faults or blocked dependencies.

The development source accepts `--inputs inputs.json` in place of `--cases`,
for both fresh construction and saved receipt replay. The input file uses
`gooo/body-composition-inputs/v1` and supplies no expected values. Fresh construction
still requires labelled `--construction-cases`; actual inputs do not choose the
program. Plain output is one JSON entry value per row. With `--json`, the receipt
has `decision=OBSERVED`, `inputs_digest`, and zero finite evaluation counts; the
original construction decision and all attempts remain in `result.construction`.
An observed receipt can later receive new inputs or labelled evaluation cases.
Native faults remain in JSON observations. Plain output returns a nonzero exit
when an entry value is unavailable, retaining any preceding successful rows.

## Reuse a selected package program

Save `package execute --json` output, then use `package replay --receipt <saved.json>`
with the workspace manifest and either `--cases` or `--inputs`. The
[diagnostic tool example](../../examples/package-diagnostic-replay/README.md)
shows a Gooo library whose three body choices can be selected by the own compact
model and subsequently reused without the model files.

Replay rereads the manifest and its current sources, reconstructs the dependency
graph, source fills and external fill plans, and verifies the saved native
projection. It compiles and runs that program twice on the requested inputs.
The selected body stays fixed; changed input rows do not initiate another search.
Copying the same manifest-relative source tree to another location is supported.
Changed declarations, dependency identities or entry points need a new execution
record. Older external-fill records that lack `external_plan` need one new
`package execute` using their original plan; source-owned records contain their
construction contract already.

With `--json`, `replayed_from_sha256` identifies the exact supplied receipt and
`result.replay.prior_result_sha256` binds its decoded result. Historical generation
records retain their old model-call counts. This invocation's replay and runtime
model-call counts are zero; provider environment settings are not consulted.
Fresh case results report `PASS` or `PROGRESS`; input-only results report
`OBSERVED` and keep correctness unknown. These decisions describe execution and
finite expectations, not correctness across all inputs. Prior runtime observations
are retained in the input receipt, rather than reused as current measurements.

Replay also recomputes `runtime.input_separation` over the fresh native traces.
Earlier source fills and external fill plans contribute their construction,
holdout and probe inputs. The retained `earlier_stages` identities explain which
sets were compared. A new root input that reaches an already observed value at
a later activity remains overlapping. [Whole-workspace example](../../examples/workspace-input-observations/README.md).

## Resolve and execute

`gooo.workspace.json` connects source files to the package graph already used
by Gooo's package runtime. Resolve a workspace with:

```sh
gooo package resolve gooo.workspace.json
gooo package resolve --json gooo.workspace.json
gooo package execute --json --cases examples/package-imports/cases.json \
  --body-plans examples/package-imports/body-plans.json \
  examples/package-imports/gooo.workspace.json
```

The cross-package binding example at `examples/package-imports` uses this
command directly. The workspace reader accepts regular `.gooo` sources and
`.gooo.fixture` files for runnable examples that are not part of the language
conformance corpus:

```sh
gooo package resolve --json examples/package-imports/gooo.workspace.json
```

Without `--json`, resolution prints each package's imports and exported
entities and activities, then lists checked imported-activity bindings. This
view helps inspect a workspace before consuming its JSON receipt. Both forms
resolve the package graph only; they do not run activity bodies.

Version 1 records an entry activity and packages. Each package has a stable path
and source paths relative to the workspace manifest. The 0.6.11 development
source reads omitted `name` and `imports` from the Gooo files. For example:

```json
{
  "schema": "gooo/package-workspace-manifest/v1",
  "entry": {"package_path": "boundedint", "activity": "Clamp"},
  "packages": [
    {
      "path": "boundedint",
      "sources": ["main.gooo"]
    }
  ]
}
```

Source files can state their package dependencies directly:

```gooo
package app
namespace app
import "example/core"
```

An alias can name a package in an explicit activity binding:

```gooo
import core "example/core"
bind core.Normalize.result -> Main.input
```

Resolution reads the listed `.gooo` sources, builds a deterministic dependency
order, and emits a `gooo/package-runtime-result/v1` receipt with source and
semantic digests. Unknown package imports and dependency cycles fail closed.
The JSON receipt also reports the selected entry and the current zero-effect
boundary.

Each package result now includes its statically resolved `exports` list.
Entities are exported with their stable IDs; activity exports bind every input
and output entity type to that ID. An activity can use an entity in
its own package or an entity declared in one of its direct imports. Unknown or
ambiguous imported types fail closed. When an entity exists locally, the local
declaration takes precedence over imported entities with the same name.

Record names are scoped to the owning package during executable lowering.
Different packages may each declare `Result` with distinct stable IDs and field
layouts; the compiler assigns distinct execution names and records the mapping
in `program.entity_aliases`. Constructors, signatures and supported assembly
expressions use the resolved type. Names inside strings, comments, field labels
and ordinary local-variable references retain their original spelling. Aliases
of the same stable record ID share one declaration when their ordered fields
agree. The [two-Result workspace](../../examples/package-record-namespaces/README.md)
exercises this behavior through real package execution and exact integer values.
Ambiguous imported constructor names need an unambiguous declaration/import
environment, as do signature types. Scalar entities keep the current
Integer/Boolean/Text body profiles.
Omit `name` to use the `package` declaration; every source in that package must
agree. Omit `imports` to use the sorted, deduplicated union of all its source
imports. JSON `null` has the same meaning as omission for these optional fields;
an empty name also requests source metadata. An explicit `imports: []` asserts
no imports and is checked against any source import declarations. An explicit
nonempty name or import list is also checked. Import-list differences fail with
`PACKAGE_SOURCE_IMPORT_MISMATCH`; differing source package names fail with
`PACKAGE_HEADER_MISMATCH`. Existing manifest-only dependency declarations remain
supported when the source has no imports.

`resolve`, `execute`, `replay`, `resume`, `construct` and Gooo assembly-policy workspaces use
the same reader. Omitted and equivalent explicit metadata produce the same
normalized graph, source/semantic identities and generated program. Each original
manifest retains its own byte digest. Dependency cycles, unresolved packages and
invalid source still produce errors. Package paths, source locations and the entry
remain explicit; this operation does not discover or download other repositories.
The [source-derived three-package example](../../examples/package-body-calls/README.md)
uses real calls, body assembly and saved execution. New library and diagnostic
starters use this form. A compiler built from this development revision is needed;
the published 0.6.10 compiler uses the earlier, fully specified form.

Gooo source imports identify package dependencies, while the workspace manifest
still supplies package paths, source files, and the executable entry. Activity
calls such as `rules.IsPartial(input0, input1)` are available in the supported
pure-body profiles. `package resolve` can also resolve an explicit imported-activity binding against the imported
package's exported activity and the local consumer. The output and input must
carry the same stable entity ID. Its receipt records producer and consumer
packages, activities, ports, and entity ID under the consumer package's
`bindings` list. Unknown aliases, activities, ports, and type mismatches fail
closed. Imported producers can feed local consumers. `package resolve` only
checks that wiring. `package execute` takes a finite case file or an input-only
document, lowers the entry's explicit producer chain into a typed activity graph, generates each
body in that chain, compiles the resulting Go, and runs it twice. Unrelated
activities in the workspace are left out of the execution graph. Case keys use
`<package-path>:<activity>`;
root inputs use that key, multi-input roots append `.<input-port>`, and expected
outputs use the activity key. The receipt preserves package-to-lowered-activity
identity and reports observed finite-case accuracy when expectations are supplied.
This path supports 1 to 16 activities: the entry runs independently, or its producers
join through explicit binds. It accepts supported
scalar or declared-record values. Ordinary bodies follow their declared Gooo
bodies deterministically; `--assembly-model` is an optional local model for assembling source-declared
body plans. A source-owned `source_fill` plan in an activity's Gooo
`assembling` block supplies typed body-hole candidates directly from source; an
optional `--body-plans` file remains available for activities without such a
declaration. An activity cannot receive both plan forms. Gooo scores each
complete candidate against its declared cases, then Laya may select among
those listed candidates through `GOOO_LAYA_URL`; with no provider configured,
selection stays deterministic. Gooo fills the chosen body before package code
generation and execution, and the receipt records each fill's input-source
digest, selected candidate, and finite-case score. `--tiny-model <model.json>`
loads the local compact model once and uses it for sequential fills; for a
multi-hole plan, its current mapping is limited to distinct supported root
operations in each candidate's first declared hole. This option cannot be
mixed with Laya endpoint or credential settings. Without a configured
provider, selection remains deterministic. The model call runs synchronously after the typed plan is
built and before the generated package is compiled. To connect Laya, set
`GOOO_LAYA_URL` to its `/v1/systemone` endpoint for the same command. Fill
receipts record provider decision time; native build and run observations
record wall time, CPU time, and peak resident memory when the host exposes them.

`--assembly-policy-workspace <policy.workspace.json>` connects a separate Gooo
policy package to record-choice construction. Its fixed entry consumes the typed
assembly-observation record after each scored attempt and returns a supported
next operation. It may call imported pure helpers. The result retains the policy's
source manifest and lowered program in `assembly_policy`; saved replay reconstructs
them and the decisions without opening policy or model files. See the
[connected package example](../../examples/package-assembly-policy/README.md)
for deterministic/model construction, partial stopping and replay. Candidate
permissions and budgets remain those declared by the target source.

For an ordinary language tool, use `--inputs <inputs.json>` instead of `--cases`.
The input document has schema `gooo/body-composition-inputs/v1` and an `inputs`
array of named root-input maps. No expected output is supplied. Plain output is
one JSON entry result per row; `--json` includes the complete receipt with
`decision: OBSERVED` and `inputs_digest`. Two matching native executions establish
replay, while finite expectation counts remain 0/0 and correctness remains
unobserved. The existing `--cases` mode still requires nonempty expectations.
See the [standalone Gooo assembly explainer](../../examples/assembly-explainer/README.md).

Declared records are parsed with the explicit EntityFields V4 profile throughout
package resolution, IR lowering, and workspace flattening. V4 preserves optional
single string, Boolean and Integer fields across workspace binds: omission stays
absent, while explicit `""`, `false` and `0` stay present. `null` remains invalid.
The runnable [optional-record workspace](../../examples/package-optional-record-flow/README.md)
checks absent fields, explicit zero values and a large exact integer through the
full package execution path. For example,
`examples/package-record-flow` sends a `Candidate` record from an imported
`Submit` activity into an app-owned body that constructs a typed `Review` record.
Run it with:

```sh
go run ./cmd/gooo package execute --json \
  --cases examples/package-record-flow/cases.json \
  examples/package-record-flow/gooo.workspace.json
```

Finite cases check the actual record values at both activities, then runtime
replay checks that the generated Go projection produces the same result again.
EntityFields V3 currently covers required single string, boolean, and integer
fields in the original record-flow example. The workspace runtime's V4 profile
also supports optional single fields of those types. Repeated, nested, and
arbitrary user-defined field types remain outside this support boundary.

## Inspect a source before choosing a split

An agent can ask the existing split command for declaration boundaries:

```sh
go run ./cmd/gooo package split-source --boundaries --json \
  --package example/agent-tool-inventory --source main.gooo \
  examples/agent-tool-inventory/gooo.workspace.json
```

The `OBSERVED` receipt binds the manifest, original source bytes and original
workspace image. Candidates are ordered by physical declaration position and
carry a resolved `before_id`, display name, byte/line address and both output
sizes. Counts include comments, blank lines and the header copied into the new
file. The current example has 82 physical lines; the cut before `Inspect`
produces 44 and 41 lines. The extra three lines are the copied header.

The caller chooses a candidate according to its own source-owned constraints,
then passes its ID to the ordinary `--before ... --into ...` proposal command.
Use explicit `--consumer-workspace` arguments or `--discover-consumers` there
when other workspaces read the same source. Applying the resulting proposal
rechecks current inputs and every declared consumer. The regression exercises
this complete path on the example and checks its eight existing expected outputs.

Boundary inspection allocates compact size observations rather than a complete
source copy per candidate. It validates the original workspace once, then scans
the selected source's declarations. It does not execute bodies, choose a model
decision or establish that a candidate meets the caller's size limit. A single
declaration produces an empty candidate list. Missing inputs and invalid source
produce no candidate prefix.

`semantic_admission` remains `UNASSESSED`: destinations, changed consumer images,
source-set limits and repository registration have not been evaluated. An
observation cannot be passed to `--apply`, even if its decision is edited. Native
callers can use `packageruntime.InspectSourceSplit` with the same boundary
semantics. No new CI job or persistent service is required.

### Choose a cut from a physical line target

An agent can supply a target instead of selecting a declaration ID itself:

```sh
go run ./cmd/gooo package split-source --line-limit 75 --json \
  --package example/agent-tool-inventory --source main.gooo --into inspect.gooo \
  --syntax-registry-root . \
  examples/agent-tool-inventory/gooo.workspace.json > /tmp/inventory-split.json
go run ./cmd/gooo package split-source --apply /tmp/inventory-split.json --json \
  examples/agent-tool-inventory/gooo.workspace.json
```

The selection minimizes the larger output's physical line count among existing
declaration boundaries that strictly reduce the original file's size. Equal
results use the stable declaration ID as a deterministic tie break. Selection
uses the compact boundary observations; it constructs and admits only the chosen
partition through the ordinary workspace and registration path. If that consumer
admission fails, it returns the failure without trying an unreported alternative.

The optional `selection` object records the original source and workspace image,
candidate count, improving candidate count, count meeting the target, and selected
boundary. For this example it selects `Inspect`: 82 lines become 44 and 41,
including the duplicated header. A target of 30 still proposes the same useful
intermediate split with `limit_satisfied: false`; it does not claim the target has
been reached. A source already within the target returns a successful `UNCHANGED`
result. A source without an improving cut returns `NO_IMPROVING_CUT`; neither
contains a writable proposal. `--line-limit` and `--before`
are exclusive. Saved selections are recomputed from current inputs on application.

Native callers can use `packageruntime.ChooseSourceSplit`. This is a deterministic
physical-size objective, not a judgment of conceptual cohesion or a model choice.
The metric executor still needs an explicit workspace consumer scope before it
can use this path; a file-length observation alone does not identify that scope.

## Native workspace consumers

Repository Go consumers can use
`internal/packageruntime/workspaceinput.Decode` and `Load` for the same definition
schema, source paths, file/count/byte limits and source-derived package metadata
as the CLI. The external definition remains a source list; the returned
`packageruntime.Manifest` contains the actual file bytes for semantic resolution.
The host supplies `ReadFile` and owns filesystem access and file identities.
Within one `Load`, each canonical source path is read once and its immutable
content is shared by every declaring package. Every declared reference still
counts toward the source-count and byte limits. A subsequent `Load` reads current
bytes again; this observation scope does not establish a filesystem-wide snapshot.
If a file fails to load, the reader returns an error and no executable partial
manifest. The original definition still names the full requested source set.

This shared reader lets a metaprogram inspect and partition a package without
importing the CLI or introducing a second interpretation of its manifest. A
native consumer regression loads imported activities, splits at a stable ID,
and executes both workspaces against the same independent expected values.
Reading the definition is not execution or rewrite authority; consumers still
validate the semantic image, current inputs and the operation's own preconditions.

`workspaceinput.PrepareSplit(reader, manifestPath, definition, request)` also
prepares the authored source list and normalized source partition together.
`SplitRequest` names a package, source, stable boundary ID and destination.
For the existing [named-inputs program](../../examples/named-inputs/main.gooo),
the boundary is `urn:gooo:example:named-inputs:net` and the destination can be
`net.gooo`. The native caller uses the same operation as `package split-source`.

The returned `Definition` appends the new source to every package that declares
the same canonical source address, retaining authored names, imports and entry.
`Result` carries the exact output bytes, byte origins and
semantic image already verified by the shared native partition operation. Both outputs
are needed to install a proposal. Preparation leaves the input definition and
reader unchanged; returned file lists do not alias caller-owned storage. Output
source-set limits are checked before returning an installable proposal.

The native regression serializes that authored definition, reloads the proposed
files through the shared reader, checks the planned image, and executes both
old and reloaded programs against independent expectations. Hosts still own
current-input checks and filesystem writes. This API does not discover other
single-file/embed consumers or establish repository-wide rewrite coverage.

The requested source path uses the loader's canonical resolution, while authored
paths such as `./main.gooo` remain unchanged in the returned definition. Use
`workspaceinput.Encode(proposal.Definition)` to serialize it. The encoder retains
the existing indented JSON plus final newline, checks the same 1-MiB reader limit,
and rejects any JSON round-trip that changes authored values. `PrepareSplit`
performs this check before returning a proposal; the CLI writer and installed
manifest check use the same encoder. A valid source partition cannot leave an
output file list that the shared reader refuses to reload.

When several packages share the source within one definition, the native operation
uses `packageruntime.SplitSharedSource`. It verifies every package's declaration
interface before and after the common byte partition. A multi-package result uses
`gooo/package-source-split/v2` and names all `affected_packages`; a single-package
result retains the existing v1 representation. The saved CLI proposal includes this
scope, and applying an old incomplete shared-source proposal is refused before writes.
The ordinary `packageruntime.SplitSource` API retains package-local virtual filename
semantics. Shared-source interpretation is explicit at the workspace boundary.

The regression reloads the proposed external definition, requires its exact planned
image, then executes the other package's entry through imported calls and bindings.
Its four expected activity outputs and runtime replay retain their own denominator.
Validity of the reloaded program alone cannot establish preservation: the original
counterexample was valid while the other package's moved declaration had disappeared.

Before publishing a saved split, the CLI reads the planned file replacements
through the current filesystem addresses and reuses native image verification.
This catches directory aliases and case-insensitive aliases that would otherwise
change an unplanned package only after writes. The check runs after staging and
the current-input recheck, including a same-byte directory-alias retarget.
An unplanned effect retains the existing files and reports `PRE_STATE_RETAINED`;
authoring canonical source paths lets the shared-package plan include those consumers.

A hard link is a separate directory entry: replacing one entry keeps the other's
old bytes. Equal content or inode alone therefore does not couple planned writes.
Ambiguous entry resolution is reported before publication. Alias resolution keeps
two name observations and reads directory entries in batches of 64, rather than
retaining a full directory listing. This extra verification runs on explicit split
application; ordinary package execution gains no new model or projection phase.
The host's existing single-writer boundary still applies, including external edits
after the last check and process interruption during publication.

### One source with multiple workspace consumers

`--discover-consumers <directory>` can supply the group from recursive `.json`
workspace definitions under a declared root. The saved proposal retains that
scope and all observed workspace definition hashes. Apply rescans before writes;
a newly introduced consumer or changed definition makes the old plan stale.
Malformed/unreadable candidates and unknown workspace versions refuse discovery.
This is scoped source-list discovery; embedded or dynamic readers and other roots
remain separate obligations. See the [runnable example and limits](../../examples/shared-source-consumers/README.md).

Changing a shared source and only one source list can leave another workspace
without its entry activity. `internal/packageruntime/workspacesplit.Prepare`
constructs one joint draft for explicitly supplied `Owner` values. Each owner
names its external definition address, authored definition and split request;
the existing Gooo compiler verifies each workspace's semantic partition.
The result contains two shared source writes and each updated definition with
its encoded JSON and separate package-image/source-origin records. Different
entry activities remain different consumer contexts.

All owners must agree on the source and destination addresses, partition bytes
and original byte ranges. Identical content at different addresses does not
identify the same occurrence. Missing dependencies, conflicting destinations,
duplicate definition addresses or changed shared bytes return an error and no
partial plan. Preparation writes nothing and retains authored source spelling.
It accepts at most 64 owners and bounds the aggregate proposed source payload
by the shared reader's 16-MiB source-set limit, including repeated consumer views.

Addresses are lexically normalized in the host reader's namespace. The host
still owns filesystem alias/symlink identity, discovery of every affected
consumer, current-input checks and atomic publication of the whole draft.
Omitted workspaces and single-file/embed consumers remain obligations. The API
provides no repository-wide coverage or execution authority. The joint native
regression reloads both updated definitions and checks six expected activity
outputs across their different entries, including runtime replay without a model.
The same output conditions run with one package per workspace and with an
additional package consuming the source inside the first workspace; both external
definitions must still reload to their planned images.
The executable input is an [authored Gooo source](../../examples/shared-source-consumers/main.gooo);
the Go tests supply the consumer definitions and independent expected results.
The syntax registration constructor generated all nine registration members from
this source and appended denominator v101 (136 to 137 syntax cases). That inventory
count is separate from the six independently expected execution outputs.

The joint plan also returns `Inputs`: observed source addresses, SHA-256 digests,
byte lengths and the definitions that read each occurrence, sorted deterministically.
Every supplied consumer still reads its inputs; a second read of any shared source
with different bytes returns `ErrSourceChanged`, including unchanged dependency
files. Equal contents at distinct addresses remain separate observations. Definition
outputs cannot overwrite any observed source address. The list records source
reads; `Prepare` receives decoded definitions and does not observe their serialized
bytes. The host checks current definition and source inputs again before
publication. Only digest/consumer metadata is retained by the source observer.

`PrepareFiles` accepts `FileOwner` requests naming the external definitions and
split choices. It reads and decodes those definitions through the same bounded
workspace contract, then calls the joint constructor. `Definitions` records each
input address, raw-byte SHA-256 and byte length separately from the source reads.
Definition inputs are limited to 1 MiB each and 16 MiB in aggregate, independently
of the source-set limit. They are reread after preparation; changed bytes, even
formatting-only changes, return `ErrDefinitionChanged` and no plan. Failed or
malformed reads likewise return no partial proposal. The original serialized
inputs are not kept as an additional payload.

These sequential reads are observations, not a filesystem snapshot or publication
permission. Current-input checks, alias resolution and complete consumer discovery
still belong to the applying host. The native execution regression now begins at
this file-backed entry point and retains the same six expected activity outputs.

### Embedded native contract sources

Native declaration contracts have no executable entry point. The internal
`sourcebundle` adapter reads an explicit ordered set of `.gooo` members, parses
each with its physical filename, and lowers their declarations through the
existing bidirectional and semantic validators. Every member must have the same
package and namespace. Imports require the package compiler and are refused here.
This adapter does not replace the workspace compiler or discover consumers.

The operation-input contract is its first connected consumer. Its embedded family
is `operation-input-contract*.gooo`; partitions within that family are read in
lexical filename order. The existing native checks still require exactly four
operations, six obligation activities, two policy activities and ten entities.
Adding a file cannot add an unimplemented executor or bypass those checks.

The bundle identity binds ordered filenames and exact bytes. It differs from the
old single-file content hash even for one member, so old operation plans need
fresh preparation. An extraction report with matching meaning but a different
contract source reports `EXTRACTION_CONTRACT_SOURCE_MISMATCH`, directing the caller
to prepare a new extraction. Historical observations retain their original identity.
Semantic identity remains separate: moving a declaration to a
different member preserves its stable ID and semantic hash while changing its
physical origin. Admission exposes each member's content hash and byte length,
plus all declaration IDs and physical spans; cached metadata is copied for callers.
Existing serialized operation plans continue to carry the contract digests.

The adapter also retains the physical origin of first-class `policy` declarations,
which the semantic IR stores separately from the entity/activity graph. The public
policy-compilation example is checked across a filename change: its meaning stays
the same while its source identity and physical origin change. This general source
support does not admit extra policies into the exact native operation ABI.

The connected regression partitions the actual authored contract into ten entities
and twelve activities, admits it through the native consumer, and checks all 22
physical origins and unchanged native facts. Missing members, conflicting IDs,
changed obligations, extra activities and unknown import scopes return no contract.
Reads are bounded to 64 files, 1 MiB per file and 16 MiB in total; an I/O failure
returns no partial source set. Mutable filesystem snapshots and publication remain
the caller's responsibility. The production contract remains one small source file;
repository-wide discovery and automatic embedded-consumer migration remain open.

### Native entity identity across display renames

The native operation-input contract binds its ten entity roles by explicit stable
ID. Authored names remain visible in extraction reports. For example, the compiler
now uses this declaration with its original identity:

```gooo
entity FunctionSource id "gooo://meta-operation-input-contract/entity/function-input"
activity ExtractFunction(FunctionSource) -> OperationResult
```

Changing `FunctionInput` to `FunctionSource` preserves the semantic hash and native
input role. It changes the source digest, so a previous plan still needs fresh
preparation. Extraction reports carry both the authored `input_entity` and its
`input_entity_id` through obligations, proof stages, preflight observations and
preparation progress. The decoder checks both against the current source contract;
proof stages also retain the previous stage's output identity as their input.
Historical reports retain their original labels and source identity.

All ten entity renames are covered at admission. Same labels with swapped,
missing, duplicated or unknown IDs are rejected with no partial contract. Actual
map-construction extraction and report decoding exercise the production rename,
including rejection of altered type IDs and a disconnected proof edge. Source
labels are not aliases for the host ABI vocabulary. Activity renames, executor
discovery and publication of a repository-wide split remain separate work.

### Observing Go embed consumers

The native `sourcebundle/embeduse.Inspect` API observes one Go package and an
optional proposed set of Gooo source members. It uses Go's own `list -test` and
`-overlay` selection. It performs no source installation, test execution, model
call or native ABI admission. The selected Go build context is retained alongside
the requested directory and the physical directory used by the tool.

A package's `EmbedFiles` is a union. Suppose one `embed.FS` variable reads
`contract*.gooo`, while another `[]byte` variable reads `contract.gooo`. After a
split, the union can include both parts while the second variable still receives
only the original filename. The observer attributes selected members to each
variable declaration. It keeps all proposed members visible, including those
selected by no observed variable.

Each query temporarily masks sibling embed directives, preserving offsets in the
query source, and lets Go resolve the remaining directive. Directory patterns,
`all:`, quoted names and test contexts use Go's resolver. Reports refer to the
original Go file hash and physical declaration span, including when a `//line`
directive supplies a different logical filename. Query overlays never become
authored source origins. Go source and Gooo member records share one flat file
table; consumer rows refer to its indices. Projected bytes are marked separately
from bytes read from disk.

The counterexample fixture compares all three contexts: production, internal
tests and external tests. Its independent Gooo reader finds three declarations
in the combined partition and only one in the single-file view, although both
views are valid inputs. A separate native probe of the actual operation-input
contract admits two projected members and preserves all 22 declaration origins,
the semantic hash and six obligations while changing source identity.

This is selected-context discovery of static embed bindings. Dynamic `ReadFile`
calls, other build contexts and the consumer's later interpretation remain
unobserved. Membership does not establish that a native consumer accepts a changed
contract. Source bytes and package selection are reread before return; publication
still requires its own current-input checks. Existing overlays supplied through
`GOFLAGS` are refused rather than silently replaced. The observation is bounded to
64 variables, 1,024 input files, 1 MiB per file, 16 MiB total and a 30-second query
context. Failure returns no report, and owned query files are removed.

### Planning a partition against each embedded source view

`sourcebundle.PreparePartition(sources, filename, beforeID, into)` resolves a
stable declaration ID in the original source set and proposes two source members.
It shares the byte partitioner used by `package split-source`: original bodies,
comments, line endings and bindings are copied with explicit input/output byte
ranges. The package headers are duplicated. Both source sets are lowered and
compared; a new file changes source identity while preserving the semantic image.
Policy declaration IDs and their physical spans participate in this preparation.

`embeduse.PreparePartition(ctx, directory, names, filename, beforeID, into)` adds
the selected Go package's consumer observations to that proposal. `names` is the
explicit original source set, and every member must have an observed input in
that package. The destination must be absent. Each consumer that read the original
member or starts selecting the new member receives a result:

| State | Meaning in this observation |
| --- | --- |
| `PRESERVED` | Both selected source sets lower to the same semantic image. |
| `CHANGED` | Both lower, but their images differ; missing and new declaration IDs remain visible. |
| `UNASSESSED` | At least one selected source set cannot be lowered in this source-bundle scope; the reason and any available image remain visible. |

For example, the single-file consumer in the earlier counterexample still parses
successfully after the split. Its plan is `CHANGED` because two declaration IDs
are missing. A wildcard consumer of the same file retains all three declarations.
A previously unrelated wildcard that starts selecting the new member also appears
in the plan. Mixed-package source sets, unresolved imports and incomplete fragments
remain `UNASSESSED`; a lowerer error does not become an empty successful image.

The actual operation-input contract regression takes the planned members selected
by each variable into the native ABI reader. The family retains all 22 declaration
origins, four operations, six obligations and two policy activities. The single-file
reader loses twelve activities and is refused by the ABI reader. Image comparison
and native admission are distinct checks. The shared workspace regression continues
to execute and replay six independent expected activity outputs after partitioning.

Planning rereads original inputs, binds proposed bytes to the after-observation,
and rejects changed Go declarations, build context or unplanned source members.
The complete call has a 30-second context, with the existing input and consumer
bounds. A host error returns no plan; an assessable change or unsupported consumer
scope stays in the returned plan. These observations are sequential reads, not an
immutable filesystem snapshot. No files are written. Native ABI acceptance,
dynamic readers, other packages/build contexts, current-input checks at publication
and coordinated installation retain their own obligations. This native planning
API adds no standalone CLI and does not grant the existing workspace apply command
repository-wide consumer coverage.

### Apply one source split across explicit workspace consumers

`gooo package split-source` accepts repeatable `--consumer-workspace <path>`
options during proposal generation. Each additional manifest must declare the
same source address inside its workspace root. The common shared-source planner
checks every owner and retains its independent package image and original
manifest digest. The saved receipt carries those additional views in `consumers`.

`--apply <proposal.json>` rechecks that declared group and stages the two source
members plus every manifest in one write set. It verifies all proposed images
through the existing loader before publication, uses the existing rollback
writer, and rereads all installed owners afterward. Stale dependencies in any
owner refuse the group; an old proposal cannot be applied twice. The primary
`application` state describes the whole supplied group. Single-workspace receipts
and their compressed form remain supported.

The [two-entry example](../../examples/shared-source-consumers/README.md) executes
both installed workspaces and checks six expected activity outputs. This command
does not infer omitted consumers or edit embedding/registration declarations.
The saved paths, per-owner images and source byte origins define its scope.
Filesystem rechecks are sequential, and rollback does not imply crash-atomic
visibility across multiple files.

### Predicting source registration changes before installation

The native `languagesyntax.InspectPartitionInventory(repository, registryRaw,
directory, proposal)` API accepts that same `sourcebundle.Partition`. It replays
the proposal against current source members and compares the repository's current
and proposed file observations with the existing syntax corpus registry. For
example, splitting `NextObservation` from `source_recovery.gooo` into
`source_observation.gooo` preserves the activity IDs but introduces a file that
the registry does not yet list. `AddedUnregistered` names that file before a write.

`Before` and `Projected` are the existing source-inventory shape, with separate
observation digests and the same registry digest. Existing missing and unregistered
paths stay visible in both views; the added-path list is only their difference.
The projection uses the ordinary observer's byte digest, line count and directory
traversal order. Tests compare it with an independently populated in-memory tree,
including the compiler's actual two-member continuation policy.

Selected members are bounded by the source-bundle reader and rechecked after the
repository walk. A changed member, altered proposal, occupied destination or read
failure returns no inventory. The complete repository walk has the same scope
and cost as `InspectSourceInventory`; it is not an incremental observation or an
atomic filesystem snapshot. The caller owns global snapshot freshness.

This API observes one additional consumer of file organization. It installs no
files, changes no registration and runs no native activity. Semantic admission
remains `UNASSESSED` in both inventories. Embedded-reader preservation and native
ABI acceptance still require their existing checks. There is no new CLI, model
call or mandatory CI job.

### Preparing registration alongside a physical partition

`languagesyntax.PreparePartitionRegistration` prepares the syntax registry side
of a source partition. It accepts the current corpus and exact partition, extends
the affected `case_sources`, `package_units[].members` and `meta_sources`, and
returns the projected registry together with its exact formatted `Corpus` bytes.
The projected inventory digest refers to those bytes. Case IDs, definitions,
selected profiles, package entries and capability denominators stay unchanged.

For example, separating `NextObservation` from the compiler's recovery policy
extends the existing continuation case from two physical members to three. Its
whole-case replay still sees the same activities. Separating the activity in
`examples/entity-fields-v1/main.gooo` creates a two-member group for the same
case, rather than registering another language capability.

Each affected case is replayed before and after under its existing profile.
Package membership receives a separately labelled `package-syntax` replay; that
observation does not execute its entry. Additional siblings are captured from
the observed inventory and rechecked before returning. Changes to sibling bytes,
an occupied or already registered destination, or unsupported projection-consumer
membership yield no plan. Unrelated missing and unregistered files remain visible.

Field-bearing sources use `sourcebundle.PreparePartitionWithEntityFieldsSupport`.
The exact selected profile travels through parsing, lowering, the physical cut
and proposal replay. The ordinary path keeps its deferred default, and consumers
retain their own profiles even if a partition was prepared under a newer one.

This is an internal preparation API. Native Go registry edits, embedded readers,
external workspace consumers and joint publication require their own bindings.
The workspace command described below joins the supported registry and workspace
consumers. The returned corpus alone cannot be installed as an accepted registry.
Inventory admission remains `UNASSESSED`, with zero repository writes and native
executions. The practical milestone is one source-owned change plan carrying all
affected consumers into the existing publisher, without hand-maintained follow-up
edits or a new capability for each file fragment.

The next construction boundary is
`syntaxregistration.PreparePartitionRegistrationCandidate(repository, directory,
partition)`. Its `Outputs` contains the two exact Gooo fragments, the projected
corpus, and every Go registry source that owns changed membership. `Inputs`
retains the current corpus, the native registry package and all affected source
members, including unchanged siblings. The continuation example produces five
files because its family and meta-source memberships have separate Go owners;
the single-case, meta-only and package examples each produce four.

Native owners are resolved through the registry's finite expressions and physical
declaration spans. Renaming an owner file does not change that binding. Direct
slices, ordered helper assembly and explicit local row declarations are supported;
arbitrary computation and reassignment remain unresolved. The existing case
definitions, package entries, projection consumers and schema must match the
corpus before construction. Re-reading the emitted Go checks the complete
projected registry. Independent tests compile that Go and compare the resulting
registry with the corpus, including a deliberately wrong expected value.

Preparation rechecks the native file population and captured bytes. Validation
reconstructs the whole candidate, so changing an output or an input invalidates
the saved result. Each of the native and Gooo input sets is bounded to 1 MiB per
file and 16 MiB total, with at most 1,024 observed files. Neither function writes
the repository. The internal native edit label describes mechanical construction;
binding the existing `SplitGoooSections` operation, admitting embedded/native
consumers and publishing the joint write set remain caller obligations.

### Publishing a workspace split with syntax registration

For a source-only package registered by this compiler repository, the existing
workspace command can construct and install one combined change:

```sh
go run ./cmd/gooo package split-source \
  --package example/agent-tool-inventory --source main.gooo \
  --before urn:gooo:agent-tool-inventory:inspect --into inspect.gooo \
  --syntax-registry-root . --json \
  examples/agent-tool-inventory/gooo.workspace.json > /tmp/inventory-split.json

go run ./cmd/gooo package split-source --apply /tmp/inventory-split.json --json \
  examples/agent-tool-inventory/gooo.workspace.json
```

The first command prepares five file changes: `main.gooo`, `inspect.gooo`, the
workspace source list, the syntax corpus, and the native registry owner. It writes
only the redirected proposal. The second command reconstructs the proposal from
current inputs and publishes through the existing staged writer. Source writes
shared by multiple consumers occur once; explicit `--consumer-workspace` or
`--discover-consumers` options add their manifest edits to the same transaction.

The proposal includes each registry output's exact text and digest, input file
digests, the selected workspace profile, and the existing Gooo `SplitGoooSections`
operation's lowered activity
identity, physical origins and `used`/`wasGeneratedBy` contract facts. The compiler
checks its embedded operation contract against the repository source family.
Its identity comes from the semantic graph, including an explicitly authored ID
when present. The two source fragments retain their existing wire location and
are not duplicated in the registration payload. The mechanical edit implementation
remains Go. This command does
not add a route to the repository metric-driven execution scheduler.

Workspace parsing and lowering use the same existing field-profile selection as
package execution. Registered syntax consumers continue to replay under their
own declared profiles. These independent consumers must agree on the exact two
source fragments; a physical split adds no capability case.

Before publication, all captured inputs and output preconditions are rechecked,
including unchanged native registry files and operation source members. Failures
during installation use the existing rollback path. Tests inject failures at
the corpus/native writes and directory synchronization and verify all five
pre-states. A native input changed during staging prevents any rename and retains
that independent edit. Successful installation is checked against the proposed
file digests and package image. The installed inventory example passes its eight
existing runtime cases without model calls. Sequential rechecks and rollback do
not provide a concurrent filesystem snapshot or crash-atomic multi-file commit.

The optional registry scope is explicit and must contain the workspace. Import
resolution still needs a source-bundle adapter. Go embed consumers in source
ancestor packages are observed through the existing Go-tool observer; affected
readers require native admission before this command can publish. Inactive Go
files in an embed-bearing package also leave admission unresolved. Even a
wildcard that captures both output files needs its runtime contract checked.
Dynamic file readers and consumers outside the supplied repository/workspace
scope remain outside this command's observation. No repository-wide consumer
completeness is claimed. The executable must match the repository registry and
operation contract; rebuild it after changing those compiler sources.

### One syntax case across physical source members

The syntax corpus can associate an existing case with an ordered `case_sources`
entry. The first member remains the case's original `path`; additional members
belong to that same package and namespace. This optional registry-v2 field
describes physical source organization. It does not add a case or change its
capability/governance scope or parser profile.

The existing `explicit-consumer-continuation` case now reads both
`source.gooo` and `source_recovery.gooo` in
`internal/meta/executioncontinuation/policy/`. The family contains the original
types and all nine continuation/recovery activities. Both files were already in
the corpus inventory. The total remains 139 cases, including 137 capability
cases, and the distinct physical file population is unchanged.

Replay formats and reparses each member separately, then checks semantic and BX
preservation on the complete source set. The shared source-bundle parser keeps
each declaration's physical span; it creates no concatenated virtual source.
The result retains ordered `source_members` with exact file hashes and physical
line counts. `source_digest_kind: gooo/source-bundle/v1` identifies the ordered
name-and-byte digest; ordinary single-file cases retain their existing digest
representation. The report also records the registry's explicit case membership.

Missing members produce an unresolved observation without a usable input prefix.
Invalid syntax, conflicting identities or inconsistent package headers fail the
case. The caller's exact EntityFields profile is retained. Changing member names,
hashes, order or membership cannot be hidden by recomputing a report seal.

This route supports bounded, explicitly listed same-package sources (up to 64
members, 1 MiB per member, 16 MiB total). Imports still require package execution;
invalid-syntax examples retain their single-file diagnostic route. Reading the
case is sequential and does not establish a mutable-filesystem snapshot.
The joint workspace publisher above consumes this capability; syntax replay
itself performs no repository writes.

The existing `syntax-profile-migration` command also consumes this membership.
It pins all members, binds the native registry to the corpus, and checks the
complete family under both the original and requested profiles. Replaying the
primary alone cannot establish that a sibling's declarations support the new
profile. A changed sibling makes the saved request or plan stale, including a
comment-only change. Replay reads the pinned bytes; generation rereads the input
set before exporting the proposal.

Migration still proposes three writes: the primary source, its existing native
case definition, and the corpus definition. Unchanged siblings remain observed
inputs, and the case population is preserved. Registering an existing family
fragment as a new capability is refused even if that path also appears in
`meta_sources`. The path-only profile hint command remains advisory for that
single file; it does not select a profile for a source family.
