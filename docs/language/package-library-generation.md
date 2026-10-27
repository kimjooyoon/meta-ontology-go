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

An authored `computes` activity can have zero inputs, including a bound producer
or the workspace entry. A source constant can therefore supply another stage:

```gooo
activity Limit() -> Integer id "urn:example:limit" computes "return 262144"
activity Check(limit Integer, observed Integer) -> Boolean computes "return observed <= limit"
bind Limit.result -> Check.input0
```

Exporting `Limit=urn:example:limit` exposes `Limit()`; the entry receives only
the unbound observation in `Evaluate(observed)`. A workspace consisting of a
constant entry exports `Evaluate()` and its composition input template contains
an explicit empty object, `{}`. Plans record `zero_inputs: true` for those
activities. Missing/null caller input data retains its incomplete meaning.

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

Generation typechecks the emitted Go in process. It neither invokes the Go
toolchain nor executes the program, asks a model, or creates test expectations.
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
