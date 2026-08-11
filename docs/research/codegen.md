# Generated Go projection: acceptance contract

This note defines the acceptance boundary for the Go projection. It is a research
and integration contract, not a generator implementation plan. The generator is a
projection of semantic IR; it must not become a second source of business meaning.

The contract is derived from [`AGENTS.md`](../../AGENTS.md), the DSL/IR laws in
[`docs/spec.md`](../spec.md), the projection boundaries in
[`docs/architecture.md`](../architecture.md), the billing handwritten example, and
the candidate implementation on the `agent/generator` branch. The implementation
branch already has useful tests for deterministic output, slot retention, locality,
and malformed markers. The cases below extend those tests to the failure modes that
would otherwise make generated code unsafe to regenerate.

## Authority and ownership

The projection is safe only when ownership is explicit:

| Surface | Authoritative owner | Generator obligation |
| --- | --- | --- |
| Business declaration and stable semantic ID | `.gooo` DSL / lowered IR | Use the ID as identity; never use a display-name change as identity change |
| Structs, function signatures, imports, markers, and source-map ranges | Generated Go projection | Emit deterministically and replace only owned regions |
| Irreducible implementation | Handwritten slot body or explicitly external handwritten helper | Preserve it across regeneration and never infer business truth from arbitrary helper code |
| Verification policy and protected kernel | CI/verifier policy | Treat policy, ontology, and verifier files as outside the application escape hatch |
| Provenance/evidence | Append-only build records | Record generation and verification as derived evidence, not as mutable source |

A generated file is therefore a composite artifact:

```text
protected header/import ownership
  + generated structural regions
      + handwritten slot bodies
  + marker-outside handwritten text
```

Only the structural regions and any explicitly owned import section may be rewritten.
The marker-outside text and slot bodies are not disposable formatting input. They are
the locality boundary that lets a Go implementation evolve without a whole-file diff.

## Marker and slot contract

The preferred marker form is attribute-bearing and uses stable semantic IDs:

```go
//gooo:generated:start id="billing://activity/pay-order" kind="activity"
func PayOrder(order Order) Payment {
	//gooo:slot:start id="billing://activity/pay-order/implementation"
	return Payment{OrderID: order.ID}
	//gooo:slot:end id="billing://activity/pay-order/implementation"
}
//gooo:generated:end id="billing://activity/pay-order" kind="activity"
```

The exact Go shape may evolve, but these invariants must remain stable:

1. Every generated region has one stable ID and one kind. The ID is not derived from
   a display name at regeneration time.
2. Every slot has a stable ID, belongs to exactly one generated region, and is
   replaced by its default body only when it has no previous body.
3. Start/end markers are balanced, non-nested at the generated-region level, and
   close the same ID and kind that they open. Legacy input may be accepted by an
   explicit migration path; new output must use one canonical spelling.
4. Duplicate region IDs, duplicate slot IDs, cross-kind ID collisions, unknown
   marker attributes, and malformed quoting fail closed before any output is
   written.
5. A slot body is the only editable content inside a generated region. An edit to
   the surrounding signature, marker, or generated statements must either be
   rejected as a protected-region edit or be discarded with a diagnostic that names
   the affected semantic ID. Silent loss is not acceptable.
6. A slot ID change is an identity migration, not a harmless rename. Unless an
   explicit migration record exists, the generator must refuse to overwrite or
   orphan the old body.
7. A generated region removed from IR must not remain as stale executable code. The
   generator should remove an owned, well-formed region; if ownership cannot be
   established, it should stop and require an explicit migration rather than guess.
8. The generator must not discover handwritten logic by scanning every Go function.
   Only marker-declared slots or registered semantic symbols participate in the
   projection contract.

The existing candidate marker parser checks nesting, duplicate slot IDs, and
unterminated regions. It does not by itself prove that a previous region's body was
unchanged outside the slot, nor can a bare marker ID prove that a human has not
retagged a region. Those checks belong in the acceptance gate or in a future signed
manifest/source-map record.

### Escape hatch rules

The handwritten escape hatch is deliberately narrow:

