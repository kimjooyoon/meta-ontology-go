# Self-hosting and bootstrap contract

This document defines the bootstrap loop for `meta-ontology-go`. It is a
promotion contract, not a promise that a self-hosted compiler is trustworthy
merely because it can compile itself. The Go implementation remains an
independent reference and recovery boundary until the conditions in Stage 4
are met, and a self-check never replaces an independent trust root.

## Scope and vocabulary

The bootstrap unit is a complete semantic build:

```text
authoritative source -> compiler -> semantic IR -> Go/CI projection
                    -> manifest + attestation + verification evidence
```

The following terms have precise meanings here:

- **SSOT** means the view that is allowed to introduce business or compiler
  meaning. A generated Go file, cache, manifest, or attestation is never SSOT.
- **Semantic equivalence** means equality after normalization of stable semantic
  IDs, declarations, relations, diagnostics, and generated-region boundaries.
  It does not require byte-for-byte equality of source formatting.
- **Provenance** records which source, compiler, inputs, and outputs produced a
  claim. Evidence is append-only; a later build adds a record rather than
  rewriting an earlier one.
- **Fallback** is a previously accepted compiler/verifier artifact that can
  rebuild or verify the candidate. A fallback must be retained as an artifact,
  not just described in documentation.
- **Trust boundary** identifies which implementation is independently trusted
  to make a promotion decision. A candidate cannot be the only verifier of its
  own release.

The small standard-library-only model in `internal/bootstrap` implements the
minimum record shape for this contract. It intentionally has no dependency on
the compiler core, has no timestamps or host paths, sorts artifact lists before
serialization, and hash-chains evidence records.

## Invariants for every stage

Every stage must preserve these invariants before it can be promoted:

1. Stable semantic IDs, not display names or aliases, identify declarations and
   compiler concepts. A rename is valid only when its ID is unchanged.
2. The source view, semantic IR, projection, and evidence describe the same
   semantic delta. Generated output and caches are derived and disposable.
3. A build is reproducible from a pinned source revision, toolchain identity,
   declared inputs, and deterministic configuration. Wall-clock time, local
   absolute paths, directory iteration order, and unrecorded environment values
   must not affect the manifest or semantic digest.
4. A manifest covers every input and output relevant to the build. Its
   `source_digest`, `compiler_digest`, and `semantic_digest` are lowercase
   SHA-256 values. An attestation binds the manifest digest to its builder and
   to an append-only evidence chain.
5. A verifier failure is fail-closed. Missing, stale, malformed, or conflicting
   evidence is a failed build, not an invitation to skip a check.
6. A new implementation first proves equivalence against the old one. Passing
   its own tests is necessary but never sufficient for promotion.

## Stage overview

| Stage | Primary capability | SSOT during promotion | Authoritative verifier | Fallback |
| --- | --- | --- | --- | --- |
| 0 | Go-hosted parser, IR, compiler, and projection | `.gooo` source plus stable semantic IDs | Go host and protected CI | Last accepted Stage 0 artifact |
| 1 | Go host builds a `.gooo` compiler subset | `.gooo` source/spec and frozen corpus | Go host; subset is a candidate | Stage 0 Go compiler |
| 2 | A `.gooo` compiler builds the compiler again | `.gooo` compiler source and conformance corpus | Go verifier plus bootstrap comparison | Stage 1 / Stage 0 Go host |
| 3 | A `.gooo` verifier produces CI evidence | Protected verifier contract and corpus | Go verifier during shadow/soak | Go verifier |
| 4 | Production path no longer needs the Go compiler fallback | Promoted `.gooo` compiler and protected policy | Independently retained recovery verifier | Signed last-known-good release |

The stages are monotonic promotions. A later stage may add a derived producer,
but it may not silently change the SSOT or relax an earlier acceptance test.

## Stage 0 — Go-hosted parser, IR, and compiler

### Contract

- **SSOT:** `.gooo` declarations and their stable semantic IDs are authoritative
  for program meaning. The Go parser/lowerer/compiler is the trusted host for
  interpreting the current language contract. Generated Go, docs, caches, and
  evidence are derived views.
- **Reproducible build:** pin the Go toolchain from `go.mod`; sort source and
  artifact paths; build from a commit-addressed checkout; exclude timestamps,
  absolute paths, and ambient environment values. Two clean builds of the same
  revision must have equal canonical manifest and semantic digests.
- **Semantic equivalence:** establish the current laws in `docs/spec.md`:
  `DSL -> IR`, `IR -> Go`, Go lifting, Get-Put, Put-Get, round-trip, and
  locality. Formatting may differ, but stable IDs, fact kinds, source spans,
  generated markers, and accepted diagnostics must agree.
