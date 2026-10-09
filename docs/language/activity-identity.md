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

See [package interfaces](package-interface.md) for the data consumed by ecosystem
tools and [body generation](body-codegen.md) for bounded model selection.