- structural boundaries come from IR and may be regenerated;
- implementation logic lives in a slot or in a separately registered handwritten
  helper, such as the `PayOrderLogic` example;
- the helper's ordinary implementation details are not lifted into semantic facts;
- a helper reference carrying a registered semantic identity may produce a sourced
  semantic delta, subject to the bidirectional rules in `docs/spec.md`;
- generated code must not silently copy a slot body into a different semantic ID;
- the first generation may use a deterministic default body, but later generations
  must preserve the user's body even when the default changes.

If a project wants a separately maintained handwritten file, the generator must not
try to own that file by filename convention alone. Registration should identify the
semantic activity and source span; otherwise a same-named helper in another package
could be mistaken for the implementation slot.

## Determinism and import stability

Determinism is a semantic safety property: a cache hit, a source-map range, and a
review diff are trustworthy only when identical IR produces identical bytes.

The canonicalization order should be:

1. Validate package names, identifiers, entity references, slot IDs, import paths,
   and duplicate identities.
2. Sort unordered collections by stable identity. Entities, activities, and imports
   must not depend on map iteration or input slice order.
3. Preserve semantically ordered collections. Activity input and output port order is
   part of the Go boundary and must not be sorted merely to make output look stable.
4. Resolve import aliases from stable inputs. Reject an ambiguous duplicate instead
   of choosing an alias based on arrival order.
5. Render with a fixed header and LF newlines, then run `go/format` on newly rendered
   structural output.
6. Build source-map ranges only after the final bytes exist.

Import acceptance must cover more than textual sorting:

- permuting the IR import list must not change output;
- the same path with incompatible aliases must be rejected or canonicalized by a
  documented rule;
- aliases must not collide with generated package-level names;
- generated imports must be used, and the result must compile, not merely parse;
- adding or removing a required type must update the owned import section without
  touching unrelated regions;
- blank imports, standard-library paths, and paths requiring escaping must have an
  explicit policy rather than relying on string concatenation.

The candidate generator sorts imports by path and alias, which is a good starting
point. Its duplicate key includes both alias and path, so two aliases for one path
are still a potential ambiguity. Its initial import block is also outside the
region-patching path. These are acceptance gaps, not reasons to make import order a
property of the DSL.

## Locality and protected areas

Locality means byte-level containment of change where byte-level comparison is
appropriate, and semantic containment where formatting necessarily changes bytes.
The following rules make that distinction testable:

| Change | Allowed output delta |
| --- | --- |
| Implementation-only edit inside one slot | That slot body, plus any intentionally regenerated formatting for that slot |
| Rename with the same semantic ID | The owning region and its source-map mapping; the slot body remains associated with the same slot ID |
| Add one entity/activity | Its new region and any required owned import/source-map entries |
| Remove one declaration | Its owned region and dependent structural references; no stale executable region |
| Change one declaration | That declaration's region and its direct structural dependants only |
| Edit a comment, helper, or license outside markers | The exact outside bytes are retained |
| Reorder unordered IR input | No output change |
| Reorder declared activity ports | The signature changes because port order is semantic |

The generator must not run whole-file formatting on the previous composite file when
patching it. Doing so turns an implementation-only edit into a repository-wide diff.
Formatting a fresh generated block is safe only if copied slot bytes are either
preserved exactly or the project explicitly chooses and tests a slot-formatting
policy. The current candidate path formats a rendered block containing a copied slot;
that makes the claimed byte-for-byte slot guarantee ambiguous. The acceptance test
must settle this deliberately.

Protected-region checks should fail before writing a partially updated file. At
minimum, the gate should detect:

- missing, nested, mismatched, duplicated, or malformed markers;
- a generated region whose non-slot content differs from the canonical projection;
- a slot body that is now associated with a different activity or ID;
- a package clause that disagrees with the IR;
- an old generated region whose identity is absent from the new IR;
- a generated file that parses but does not compile because its import/type boundary
  is stale;
- attempts to use a generated-region marker to edit ontology, verifier semantics, or
  CI policy files.

