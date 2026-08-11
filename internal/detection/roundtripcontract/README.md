# Round-trip evidence contract

This prototype records the smallest reusable input/output boundary between a
round-trip detector and later AST, semantic IR, BX, codegen, LSP, cache,
provenance, and CI implementations.

## Falsifiable hypothesis

`H-roundtrip-stable-id-v1` claims that presentation-only renames and ordering
changes preserve semantic evidence, while stable-ID or semantic-fact changes
produce a finding. The falsifier is either a changed canonical evidence record
for a presentation-only mutation or no finding for a semantic mutation.

The embedded fixture has 3 nodes, 2 facts, 3 generated regions, and its measured
DSL+Go source byte count. The scenarios include:

- a pass case for a display rename;
- a negative case for changing `billing://entity/order` to another stable ID;
- a deferred case when the future gooo-hosted adapter is absent.

Only `pass` is merge-eligible. `fail` and `deferred` are observable evidence,
never implicit success.

## Stage contract

Adapters emit `Evidence` with artifact URI/digest references, deterministic
findings, measurements, and an explicit status. Go-hosted code can produce the
same record today. A future gooo-hosted implementation may produce it after its
adapter exists; until then its result remains `deferred`.

`CanonicalJSON` is the boundary for CI artifacts, content-addressed caches, and
PROV evidence entities. `Scenario` and `Assessment` are the boundary for
falsifiable research and merge policy.
