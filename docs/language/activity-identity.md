# Keeping an activity identity across a rename

Development source supports an optional `id` after an activity's result type:

```gooo
package arithmetic
namespace arithmetic
entity Integer id "urn:gooo:type:integer"
activity Calculate(Integer) -> Integer id "urn:gooo:activity:sum" computes "return input"
```

The name is how source code calls the activity. The ID is how the semantic graph,
generated Go markers, package interface and execution trace identify it. Think of
the name as a label on a drawer and the ID as the drawer's permanent inventory
number. Changing the label to `합계` retains `urn:gooo:activity:sum`.

Write `id` before `computes` or `assembling`. An absolute URI-like ID such as a
`urn:` is accepted. An empty or relative ID is rejected. Two declarations within
a package cannot share an ID, including across source files. An explicit activity
ID also cannot collide with an entity's ID. Workspace execution lowers reachable
packages into one graph; identities in that graph must be distinct.
URI scheme and host casing normalize when checking identities. For example,
`URN:gooo:activity:sum` and `urn:gooo:activity:sum` identify the same activity.

Declarations without `id` continue to derive their activity ID from namespace and
name. Formatting preserves an authored ID even when it equals the derived value.
This allows a later rename to keep that identity.

## What a rename still requires

Update calls, bindings, the workspace entry and input keys that use the old name.
The activity ID does not create an alias for those references. Source hashes and
interface digests change when source names change. A saved execution receipt is
bound to its exact source and must be regenerated for the renamed source.

The same ID expresses continuity of identity. A body or signature change still
needs its own behavioral observations. Finite examples describe the cases they
cover; they do not establish correctness for every possible input.

## Evidence in this change

Parser and formatter round-trips retain explicit IDs, including the authored
default ID. Bidirectional Get-Put and Put-Get checks retain the identity and its
PROV relations. Package interfaces retain the ordered input/output types after
a rename. Native workspace tests use English and Korean names, verify generated
markers and delivery trace IDs, and replay saved receipts with zero model calls.

The own compact QAT model was also used with the
[ecosystem workbench](https://github.com/kimjooyoon/gooo-ecosystem-workbench/tree/59bc4118a5c4b50843412ed793d61373b065183f).
Changing `Calculate` to `합계` with the same ID produced two observations: a changed
entry name and a changed declaration name. Both deterministic and model selection
classified them as package wiring and source name changes. The authored policy
selection covered 51/51 fields in 17 cases; this rename input had no independent
correctness labels (0/0). The model was called once, and saved replay made zero new
model calls. No training or weight update was performed for this observation.

Whole-command measurements on one local run were 1.35 seconds for deterministic
selection and 0.76 seconds with the model; maximum resident size was 87,539,712 and
87,457,792 bytes respectively. Command order and warmed build caches affect these
times. They measure the workbench command and its native toolchain activity,
including model selection, rather than isolated model latency or host CPU load.

See [package interfaces](package-interface.md) for the data consumed by ecosystem
tools and [body generation](body-codegen.md) for bounded model selection.
