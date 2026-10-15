# Refactoring a Gooo activity

[한국어 사용 안내](activity-refactoring.ko.md)

An activity has a source name and a stable semantic ID. Changing `Calculate` to
`요금` should keep the activity's identity, its callers and its source-owned
requirements connected. The development compiler now has a small operation for
that job:

```sh
gooo refactor rename-activity --from Calculate --to 요금 main.gooo > renamed.gooo
gooo refactor rename-activity --from Calculate --to 요금 --json main.gooo
```

The first command prints the complete renamed source. Save it to a **different
file**; shell redirection to `main.gooo` would truncate the input before the
command can read it. The second prints an edit receipt containing the source,
stable activity ID, byte ranges and hashes of both versions. Both commands read
the original file without changing it.

## What stays connected

- A declaration without an explicit ID gains its existing resolved ID before
  its name changes. An authored ID keeps its original spelling.
- Direct calls in `computes` bodies, including nested and parenthesized calls,
  refer to the new name. Parameters and local variables retain their binding.
- Local `bind` endpoints, `criteria` predicates and `observe` oracle references
  use the new name.
- Declared expression alternatives and checkpoint baselines follow changed
  calls. Choice IDs, selected labels, cases and separate evaluation inputs keep
  their original values.
- Comments, intent text, string values and field labels keep their wording.

The transformation uses parsed source locations. It checks the renamed file's
identities, PROV relations and runtime binding identities before returning any
output. A local variable that would capture the new call name produces an error.
So do declaration collisions and names reserved by the current Go projection.

In the experimental LSP, renaming an activity from its declaration uses this same
operation. It returns edits for the client to apply. Free direct computes calls also resolve to the activity identity. Cursor targets
in alternatives, imported references and workspace-wide rename remain future work.

## Scope and evidence

This first operation covers one file without imports, up to 128 KiB. Supported
source contains ordinary declarations, statement-based `computes` bodies and declared
assembly choices. `source_fill` and generated `search` candidate scopes require
additional mapping and currently return an error. An expression alternative with
an ambiguous enclosing local name also returns an error. A syntax or resolution
failure returns no partial proposal. Legacy value-program spellings such as
`computes "int.add:1"` need a separate profile adapter and currently return an error.

External callers, workspace entry names, case-file input keys and consumers of
package interfaces still need coordinated changes. The command cannot discover
those consumers from one file. Keep that boundary in mind before replacing a
public source file.

The receipt describes a source transformation. Source and interface hashes can
change even when the stable ID stays the same. Existing execution receipts remain
attached to the exact old source. Checks and observations for the renamed program
must have their own source identity. This operation executes no program, performs
no model prediction and supplies no new runtime correctness score.

The [example](../../examples/activity-rename/README.md) records the first source
transformation. Source-only regression checks cover identity pinning, explicit
IDs, Korean names, raw strings, lexical capture, fixed criteria, oracle and binding
references, checkpoint baselines, and the CLI/LSP entry points.

The integration check also renames a helper, its fixed predicate and the owner to
Korean names. Their typed predicate graph and candidate structural feature bytes
remain unchanged. This depends on the decision runtime's Unicode identifier
correction in [v0.2.48-experimental](https://github.com/kimjooyoon/gooo-decision-runtime/releases/tag/v0.2.48-experimental);
plan IDs and UTF-8 byte limits retain their existing rules. A complete
source-bound plan digest can change with its recorded activity name.

## Why this belongs in the language

Gooo keeps allowed choices and requirements beside the activity. A useful
refactoring therefore follows those relationships along with ordinary calls.
The stable ID is the activity's inventory number; its source name is the label
used by people and callers. This operation changes the label and reconnects the
references that the current source file owns.

A decision model could later propose which activity to split, merge or rename.
Gooo would still need a source operation that records exactly what changed. The
current rename path is deterministic and provides that first reusable operation.

The editor resolves free direct computes calls to activity IDs before renaming.
Other unresolved activity-name requests ask to start at the declaration. A cursor
that cannot resolve must not fall back to changing only a declaration by spelling.

## Select by source position

Use either `--from NAME` or `--at-byte OFFSET` with `--to NAME`. The offset is a
zero-based byte in the original file, within a declaration name or a free direct
computes call. A quoted body maps escapes to decoded AST positions; raw bodies
retain their source bytes. Local bindings, selectors and prose do not identify
activities. No imports or assembly-alternative cursor targets are supported.

The experimental LSP prepare request returns the actual encoded call-name range,
and rename uses the same identity-preserving source operation. UTF-16 editor
positions are converted through the existing source-position adapter.

[The call-selector example](../../examples/activity-call-rename/README.md) retains
one actual source transformation and its development build/source identity.