- **Provenance/evidence:** record source-set digest, Go host/compiler digest,
  input/output artifact digests, semantic digest, parser/lower/project steps,
  generated-region freshness, and test results. Evidence is stored as an
  append-only attestation and is invalid if its manifest binding or chain is
  broken.
- **Rollback/fallback:** use the last accepted Go host and its manifest. A
  failed candidate is discarded with its attestation; no generated output is
  promoted from a failed build.
- **Trust boundary:** protected Go source, the pinned toolchain, and the
  independent CI runner are trusted. The DSL is authoritative for business
  intent, but it is not trusted to rewrite verifier policy.

### Acceptance tests

1. Parser and lowering golden tests cover valid and invalid syntax, source spans,
   namespace-safe IDs, and deterministic diagnostics.
2. Round-trip and locality tests pass for the conformance corpus, including an
   implementation-only Go edit that leaves unrelated generated regions intact.
3. Two clean builds produce byte-identical canonical manifests and equal
   semantic digests; a changed source or compiler digest changes the manifest.
4. Generated Go is formatted, contains stable generated-region markers, and
   passes `go test`, `go vet`, and `go test -race`.
5. Tampering with an input, output, semantic digest, or evidence predecessor is
   rejected by an independent verifier.

Stage 0 is the reference implementation against which all later stages are
compared. It is not removed by self-hosting; its accepted artifacts become the
first recovery package.

## Stage 1 — Go-hosted `.gooo` compiler subset

### Contract

- **SSOT:** the explicitly frozen `.gooo` compiler subset grammar, semantics,
  and conformance corpus are authoritative. The Go host remains authoritative
  for accepting the subset and for the promotion decision. The new `.gooo`
  compiler implementation is a candidate.
- **Reproducible build:** Stage 0 builds the subset compiler from a pinned
  `.gooo` source tree. The manifest includes the Stage 0 compiler digest,
  compiler-source digest, corpus digest, and all produced artifacts. A second
  host must reproduce the same subset compiler and manifest.
- **Semantic equivalence:** for every corpus item, compare the Go-hosted and
  `.gooo`-hosted results after normalization: IR facts, stable IDs, generated
  markers, accepted/rejected status, and diagnostic codes. Unsupported syntax
  must be rejected identically rather than accepted by one side.
- **Provenance/evidence:** emit one evidence record per corpus case plus summary
  digests for parse, lower, project, and rejection parity. The candidate's
  evidence uses the same `gooo.bootstrap/v1` shape as Stage 0 so an independent
  consumer can compare it without importing compiler internals.
- **Rollback/fallback:** keep Stage 0 as the default compiler. If any case,
  digest, or diagnostic differs, quarantine the Stage 1 artifact and retain its
  evidence for diagnosis; do not update the authoritative generated tree.
- **Trust boundary:** the Go host, protected CI, and the frozen corpus are
  independent of the candidate implementation. The candidate cannot edit the
  corpus, the verifier, or the promotion rule as part of its own build.

### Acceptance tests

1. The full subset corpus has zero differential semantic mismatches and zero
   unexplained diagnostic mismatches.
2. Stage 0 can rebuild the candidate twice from the same source and obtain the
   same compiler artifact digest.
3. A Stage 1 compiler cannot compile a feature outside the declared subset
   without producing the Stage 0 rejection result.
4. Source-map spans and stable IDs agree across both implementations.
5. Deliberately perturbing one compiler output causes the independent Go
   comparison to fail and leaves the Go result authoritative.

Promotion at this stage means “safe candidate exists,” not “Go is obsolete.”

## Stage 2 — `.gooo`-hosted compiler builds itself

### Contract

- **SSOT:** the `.gooo` compiler source, its language-level build description,
  and the frozen compiler conformance corpus. Generated Go and the bootstrap
  binary remain derived. The Go host still owns the first bootstrap and the
  acceptance decision.
- **Reproducible build:** use a two-step fixed-point bootstrap. Stage 0 builds
  compiler `C1` from the `.gooo` compiler source; `C1` builds `C2` from the
  same source and pinned inputs; repeat until the declared fixed-point rule is
  met. Record parent manifest digests so the chain is auditable. A clean host,
  stable toolchain, and stable source must produce equal `C2` compiler and
  semantic digests.
- **Semantic equivalence:** compare `C1`, `C2`, and the Go reference on the
  language corpus and compiler corpus. The fixed-point condition is semantic
  first and byte equality second: normalized IR and diagnostics must agree, and
  the compiler artifact must be reproducible when the toolchain permits it.
