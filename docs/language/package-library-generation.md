# Generate a Go library from Gooo

The development compiler can emit an importable Go source file from a workspace
whose entry and reachable activities already have authored `computes` bodies:

```sh
gooo package generate --go-package surface --out ./surface \
  examples/runtime-package-surface/gooo.workspace.json
```

This writes `surface/semantic.gooo.go`. The caller can use it directly:

```go
result := surface.Evaluate(surface.R_Observation{
    F_origin_ref: "local-go-list",
    F_import_paths: observedPaths,
})
fmt.Println(result.F_status, result.F_missing_count)
```

`--out` creates the output directory after successful generation and publishes
the complete source using the CLI's shared atomic writer. Generation errors leave
the previous file untouched. Existing file permissions are preserved, and equal
output bytes leave the existing file in place. Workspace inputs cannot also be
output targets. Keep handwritten host adapters in separate files: the generated
file is replaced as a whole. Publication errors are reported, including any
rollback failure. If only the success message cannot be delivered, the completed
file remains available and the command exits unsuccessfully.

Omit `--out` to stream Go source to stdout. Shell redirection opens its destination
before the compiler runs; callers using that route own staging and replacement.

The existing Gooo declaration owns the expected package paths and comparison
logic. The compiler owns record layout, field references, and the call sequence.
The caller supplies observations. There is no domain-specific Python generator
or separately maintained Go record schema in this route.

`Evaluate` takes unbound root inputs in the typed plan's order, executes bound
producers, and returns the entry result. Comments identify every argument's
activity, port and entity ID. Pure helper calls remain inside the generated code.
Use `package inspect --json` to inspect the original package/activity mapping.

A host can also read a source-owned plan before supplying its observations.
Expose selected stages by their stable activity IDs:

```sh
gooo package generate --go-package named --out ./named \
  --export-stage Credit=urn:gooo:example:named-inputs:credit \
  --export-stage Net=urn:gooo:example:named-inputs:net \
  examples/named-inputs/workspace.json
```

`Credit` and `Net` receive their own authored body inputs. `Credit(100)` returns
123; `Net(100, 5)` returns 95. `Evaluate(100, 5)` follows the declared bind from
`Credit` into `Net` and returns 118. All functions share the same generated
record types and field IDs. This lets Go read a declared plan and pass its typed
values to another stage without maintaining a second schema or copying fields.

`--export-stage` may be repeated with distinct exported Go names and distinct
planned activity IDs. A display rename of the source activity preserves this
selection. The emitted comment binds each public name to its activity ID.
Exports select activities in the entry's planned producer chain; a pure helper
called inside a body stays an implementation detail. To evaluate another
independent activity, choose it as the workspace entry. Invalid/ambiguous exports
fail before publishing output, retaining the previous generated file.

Records use `R_<source-name>` and exported fields use `F_<source-name>`.
Keeping the spelling after a fixed prefix avoids guessing acronym casing or
merging names such as `origin_ref` and `Origin_ref`. Conflicting record names
across packages use the workspace resolver's existing identity-derived names.
A source display rename changes this Go naming view; it retains the declared
stable IDs. Field tags carry `gooo:"<field-id>"`, original JSON names and declared
text bounds. Required values, optional pointers, slices and optional slices keep
their existing representation and source field order. No field-wise copy bridge
is generated; typed references are renamed together.
Source locals that collide with public types or private helpers receive fresh
names through their typed bindings. String values and other scopes stay intact.

Agents can combine this file with the existing `package interface --json` output.
Match the record's generated-region ID to its interface declaration, then match
each field's `gooo` tag within that record. For example, `Observation.origin_ref`
uses ID `urn:gooo:runtime-package-surface:observation:origin`; its field name cannot
be inferred from the last segment of that ID. Keep the owning entity when joining
fields instead of treating every field label or ID as a global key. Bind function
arguments by their recorded activity, port and position: two inputs with the same
entity type can carry different values. This recovers the generated view's
correspondence; a host's existing public API may still require an explicit adapter.