The verifier's path scope is useful as a coarse guard, but path scope alone is not a
substitute for semantic locality. The acceptance record should include the changed
semantic IDs and the generated-region IDs so a guardian can distinguish a legitimate
dependency update from an unrelated rewrite.

## Acceptance test matrix

These are proposed acceptance cases for `internal/generator` and the semantic
conformance job. Each test should use a temporary directory and compare bytes or
semantic ranges explicitly; substring-only assertions are insufficient for locality.

| ID | Scenario | Expected result |
| --- | --- | --- |
| CG-01 | Generate the same normalized IR twice | Bytes and source-map mappings are identical |
| CG-02 | Permute entities, activities, and imports in the input | Output is byte-identical to CG-01 |
| CG-03 | Generate fresh output, run `go/format`, and generate again | Output is already gofmt-stable and the second run is unchanged |
| CG-04 | Parse the generated file with comments and compile a fixture package | Both parsing and `go test` succeed; no unused or missing imports |
| CG-05 | Change only a slot body containing comments, blank lines, and a string literal | The body is retained according to the documented byte/format policy |
| CG-06 | Change a slot default after the first generation | Existing handwritten body wins; default is used only for a new slot |
| CG-07 | Rename an activity while retaining its semantic ID | The same region and slot IDs survive; only intended structural names change |
| CG-08 | Rename an activity and change its semantic ID without migration metadata | Generation fails closed; the old body is not copied to the new ID |
| CG-09 | Edit text outside all generated regions | Outside bytes are identical before and after regeneration |
| CG-10 | Change one activity in a file with two entities and two activities | Unrelated region bytes and mappings are unchanged |
| CG-11 | Add one declaration | Only the new owned region plus required owned structural entries are added |
| CG-12 | Delete one declaration | Its owned region is removed; stale generated code is not retained |
| CG-13 | Feed an unterminated, nested, mismatched, duplicated, or retagged marker | No output is written and the error names the marker/semantic ID |
| CG-14 | Put a slot outside a generated region or duplicate a slot ID | Generation fails before patching |
| CG-15 | Use the same semantic ID for an entity and an activity | Generation fails; no map-key overwrite or region loss occurs |
| CG-16 | Use an invalid package, identifier, field, port, or entity reference | Generation fails with a deterministic diagnostic |
| CG-17 | Permute imports, add aliases, and add a path that is unused | Canonical order is stable; ambiguous/unused imports fail with a useful error |
| CG-18 | Add/remove a type that changes imports in an existing file | The owned import section is updated; outside text and unrelated regions stay fixed |
| CG-19 | Reverse declared activity port order | The signature changes in the declared order; the generator does not sort it away |
| CG-20 | Inspect source-map ranges after every locality case | Every region and slot range is in bounds, non-negative, and points at its marker/body |
| CG-21 | Run DSL → IR → Go → lifted IR on a fixture with one registered semantic call | The result is semantically equivalent, and only source-backed deltas are added |
| CG-22 | Run generation twice concurrently with the same IR | Both results are identical and no shared mutable state or race is reported |
| CG-23 | Attempt to edit a generated statement outside a slot | The gate rejects it or produces a deterministic overwrite diagnostic; it never silently preserves it as truth |
| CG-24 | Feed legacy markers through the migration path | Migration is explicit, deterministic, and does not lose slot bodies or outside text |

The current candidate tests cover the core of CG-01, CG-03, CG-05 in a reduced form,
CG-09/10, and part of CG-13. The repository freshness script covers fresh-run
determinism, gofmt, and marker presence, but not previous-source patching. It should
remain a smoke check rather than the complete acceptance gate.

Suggested test organization, without requiring this research branch to edit the
generator, is:

```text
internal/generator/generator_test.go   # normalization and fresh projection
internal/generator/locality_test.go    # previous-source patching and byte locality
internal/generator/markers_test.go     # marker grammar and fail-closed behavior
internal/generator/imports_test.go     # canonical imports and compile fixtures
internal/verify/conformance_test.go    # cross-package round-trip and evidence laws
scripts/generated-freshness.sh         # deterministic repository smoke test
```

