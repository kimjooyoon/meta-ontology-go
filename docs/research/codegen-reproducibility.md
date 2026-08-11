# Codegen follow-up: reproducibility experiments

This note extends the code-generation acceptance contract with executable experiment
design. It is intentionally independent of the open codegen research PR: it does not
modify that PR's branch or claim that the generator is already available on
`integration`.

The experiments cover five risks that are easy to miss when a generated file looks
valid once:

1. imports change with input order or alias allocation;
2. a protected marker can be malformed or retagged to move handwritten code;
3. a local handwritten edit causes an unrelated rewrite;
4. a new generator cannot safely consume an old artifact, or vice versa;
5. a Go-hosted bootstrap and a future `.gooo`-hosted build produce different
   artifacts or evidence.

The governing laws remain the DSL/IR round-trip and locality rules in
[`docs/spec.md`](../spec.md). The trust boundary follows the staged verifier plan in
[`.github/conformance-plan.md`](../../.github/conformance-plan.md) and the branch
ownership rules in
[`internal/verify/scope.go`](../../internal/verify/scope.go). The malformed-input
result is always a failed verification with evidence; an unimplemented stage is
reported as deferred, never as a successful self-hosting result.

## Evidence vocabulary

Each experiment produces a small record with stable fields. Host-specific details may
be retained as provenance, but must not enter the canonical digest:

| Field | Canonical? | Purpose |
| --- | --- | --- |
| Fixture ID and schema version | Yes | Names the exact experiment contract |
| Input DSL/IR digest | Yes | Identifies authoritative semantic input |
| Normalized semantic digest | Yes | Compares meaning independently of Go text |
| Generated source digest | Yes | Detects byte drift when bytes are promised stable |
| Generated-region/slot digest | Yes | Detects ownership and locality drift |
| Import projection digest | Yes | Isolates ordering and alias decisions |
| Source-map digest | Yes | Detects navigation/evidence range drift |
| Generator implementation/version | Yes | Explains version-skew outcomes |
| Go toolchain identity | Recorded, not canonical | Reproduces environment without hiding semantic equality |
| Absolute paths, timestamps, temp directories | No | Prevents host-dependent bootstrap digests |
| Builder/verifier identity | Separate attestation field | Distinguishes implementations without making them unequal |
| Failure class and rule ID | Yes | Makes expected rejection comparable across implementations |

An evidence record is valid only when all required canonical digests are present for
the experiment. A missing generator or CLI stage is `DEFERRED`, not `PASS`.

## Experiment A: deterministic imports

### Question

Does one semantic IR produce one import projection regardless of input order, map
iteration order, or process scheduling?

### Fixture family `CG-IMP`

Use one activity that requires two standard-library types and one aliased package:

```text
fixture: imports/order-v1
package: billing
entities: Order, Payment
activity: PayOrder(Order) -> Payment
imports: ["time", "encoding/json", {name: "clock", path: "time/tzdata"}]
```

Create permutations of entities, activities, ports, and imports independently. Ports
are deliberately tested separately: unordered declarations may be canonicalized, but
activity input/output order is part of the function boundary and must remain ordered.

### Procedure

1. Normalize each permutation in a fresh process.
2. Generate source, source map, semantic digest, and import projection digest.
3. Compile a temporary package containing the generated source; parsing alone is not
   sufficient because unused imports and aliases are compile failures.
4. Repeat under `go test -count=20` and `go test -race`.
5. Compare canonical fields byte-for-byte. Compare only semantic fields when a
   formatter version is intentionally different.

### Expected outcomes

- Permuting unordered input produces identical source and import digests.
- Reversing declared ports changes the signature and the semantic digest.
- Two incompatible aliases for one path are rejected with `CG-IMP-002`, not resolved
  by arrival order.
- An alias colliding with a generated package-level name is rejected before writing.
- An unused import is rejected by compile validation and classified as
  `CG-IMP-004`, not as a formatting failure.

### Failure classes

