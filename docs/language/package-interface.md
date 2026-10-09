# Package interfaces

`gooo package interface` exports the declarations of a workspace as data for
documentation, editors, and API change tools:

```sh
gooo package interface examples/package-optional-record-flow/gooo.workspace.json
gooo package interface --json examples/package-optional-record-flow/gooo.workspace.json
```

It uses the same parser, semantic lowering, package type resolution, and import
order as `package resolve`. A workspace can have one file or multiple packages.
Package names and imports can be omitted from the manifest when the source
already declares them. The command reads source files and writes its result to
standard output; activity bodies are not executed.

The JSON receipt has schema `gooo/package-workspace-interface-receipt/v1`, a
manifest digest, and an `interface` with schema `gooo/package-interface/v1`.
Each package contains source digests and its own declarations. Types imported
by an activity point to the original entity ID; their full declarations occur
in the defining package. Temporary type placeholders used during lowering are
excluded.

For each declaration the interface records its name, stable ID, kind, and
source filename. Entity `shape` is `record` for an explicit fields block and
`nominal` otherwise. Record fields preserve declaration order and include:

- `id` and `name`, plus aliases when present in the semantic IR;
- `type_id`, resolved by the compiler;
- `presence`: `required` or `optional`;
- `cardinality`: the validated semantic cardinality.

Activities include ordered input type IDs and an output type ID. Repeated
input types remain repeated. Activity IDs currently come from the compiler's
namespace/name convention; renaming an activity therefore changes its ID.
Entity and field names can change while their explicit IDs stay the same.

For example, the optional-record fixture exposes `Profile.note` as an optional
string and `Profile.count` as an optional integer. A later change from optional
to required remains visible even if the `Profile` and `count` IDs are unchanged.
That gives a change tool the information needed to explain which callers need
attention.

The interface digest binds the full projection. `image_digest` identifies the
underlying package build. Reordering manifest package/source lists preserves
the normalized interface; changing source contents changes the source evidence.
Current field profiles and their existing restrictions apply. Invalid sources,
unresolved imports/types, and conflicting identities within a package return a
failed receipt without a partial interface.

This command describes declared structure. A compatibility tool also needs to
consider whether a type is used as input or output, source names, bindings, and
runtime behavior. The existing `emit --kind operation-interface` v1 format
continues to provide the compact declaration signature for its current users.
