# Read an entry's source call graph

Before executing a workspace, inspect the packages and call closure prepared from
its declared source files:

```sh
gooo package inspect --json examples/workspace-call-identity/workspace.json
```

This command is available in development source containing this change. It reads
the same manifest and source loader as package execution, then prepares the source
carrier without building or running a native program. The result is `PREPARED`
when the package graph and entry closure resolve. Source files remain unchanged.

## Prepare editable inputs

In development source containing this change, an agent can ask the same prepared
graph for one input row:

```sh
gooo package inspect --inputs-template examples/workspace-call-identity/workspace.json > inputs.json
```

For this example, the generated row contains `"example/app:Main": 0`.
The output uses `gooo/body-composition-inputs/v1` and original package/activity
keys, such as `example/app:Main`. For an activity with multiple unbound inputs,
each key ends in its declared port (`.input0`, `.input1`, and so on). Existing
bindings remove producer-fed ports from the row. Values are type-shaped zeros to
edit before passing the file to `gooo package execute --inputs inputs.json`.
Required repeated record fields start as empty arrays; optional fields are
omitted.
The command does not run a body, select a candidate, predict with a model, or
claim that the values are useful cases or expected outputs. A source graph with
no unbound caller ports has no input row to produce and returns an error.

## Identity follows the declaration

`program.pure_calls.activities` links each original package activity to its name
and identity in the lowered Gooo source. When an activity declares an explicit
ID, both `activity_id` and `lowered_activity_id` retain that canonical ID. The
internal function name can change while referring to the same source activity.
An activity without an explicit ID keeps the existing namespace/name derivation.

Earlier call reports derived IDs from names even when the declaration supplied
an ID. Source flattening retained the authored ID, so that report could disagree
with the lowered source. The corrected report uses the declaration's ID for both
views. Existing saved reports retain their original bytes; reconstruction compares
the complete program carrier, so an old inconsistent carrier can require a new
observation rather than being silently rewritten.

## Read the coverage

| Field | Information |
| --- | --- |
| `manifest_digest` | Exact manifest bytes read for this inspection |
| `program.workspace_image` | Resolved package exports, source digests and bindings |
| `program.entry` | Requested entry and its lowered name |
| `program.pure_calls.activities` | Original and lowered identities of call participants |
| `program.pure_calls.sites` | Caller/callee package keys, source surface and relative byte spans |
| `program.lowered_gooo_source` | The prepared source carrier for later construction/execution |

Call-site offsets belong to their reported body or alternative surface, not an
absolute source-file offset. The entry closure includes its binding producers,
called helpers, declared criteria and supported alternatives. The workspace image
covers the declared packages; the entry call closure covers its dependencies.
These are separate scopes when planning a rename across files.

Activity signatures can use visible entity names such as `Integer` or qualified
import types such as `rules.Integer` in development source containing the
[qualified activity type change](qualified-activity-types.ko.md). The qualified
spelling resolves through that declaring file's import alias to the entity ID.
Qualified calls such as `rules.요금(input)` remain supported in the body. Bare
type ambiguity and conflicting aliases within a file remain resolution errors.

The existing loader limits apply: up to 64 packages and 512 source files, 1 MiB
per source and 16 MiB across the source set. Source preparation uses the existing
package and body profiles. This inspection makes zero model predictions and
program executions. Actual output values, finite-case coverage and behavioral
agreement need their own execution records.

See [the source example](../examples/workspace-call-identity/README.md).