| Rule | Class | Disposition |
| --- | --- | --- |
| `CG-IMP-001` | input-order drift | Reject; deterministic output law broken |
| `CG-IMP-002` | ambiguous duplicate path/alias | Reject; import meaning is ambiguous |
| `CG-IMP-003` | alias/name collision | Reject; generated binding is unsafe |
| `CG-IMP-004` | parse succeeds, compile fails | Reject; projection is not a valid Go package |
| `CG-IMP-005` | port order silently normalized | Reject; semantic boundary changed |

## Experiment B: protected markers and tampering

### Question

Can a refresh distinguish a legitimate slot edit from a structural edit that would be
overwritten, retagged, or silently adopted by the generator?

### Fixture family `CG-MARK`

Start from a canonical generated activity with one slot and one handwritten block:

```go
//gooo:protected:start id="fixture/handwritten"
const Keep = "outside"
//gooo:protected:end id="fixture/handwritten"

//gooo:generated:start id="fixture/activity" kind="activity"
func Activity() int {
	//gooo:slot:start id="fixture/activity/implementation"
	return 1
	//gooo:slot:end id="fixture/activity/implementation"
}
//gooo:generated:end id="fixture/activity"
```

Apply one mutation at a time:

- remove an end marker;
- add a nested generated marker;
- close a region with a different ID or kind;
- duplicate a region or slot ID;
- move a slot outside its generated region;
- retag the activity region with another activity ID;
- edit a generated signature outside the slot;
- edit the protected handwritten block;
- insert an unknown marker attribute;
- change only the slot body.

### Procedure

1. Validate the previous artifact before rendering a replacement.
2. Record the output path set and bytes before generation.
3. Run the generator in memory, expecting either a complete result or no result.
4. Assert that every rejection has a stable marker/semantic ID and that no partial
   file was written.
5. For the slot-only mutation, assert the documented slot formatting policy. For all
   other protected mutations, assert a fail-closed result.

### Expected outcomes

The parser may accept legacy markers only through an explicit migration fixture. New
output has one canonical spelling. Marker IDs are ownership claims, not comments; a
retagged body must not be treated as belonging to the new activity without migration
metadata or an independently checked manifest.

| Rule | Mutation | Expected class |
| --- | --- | --- |
| `CG-MARK-001` | unterminated/nested marker | `MALFORMED_ARTIFACT`, reject |
| `CG-MARK-002` | mismatched close ID/kind | `OWNERSHIP_MISMATCH`, reject |
| `CG-MARK-003` | duplicate region/slot ID | `IDENTITY_COLLISION`, reject |
| `CG-MARK-004` | slot outside generated region | `BOUNDARY_ESCAPE`, reject |
| `CG-MARK-005` | generated edit outside slot | `PROTECTED_EDIT`, reject or explicit overwrite diagnostic |
| `CG-MARK-006` | retagged region | `OWNERSHIP_UNPROVEN`, reject |
| `CG-MARK-007` | slot-only body edit | `ALLOWED_SLOT_CHANGE`, preserve per policy |

## Experiment C: handwritten locality

### Question

Does an implementation-only edit remain local to its slot, while structural edits
change only the intended generated regions and direct dependants?

### Fixture family `CG-LOC`

Keep two activities, two entities, an import block, a protected handwritten block,
and a marker-outside helper in one file. Store these projections:

```text
fixtures/codegen/locality/
  baseline.go
  slot-edit.go
  rename-same-id.go
  add-activity.go
  remove-activity.go
  protected-edit.go
  expected.json
```

`expected.json` is a design sketch, not a Go package API. It records allowed region
IDs, disallowed protected region IDs, and the expected failure class. The exact bytes
should be loaded as files once the generator package lands.

### Procedure

1. Parse and index all marker ranges in `baseline.go`.
2. Compare each candidate with a generated skeleton in which owned region bodies are
   replaced by stable `(kind, semantic ID)` sentinels.
3. Compare protected slots, handwritten blocks, and marker-outside text byte-for-byte.
4. Compare generated region sets and source-map ranges by semantic ID.
5. Classify a change by the smallest allowed semantic closure, not by changed file
   path alone.

### Expected outcomes

- A slot edit changes one slot body only.
- A same-ID rename changes the owning generated region but keeps the slot ID and
  handwritten body associated with that activity.
- Adding one activity adds one region and its explicitly owned imports.
- Removing one activity removes its owned region; stale executable code is a failure.
- Editing an outside comment/helper or protected block is not silently preserved as a
  generated truth.