## Failure modes and required diagnostics

The generator should classify failures by whether the input is invalid, the previous
artifact is unsafe to patch, or the projection would be semantically ambiguous. A
stable prefix makes CI and agents able to act on the result without parsing prose.

| Failure mode | Why it is dangerous | Required behavior |
| --- | --- | --- |
| Display-name-derived marker ID | A rename moves or destroys handwritten logic | Use stable semantic ID; reject missing/changed identity |
| Stale region after declaration deletion | Removed business behavior remains executable | Remove owned region or stop for explicit migration |
| Marker ID retagging | A body can be stolen by another activity | Verify ownership/provenance; fail closed |
| Non-slot edit inside generated region | Manual structural change disappears on regeneration | Reject or report protected-region edit before writing |
| Unterminated/nested marker | Region boundaries become ambiguous | Return deterministic marker error and write nothing |
| Duplicate IDs or cross-kind collision | Map overwrite can silently drop a type/function | Reject globally before rendering |
| Slot ID reuse | Implementation for one activity appears under another | Require stable owner and migration record |
| Default body overwrites handwritten body | Regeneration destroys irreducible logic | Previous slot body always wins |
| Whole-file gofmt on patch | Unrelated handwritten code churns | Format fresh owned blocks only, or record explicit policy |
| Input-order-dependent imports | Cache misses and noisy diffs | Sort by stable canonical key |
| Alias collision or stale import section | Generated code fails to compile or binds wrong package | Resolve/reject aliases and own the import section |
| Port reordering by adapter | Go signature no longer matches DSL semantics | Preserve declared semantic order |
| Invalid Go name/type | Text may be emitted but cannot build | Validate before writing any output |
| Source-map range before formatting | Navigation/evidence points to wrong code | Map final bytes only and bounds-check every range |
| Unknown previous markers | Generator may destroy another tool's region | Preserve only with explicit ownership; otherwise stop |
| Partial write after validation failure | Repository is left in a mixed generation state | Render and validate in memory, then atomically replace |

Diagnostics should include the rule, stable ID, source file or previous artifact, and
whether the failure is recoverable by changing DSL, adding migration metadata, or
repairing generated output. Example shapes:

```text
gooo:marker:duplicate-region id="billing://activity/pay-order"
gooo:slot:owner-mismatch slot=".../implementation" region=".../other"
gooo:import:ambiguous path="example/time" aliases="clock,time"
gooo:locality:protected-edit id="billing://activity/pay-order"
gooo:stale-region:requires-migration id="billing://activity/old-name"
```

The exact wording is not fixed, but the category and stable ID are part of the CI
interface. A guardian should be able to decide whether to repair the previous file,
update the DSL, or request a scope expansion without reading a generated diff alone.

## Verification and integration gate

The acceptance sequence should be deterministic and layered:

1. Normalize and validate IR without touching the filesystem.
2. Render fresh output in memory; parse it, format-check it, and compile a fixture.
3. Patch a previous composite file in memory; validate marker ownership, locality,
   source-map ranges, and compilation.
4. Compare semantic deltas and generated-region IDs against the allowed scope.
5. Write output atomically and emit provenance/evidence for the input hash, generator
   version, output hash, and verification results.
6. Run the repository checks required by `AGENTS.md`: `gofmt -w .`, `go vet ./...`,
   `go test ./...`, the race suite, and `go run ./cmd/gooo check examples/billing/main.gooo`.

For the current repository, the generator conformance test is build-tagged and the
integration branch may not contain every implementation package at the same time.
That is a sequencing issue, not a reason to weaken the gate. Once the generator and
its dependencies are integrated, `scripts/generated-freshness.sh` should be a
required smoke check and the CG matrix should run in the semantic conformance job.

The final review question is not “did code get generated?” It is:

```text
Did the same semantic IR produce the same valid Go projection,
did regeneration preserve only the authorized handwritten parts,
and is every changed byte or semantic ID inside the declared locality?
```

If any answer is unknown, the safe result is a failed generation with evidence, not
an apparently successful but unverifiable rewrite.
