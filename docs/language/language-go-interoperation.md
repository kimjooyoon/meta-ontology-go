# Language Go Interoperation

## Decision

`LANGUAGE-GO-INTEROPERATION` is satisfied only when the versioned `24`-case metaprogram reports `PASS / EXACT`.
The denominator is fixed as `8` existing-generator projections, `8` Go 1.27 boundaries, and `8` expected fail-closed boundaries.
No result is inferred from file presence or a successful parser call.

## Meta operation

The operation `reify-go-projection-and-prove-type-identity` executes this in-memory path:

1. Construct a fixed `generator.SemanticIR` value.
2. Execute the existing generator without modifying it.
3. Parse the generated source into `go/ast`.
4. Project the AST back to canonical Go with the fixed Go 1.27 formatter.
5. Check the package with `go/types` and normalize exported objects and methods.
6. Replay generation or canonical AST projection and compare API digests.
7. Reject invalid syntax, invalid types, imports, empty exported APIs, and unknown payloads.

This follows gomacro's useful idea that code can be ordinary data represented by Go AST nodes. It deliberately rejects gomacro's ambient file and network authority because those effects would make this CI witness non-deterministic. Gomacro is a design reference, not a dependency or novelty claim: <https://github.com/cosmos72/gomacro>.

## Go 1.27 boundary

The fixed boundary uses generic methods, generic aliases, generalized function assignment inference, materialized `go/types.Alias` nodes, and `go/types.Hasher` equality. The release contract is <https://go.dev/doc/go1.27>.
Formatting equality is valid only under the CI-pinned Go 1.27 toolchain because `go/format` output may change across Go versions.

## Exact indicators

The artifact publishes exactly `18` indicators: `3` outcomes, `8` drivers, and `7` guardrails.
The successful fixed values are `24/24` cases, `16/16` accepted positive boundaries, `8/8` expected rejections, `8/8` generator projections, `8/8` Go 1.27 boundaries, `16/16` canonical replays, `16/16` type-identity replays, `5` generic methods, `2` aliases, `8/8` source maps, and `32/32` AST reifications.
All unresolved, invalid acceptance, unknown acceptance, ambient authority, repository write, and mutation-authority values must be exactly `0`.

## Munchhausen choices

`FOUNDATION` binds the versioned registry, Go 1.27 toolchain, concept, code, metrics, and use cases.
`COHERENCE` binds generator output, AST, types, normalized API, and replay identities.
`REGRESSION` proves that all eight negative boundaries are rejected without ambient effects or mutation authority.

## Use cases

`project-gooo-to-go-api` proves that the existing Gooo projection has a stable exported Go boundary.
`consume-go-1.27-boundary` proves that downstream Go 1.27 generic method and alias syntax can be represented and type-checked.
`reject-go-boundary-unknowns` lowers resolution and fails closed for invalid, ambient, empty, or unknown boundaries.