- An unrelated region has identical bytes and source-map range semantics.

### Failure classes

| Rule | Class | Disposition |
| --- | --- | --- |
| `CG-LOC-001` | outside-marker bytes changed | Reject as `UNOWNED_CHANGE` |
| `CG-LOC-002` | protected slot/handwritten body changed unexpectedly | Reject as `PROTECTED_CHANGE` |
| `CG-LOC-003` | unrelated generated region changed | Reject as `LOCALITY_ESCAPE` |
| `CG-LOC-004` | removed semantic ID remains executable | Reject as `STALE_REGION` |
| `CG-LOC-005` | source-map range points at old bytes | Reject as `STALE_MAPPING` |
| `CG-LOC-006` | expected slot-only edit reformatted by policy | Quarantine until byte/format policy is explicit |

The locality detector should be a guardian, not a generator. It must compare outputs
and report evidence; it must not repair a candidate projection or approve its own
repair.

## Experiment D: generator version skew

### Question

Which changes are compatible across generator versions, and which require migration
or a clean regeneration?

### Fixture family `CG-SKEW`

Pin a small matrix of generator identities and artifacts:

```text
fixtures/codegen/skew/
  ir-v1.json
  artifact-v1.go
  manifest-v1.json
  source-map-v1.json
  ir-v2.json
  artifact-v2.go
  manifest-v2.json
```

The fixture must include one stable semantic ID, one renamed display name, one slot,
one import alias, and one removed declaration. This distinguishes harmless formatting
changes from identity and ownership changes.

### Compatibility matrix

| Reader | Artifact | Expected result |
| --- | --- | --- |
| v1 | v1 | Read, regenerate, and reproduce the v1 canonical digest |
| v2 | v1 | Read only if schema/migration is declared; otherwise `VERSION_SKEW` |
| v1 | v2 | Never silently downgrade; `VERSION_SKEW` or explicit migration |
| v2 | v2 | Reproduce v2 golden and semantic digests |
| v2 | v1 with same IDs, new formatting | Semantic equivalence may pass; byte equality is not required |
| any | changed slot owner or semantic ID | Reject as `OWNERSHIP_MIGRATION_REQUIRED` |

### Procedure

1. Record generator schema, implementation digest, and canonical IR digest.
2. Run old-reader/new-writer and new-reader/old-writer cases in isolated temporary
   directories.
3. Compare semantic digests, generated-region identity sets, slot-owner maps,
   imports, source maps, and final Go compilation independently.
4. Require an explicit migration record for marker-schema changes, slot ID changes,
   or removed-region behavior.
5. Keep old artifacts as immutable goldens. Never overwrite a golden to make a skew
   test green.

### Failure classes

| Rule | Class | Disposition |
| --- | --- | --- |
| `CG-SKEW-001` | unknown artifact schema | Reject as `VERSION_SKEW` |
| `CG-SKEW-002` | semantic digest differs | Reject as `SEMANTIC_DRIFT` |
| `CG-SKEW-003` | same semantic ID maps to a new slot owner | Reject as `OWNERSHIP_MIGRATION_REQUIRED` |
| `CG-SKEW-004` | old reader silently accepts new marker grammar | Reject as `UNDECLARED_COMPATIBILITY` |
| `CG-SKEW-005` | formatting differs but semantic/source-map contract matches | Record as `TEXTUAL_DRIFT`, allow only by policy |
| `CG-SKEW-006` | generator version is missing from evidence | Reject as `INCOMPLETE_EVIDENCE` |

Version compatibility is a contract, not an accidental property of a permissive
parser. A reader that can parse an artifact but cannot prove ownership must fail.

## Experiment E: bootstrap artifact reproducibility

### Question

Can the Go-hosted stage and a future `.gooo`-hosted stage be compared on the same
semantic input without making the candidate verifier its own sole trust root?

### Fixture family `CG-BOOT`

Use a pinned source snapshot containing the billing DSL, generator schema, verifier
policy, and all declared fixture files:

```text
fixtures/bootstrap/stage-0/
  source-manifest.json
  semantic-input.gooo
  expected-ir.json
  expected-generated.go
  expected-evidence.jsonl
  README.md
```

