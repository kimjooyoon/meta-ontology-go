# Generate a Go library from Gooo

The development compiler can emit an importable Go source file from a workspace
whose entry and reachable activities already have authored `computes` bodies:

```sh
gooo package generate --go-package surface \
  examples/runtime-package-surface/gooo.workspace.json > surface_generated.go
```

Place the file in a Go package named `surface`. The caller can use it directly:

```go
result := surface.Evaluate(surface.R_Observation{
    F_origin_ref: "local-go-list",
    F_import_paths: observedPaths,
})
fmt.Println(result.F_status, result.F_missing_count)
```

The existing Gooo declaration owns the expected package paths and comparison
logic. The compiler owns record layout, field references, and the call sequence.
The caller supplies observations. There is no domain-specific Python generator
or separately maintained Go record schema in this route.

`Evaluate` takes unbound root inputs in the typed plan's order, executes bound
producers, and returns the entry result. Comments identify every argument's
activity, port and entity ID. Pure helper calls remain inside the generated code.
Use `package inspect --json` to inspect the original package/activity mapping.

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