- **Provenance/evidence:** attest each bootstrap edge (`source -> C1`,
  `C1 + source -> C2`) and include parent manifest digests, compiler digests,
  semantic digests, corpus results, and fixed-point status. The Go verifier
  validates the chain; the self-hosted compiler only produces claims.
- **Rollback/fallback:** Stage 1 and Stage 0 artifacts remain runnable. A
  mismatch rolls back the selected compiler artifact, never the source SSOT.
  Re-bootstrap from the last accepted Go-hosted artifact after repairing the
  candidate.
- **Trust boundary:** the independent Go verifier and protected CI decide
  whether fixed-point evidence is sufficient. A compiler proving that it can
  compile itself is not proof of semantic correctness by itself.

### Acceptance tests

1. A clean two-step bootstrap reaches the same normalized compiler IR and
   compiler digest on at least two independent workspaces.
2. `C2` reproduces the Stage 1 corpus result, including negative tests and
   diagnostics, with no tolerated mismatch list.
3. A third bootstrap iteration is semantically equivalent to `C2`; a change in
   source or compiler input invalidates the parent manifest chain.
4. The Go verifier rejects a tampered parent digest, source digest, or compiler
   artifact before any output is promoted.
5. The bootstrap can be rolled back to the signed Stage 0/1 package without
   requiring the candidate compiler to repair itself.

The fixed-point result is a release candidate. It does not yet authorize the
candidate to define CI policy, because that is a separate trust transition.

## Stage 3 — `.gooo`-authored CI verifier

### Contract

- **SSOT:** the protected verifier contract, policy rules, and conformance
  corpus remain authoritative. A `.gooo` verifier implementation is a new
  producer of evidence, not a new source of policy. Any policy change must be
  reviewed as a policy change, independently of the implementation language.
- **Reproducible build:** both verifiers receive the same checkout, manifest,
  environment allowlist, and corpus. They run in isolated workspaces and emit
  the same canonical evidence schema. The CI job records both builder digests
  and the common input/semantic/evidence digests.
- **Semantic equivalence:** compare the pass/fail verdict, rule identifiers,
  normalized diagnostics, changed-path scope, Go size caps, round-trip checks,
  generated freshness, and evidence digest. A successful self-check that omits
  a rule is a mismatch, not a pass.
- **Provenance/evidence:** store two attestations over the same manifest, one
  from the Go verifier and one from the `.gooo` verifier. Compare the common
  manifest and evidence projection. It is valid for the complete attestation
  digests or `builder_digest` values to differ; it is not valid for the
  semantic/evidence digests or verdicts to differ.
- **Rollback/fallback:** use a feature flag or protected workflow setting. In
  shadow and soak mode, the Go verifier remains authoritative and the `.gooo`
  verifier is diagnostic. Any mismatch blocks promotion of the new verifier
  and retains the Go result. The previous Go workflow remains deployable.
- **Trust boundary:** the protected CI runner, Go verifier, branch policy, and
  review rules are independent of the candidate. The `.gooo` verifier cannot
  approve its own source change, alter branch protection, change the common
  evidence schema, or remove the Go comparison job.

### Required dual-verifier promotion sequence

1. **Shadow:** run Go and `.gooo` verifiers on the same revision and compare
   evidence. Only Go decides the status; all `.gooo` mismatches are retained.
2. **Soak:** make both jobs required, but keep Go authoritative. Require a
   meaningful corpus of positive, negative, scope-boundary, stale-evidence,
   and tamper cases across the agreed release window.
3. **Switch:** after the Stage 3 exit criteria below, make the `.gooo` verdict
   the primary result while the Go verifier remains a required non-authoritative
   canary for at least one more release window.
4. **Retire or retain:** remove only the primary Go invocation if the Stage 4
   conditions are met. Keep a signed Go verifier artifact and a manual recovery
   workflow. A temporary rollback changes the authoritative flag back to Go;
   it never disables both verifiers.

### Acceptance tests and exit criteria

1. Positive and negative corpus cases have identical verdict and normalized
   rule/evidence digests over the soak window; there is no allowlisted semantic
   divergence.
2. Injected failures in each verifier are detected by the other or by the
   protected comparison job. A verifier that returns “pass” with missing
   evidence fails the comparison.
3. Changes to verifier source, policy, corpus, toolchain, or evidence schema
   trigger the expected independent review and rerun both implementations.
4. The `.gooo` verifier cannot write or modify the manifest it verifies, and
   the Go verifier checks the `.gooo` verifier's attestation externally.
5. The common evidence comparison is deterministic across repeated CI runs and
   across at least two supported runner environments.