The same fixture is consumed by the Go-hosted reference and, when implemented, the
`.gooo`-hosted candidate. The fixture README must state which stage is implemented;
an absent candidate is `DEFERRED`, never a passing comparison.

### Procedure

1. Build the Go-hosted reference in two clean directories with the pinned toolchain.
2. Build the candidate only if its implementation and promotion stage are enabled.
3. Generate artifacts from the same normalized source manifest.
4. Compare source, semantic, generated-region, source-map, and canonical evidence
   digests. Compare builder identity separately.
5. Re-run each build with a different temporary root and environment variable order.
6. Verify that timestamps, absolute paths, hostnames, and process IDs are absent from
   canonical artifacts.
7. Mutate one input byte, one generator byte, and one verifier-policy byte in separate
   runs. Each mutation must affect the expected digest and failure class.
8. Preserve the last-known-good Go artifact and rerun the comparison after forcing a
   candidate mismatch. The fallback must remain available.

### Required equivalence

The two implementations need not share a compiler or builder digest. They must agree
on the common contract:

| Value | Required relationship |
| --- | --- |
| source/input artifact set | Identical canonical digest |
| normalized semantic IR | Identical semantic digest |
| generated region and slot ownership | Identical identity/ownership digest |
| generated source | Identical only when byte reproducibility is promised; otherwise semantic/source-map agreement |
| verdict and normalized rule IDs | Identical |
| canonical evidence projection | Identical |
| builder/attestation identity | May differ, recorded separately |

### Failure classes

| Rule | Class | Disposition |
| --- | --- | --- |
| `CG-BOOT-001` | source manifest differs | Reject as `INPUT_DRIFT` |
| `CG-BOOT-002` | semantic digest differs | Block promotion as `SEMANTIC_DRIFT` |
| `CG-BOOT-003` | generated-region ownership differs | Block promotion as `PROJECTION_DRIFT` |
| `CG-BOOT-004` | canonical evidence differs | Block promotion as `EVIDENCE_MISMATCH` |
| `CG-BOOT-005` | output contains host path/time | Reject as `ENVIRONMENT_LEAK` |
| `CG-BOOT-006` | candidate mismatch cannot roll back | Block promotion as `RECOVERY_UNAVAILABLE` |
| `CG-BOOT-007` | candidate stage is absent but reported pass | Reject as `UNIMPLEMENTED_STAGE_CLAIM` |

This is the evidence needed for the staged model in
`.github/conformance-plan.md`: Stage 0 may remain Go-authoritative; Stage 1 requires
dual evidence; later promotion requires reproducible builds, independent comparison,
and rollback. No fixture in this note is evidence that those stages have passed.

## Golden and round-trip fixture catalog

The fixture catalog is intentionally small and orthogonal. Each fixture has one
authoritative input and one immutable expected outcome.

| Fixture | Authoritative input | Golden/output | Round-trip assertion |
| --- | --- | --- | --- |
| `imports/order-v1` | IR permutations | canonical Go/import/source-map digests | same IR meaning after projection |
| `markers/tamper-v1` | previous generated Go mutations | failure class and no-write evidence | no malformed artifact enters IR |
| `locality/pay-order-v1` | baseline plus one edit class | changed region/slot set | implementation-only edit preserves unrelated regions |
| `skew/v1-v2` | pinned IR and two artifact schemas | migration/version result | semantic ID and slot ownership survive only when declared |
| `bootstrap/stage-0` | pinned source manifest | Go-hosted artifact/evidence | repeated clean builds have equal canonical digests |
| `bootstrap/stage-1-deferred` | same manifest, absent candidate | explicit deferred record | no false parity claim |

Golden comparison must use a three-way result, not a boolean:

```text
PASS       expected invariant holds
FAIL       invariant was expected and did not hold
DEFERRED   implementation or stage is not available; not evidence of success
```

When text is allowed to differ, the fixture must state which canonical layer replaces
byte equality: semantic IR, region ownership, source map, import projection, or
evidence. A generic “formatting changed” label is too weak to approve a generator.

## Failure taxonomy and triage