The output records compiler source/status, manifest, workspace image, original
source-file, lowered source and typed-plan digests. Entity/activity regions retain their stable ID
markers. These identify generation inputs, not producer authentication or a
claim that the program has run. Keep the exact original workspace to reproduce
the file; identical inputs, package name and compiler build produce identical
output bytes. Compiler source can be `UNBOUND_LOCAL_SOURCE` when build metadata
cannot bind it; the command does not infer a producer from the current checkout.
Hash the complete generated file separately when recording its consumer. A later
commit with equal source bytes does not become the original generation producer.
`source_json_string_sha256` uses the existing package image representation:
`sha256(json.Marshal(source.Content))`, with Go `encoding/json` string escaping.
The manifest and lowered-source hashes use their own raw bytes. Consumers can
retain raw original-file hashes alongside the explicitly labeled string hashes.

Add `--embed-origin` when the importing program needs this generation context
at runtime. The optional `GoooGenerationOrigin` string constant contains JSON
with schema `gooo/library-generation-origin/v1`, the entry's stable activity ID,
lowered-source and typed-plan hashes, selected stage IDs and public names, and
the compiler/workspace/original-source observations described above. For example:

```go
report.GenerationOrigin = surface.GoooGenerationOrigin
```

Reading the constant requires no file access, source parsing or Gooo evaluation.
The caller can retain it as a string or decode it using its ordinary JSON tools.
Its context belongs to the generator build, even when imported by another module;
it does not identify that module's execution or authenticate the producer.
The ordinary output remains unchanged when the flag is absent. The additional
constant shares the 256 KiB projection limit and rejects a conflicting public
export name before publication. Source input contents are represented by hashes,
not embedded in the JSON.

Generation typechecks the emitted Go in process. It neither invokes the Go
toolchain nor executes the program, asks a model, or creates test expectations.
The sixteen-record body budget applies to the union of selected bind stages,
called helpers and permitted construction alternatives. Every declared layout
is still validated, and the complete source identity, semantic fingerprint and
public record declarations remain available. Adding an unrelated valid record
can change that source identity while preserving the same importing Go program.
Realized candidates are checked again through their typed uses before scoring
or projection, including types absent from an original body hole.
The current route rejects every `assembling` contract, including a saved selection
checkpoint that still retains that contract. Existing workspace
and body-profile bounds apply, including the 256 KiB projection bound.

Go callers own decoding, UTF-8 and text-bound checks, missing/duplicate JSON
keys, null handling, and external effects. Struct tags describe the boundary;
ordinary `encoding/json` does not enforce all Gooo input rules. Use the existing
`package execute --cases ...` route and independent Go tests to check behavior.
Generation by itself does not report a functional completeness percentage.

Regression coverage imports generated libraries into a fresh Go module, checks
the existing fourteen runtime-package-surface expectations, optional-record
presence and exact integers, and ordered cross-activity inputs. These finite
cases establish the tested behaviors, not equivalence for all possible inputs.

A second regression changes the runtime surface declaration by adding one package
while preserving its record schema. The exact same Go consumer is compiled before
and after regeneration. Three fixed inputs per revision check missing packages,
undeclared packages and missing origin: six independent expected outcomes, with
zero handwritten consumer Go edits. This measures one same-schema policy change;
new fields, renamed Go views and other consumers can still require integration
work. The regression and generator have their own initial implementation cost.

The stage-export regression checks three fixed inputs against public producer,
public consumer and complete-graph results before and after an activity display
rename: 18 independent expected outcomes with unchanged handwritten Go. It also
checks the equivalent manual composition through their shared record type,
11 invalid export requests, and preservation of prior output after rejection.

The embedded-origin regression imports one generated program into a separate
module, reads its origin without evaluating a zero divisor, and checks two
independent arithmetic expectations. It also checks original-source hashes,
stage identity, duplicate flags, a name conflict, default omission and the
shared output limit. These establish this optional interface's boundaries;
they do not increase the number of supported Gooo grammar constructs.