## Stage 4 — conditions for removing the Go compiler fallback

“Removing the fallback” means that the normal production build invokes the
promoted `.gooo` compiler. It must not mean deleting every Go recovery artifact
or allowing the self-hosted toolchain to become its own sole root of trust.

The normal path may change only when all conditions below are satisfied and
recorded in a signed release attestation:

1. **Bootstrap fixed point:** Stage 2 passes on two independent clean hosts for
   two consecutive release candidates, with no semantic or diagnostic
   divergence from the Go reference corpus.
2. **Reproducibility:** repeated builds reproduce the same canonical manifest,
   compiler artifact, generated-region digest, semantic digest, and evidence
   digest. Any intentionally non-reproducible output is outside the promoted
   artifact and explicitly documented.
3. **Verifier soak:** Stage 3 has completed the agreed soak window (recommended:
   two releases or at least 20 independent CI runs) with zero unexplained
   verdict/evidence mismatches. The Go verifier has been run as the
   authoritative comparator throughout that window.
4. **Adversarial coverage:** corpus tests cover malformed syntax, namespace and
   identity changes, generated-region tampering, stale manifests, changed-path
   escapes, function/file cap violations, race-sensitive checks, and missing or
   forged evidence.
5. **Independent recovery:** a signed last-known-good Go compiler/verifier,
   bootstrap manifest chain, and recovery instructions are archived outside
   the candidate's output. Recovery is tested from a clean checkout.
6. **Protected policy:** branch protection still requires an independent
   verifier or comparison job. A `.gooo` verifier may not change its own trust
   root, CI permissions, or evidence schema without the old independent path.
7. **Review separation:** the builder, verifier/policy reviewer, and release
   approver are separate roles. The same change is not implemented, weakened,
   and approved by one agent.

If any condition stops holding, revert the authoritative flag to the archived Go
verifier/compiler and open a new bootstrap investigation. Do not repair a
failed release solely by asking the candidate to verify or rewrite itself.

## Evidence comparison contract

The dual-verifier job compares these values from the same input snapshot:

| Value | Must match? | Reason |
| --- | --- | --- |
| `source_digest`, input/output artifact list | Yes | Same build boundary |
| `semantic_digest` | Yes | Same normalized meaning |
| verdict and normalized rule IDs | Yes | Same policy decision |
| canonical evidence projection/digest | Yes | Same proof claims |
| `builder_digest` | No | Implementations are intentionally different |
| complete attestation digest | No | It includes builder identity |

The prototype exposes `Manifest.Digest`, `Attestation.Validate`, and
`Attestation.EvidenceDigest` for this comparison. A verifier should first
validate each attestation against the common manifest, then compare the common
fields. Comparing only two final booleans is insufficient: a verifier that
silently skips evidence would otherwise appear equivalent.

## Rollback protocol

Every promotion stores the previous accepted manifest digest, compiler digest,
verifier digest, and evidence location. On mismatch or suspected compromise:

1. freeze promotion and preserve both attestations and logs;
2. switch the protected authoritative flag to the previous accepted Go path;
3. verify the previous artifact against its archived manifest from a clean
   workspace;
4. rerun the differential corpus to classify source, compiler, policy, or
   evidence drift; and
5. resume promotion only with a new candidate and a new review.

Rollback is an evidence-preserving operation. It must not rewrite the failed
attestation or silently replace the source revision it describes.

## Integration risks and mitigations

| Risk | Detection | Mitigation |
| --- | --- | --- |
| Grammar or diagnostic drift | Differential corpus | Freeze subset and compare rejection codes |
| Stable-ID or namespace drift | Semantic digest mismatch | Treat IDs as SSOT; add rename/alias tests |
| Host-dependent output | Repeated/cross-host manifest mismatch | Canonical paths/order; remove timestamps and ambient values |
| Incomplete evidence | Attestation validation/comparison | Fail closed; require common evidence digest |
| Correlated verifier bug | Independent Go verifier and adversarial corpus | Keep old verifier authoritative through soak |
| Self-modifying trust root | Protected workflow diff and review boundary | Candidate cannot edit policy or approve itself |
| Unusable rollback | Clean-checkout recovery test | Archive signed Go artifacts and parent manifests |
| CI scope escape | Changed-path policy | Keep branch ownership explicit and fail closed |

The intended endpoint is therefore a loop with an external anchor:

```text
Go reference -> gooo subset -> gooo fixed point -> dual CI evidence
     ^                                                   |
     +------------- signed recovery and independent review
```

Self-hosting removes an implementation dependency from the normal path; it does
not remove the obligation to preserve an independently verifiable history.