All failures should be normalized into one primary class, with secondary facts kept as
evidence. The primary class determines the safe action:

| Primary class | Meaning | Safe action |
| --- | --- | --- |
| `INVALID_INPUT` | DSL/IR/import/name is invalid | Fix authoritative input |
| `MALFORMED_ARTIFACT` | Previous Go or marker structure cannot be trusted | Stop; repair or migrate artifact |
| `OWNERSHIP_MISMATCH` | ID, slot, region, or protected body changed owner | Stop; require explicit migration |
| `PROTECTED_EDIT` | A non-slot generated/protected body was manually changed | Stop; restore or record approved change |
| `UNOWNED_CHANGE` | Marker-outside bytes changed during refresh | Stop; inspect locality boundary |
| `LOCALITY_ESCAPE` | Unrelated generated region or source map changed | Stop; inspect generator normalization |
| `STALE_REGION` | Removed declaration remains in output | Stop; remove with ownership proof |
| `IMPORT_DRIFT` | Order, alias, usage, or import ownership changed unexpectedly | Stop; inspect canonical import resolver |
| `SEMANTIC_DRIFT` | Normalized meaning changed unexpectedly | Block round-trip/bootstrap promotion |
| `TEXTUAL_DRIFT` | Bytes changed while semantic contract still matches | Allow only with documented formatter/version policy |
| `VERSION_SKEW` | Reader/writer schema or generator versions are incompatible | Migrate explicitly or use pinned reader |
| `EVIDENCE_MISMATCH` | Digests, claims, or rule IDs differ | Block; compare independent evidence |
| `ENVIRONMENT_LEAK` | Host path/time/tool order enters canonical output | Reject as non-reproducible |
| `RECOVERY_UNAVAILABLE` | Known-good Go fallback cannot be rerun | Block promotion; restore recovery artifact |
| `DEFERRED` | Required implementation/stage is absent | Report honestly; do not promote |

Triage order is deterministic:

1. `MALFORMED_ARTIFACT`, `OWNERSHIP_MISMATCH`, and `PROTECTED_EDIT` stop before any
   write.
2. `SEMANTIC_DRIFT`, `EVIDENCE_MISMATCH`, and `RECOVERY_UNAVAILABLE` block promotion
   even if the generated Go compiles.
3. `LOCALITY_ESCAPE`, `STALE_REGION`, and `IMPORT_DRIFT` block the projection until
   the affected semantic IDs are explained.
4. `TEXTUAL_DRIFT` may be accepted only when the fixture names the permitted layer
   and both semantic/source-map checks pass.
5. `DEFERRED` remains visible in CI and release evidence.

## Independent test adapter contract

Because `internal/generator` may not be present on every integration snapshot, the
fixture runner should use a small adapter boundary rather than importing an absent
package directly:

```text
GeneratorAdapter
  Normalize(input) -> normalized IR + semantic digest
  Generate(input, previous) -> source + source map + region digest
  Lift(source) -> sourced semantic delta
  Version() -> schema and implementation identity
```

The Go-hosted adapter is the first implementation. A future `.gooo`-hosted adapter
must consume the same fixture corpus and emit the same canonical evidence fields.
The adapter may return `DEFERRED` for an unavailable operation, but it must not return
an empty successful result. The fixture runner owns comparison and classification; an
adapter must not redefine the failure taxonomy to make its own output pass.

## CI integration sequence

The safe rollout does not wait for the generator implementation to land:

1. Land this experiment design and immutable fixture catalog as a research artifact.
2. Add the branch-scope alias through the CI-owned change; do not bypass the allowlist
   or weaken the protected-region policy from this lane.
3. Run the independent fixture runner against the Go generator when its package is
   available; until then, record `DEFERRED` with the missing package/command.
4. Add generator and bootstrap adapters in their owning lanes, preserving the same
   fixture IDs and failure classes.
5. Promote only after repeated clean runs, race checks, source-map checks, and
   dual-evidence comparisons pass.

The current integration baseline explicitly defers `gooo check` and
generated-freshness while the CLI is a stub. That state is a known baseline fact, not
a passing codegen experiment. A future CI run should report the transition from
`DEFERRED` to `PASS` in evidence rather than rewriting old goldens.
