# Source ownership during automatic Go transformations

`internal/meta/sourceownership/policy.gooo` owns the permission to rewrite a Go
view. Native parsing supplies the presence of a standard generated-file header
or a Gooo generated/slot comment. The `Resolve` activity returns the permission,
reason, stable rule ID and next operation. Text resembling a marker inside a Go
string is ordinary data.

Generated structure routes to `regenerate-source-owner`. The Go measurement
observer, source-policy evaluator, repository projector and direct extraction
path consume the same compiled Gooo policy. This route describes the required
owner operation; a general automatic owner-regeneration executor is not provided.
Use the known source compiler and workspace for the actual artifact.

```sh
go run ./cmd/gooo package generate --go-package sourceownership \
  --out internal/meta/sourceownership \
  internal/meta/sourceownership/gooo.workspace.json
```

File/function sizes and repository line totals retain generated output. Its raw
projection size-debt observation also remains visible. Generic Go rewriting is
`NOT_APPLICABLE` under the source-owned rule, so its size does not select a
handwritten-code transformation. This does not assert freshness, reduce an
observed size, relax the compiler's output bound or establish a speed improvement.
Canonical regeneration and independent program tests still check the artifact.
A source-subject witness retains the observed `generator_owned` flag in its
identity and checks its correspondence with the indicator through the shared
source-policy validator. The CI observation gate uses that existing native
consumer before checking JSON shape. It accepts neither an unbound exemption nor
an owner flag that contradicts an ordinary rewrite indicator. Historical reports
without the optional flag keep their existing representation.
A direct or stale extraction request is rejected before generating replacement
files. Mixed files containing Gooo-owned regions require an owner-aware route.

When handwritten Go is split, an unaliased import's qualifier comes from its
actual package declaration. A directory name, hyphen or `/v2` suffix cannot
establish that qualifier. The splitter reads the owning package, makes a differing
name explicit in its AST, then retains that binding in the generated parts.
Unresolved package declarations remain unresolved instead of silently losing
their imports. Explicit, blank and dot-import handling keeps its existing contract.

Regression scopes cover three declared-name/directory variants before and after
native splitting, generated-versus-handwritten size observations, a marker stored
as data, source policy regeneration, contradictory metric identities and direct
generated-source requests. They establish these finite boundaries, not all Go
transformation equivalences.
