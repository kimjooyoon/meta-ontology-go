# Bootstrap evidence bridge

This document is the small, implementation-facing bridge between the
[self-hosting research note](research/self-hosting.md), the
[staged CI plan](../.github/conformance-plan.md), and the semantic evidence
model. It defines what Go-hosted and future gooo-hosted runs must emit so they
can be compared without treating either host as business intent.

The current Go-hosted baseline remains authoritative for execution and recovery.
The gooo-hosted path is a proposed candidate; it has no promotion authority and
must not be reported as successful before its evidence is actually produced and
independently compared.

## Stable producers

Evidence uses stable producer identities rather than implementation names in
free text:

| Role | Identity | Status |
| --- | --- | --- |
| Go-hosted compiler | `gooo://host/compiler/go` | Current baseline host |
| Future gooo-hosted compiler | `gooo://host/compiler/gooo` | Proposed candidate; not promoted |
| Go verifier / recovery authority | `gooo://host/verifier/go` | Current independent verifier |

The semantic IR uses the URI identities above. The smaller CI comparison
envelope uses the validated producer tokens `go` and `gooo` and keeps producer
identity out of the canonical payload. An adapter records the URI identity in
its surrounding semantic evidence rather than inventing a new payload field.

The semantic evidence model also distinguishes `compiler-run`, `verification`,
and `comparison` records. Evidence is append-only. Producer identity, source
span, and audit metadata may differ between hosts; the normalized semantic claim
must still be comparable.

## Comparable evidence envelope

Each host emits an `internal/verify.EvidenceArtifact` for the same pinned source,
toolchain, fixture set, and policy revision. The implemented schema is
`gooo/evidence/v1`; the old `gooo/bootstrap-evidence/v1` shape with free-form
check fields is not a valid comparison contract.

The artifact has a producer and a producer-independent bundle:

```json
{
  "producer": "go",
  "bundle": {
    "schema": "gooo/evidence/v1",
    "stage": 0,
    "fixture": "examples/bootstrap/main.gooo",
    "decision": "deferred",
    "facts": []
  }
}
```

The allowed producer values are `go` and `gooo`. Stage values are `0` Go
baseline, `1` dual evidence, `2` gooo fallback, and `3` gooo authoritative.
Facts require a non-empty stable ID and kind, are sorted by `id/kind/value`, and
duplicate fact IDs fail closed. A derived manifest records the schema, producer,
stage, fixture, decision, and SHA-256 of the canonical bundle JSONL payload.

`deferred`, `not-run`, `candidate`, `mismatch`, and a missing digest are never
equivalent to `pass` for promotion. The checked-in Go fixture records a valid
Stage 0 artifact while the semantic CLI is deferred. The gooo fixture records a
Stage 1 candidate with `not-run`; neither fixture grants promotion authority.

## Comparison rules

For two runs over the same input snapshot:

- `source_digest`, input/output artifact digests, normalized semantic digest,
  normalized provenance/evidence projection, verdict, and rule IDs must match;
- `producer`, builder/tool digest, source spans, and complete attestation digest
  may differ because hosts are intentionally different;
- a missing digest, deferred check, candidate-only result, or unavailable
  verifier blocks promotion and retains the last accepted host;
- comparison must be performed by an independent verifier. A candidate cannot
  write the manifest it verifies or certify its own promotion;
- every failed comparison is preserved as evidence and does not rewrite the
  source SSOT, previous attestation, or rollback artifact.

The semantic IR remains the comparison form. `.gooo` remains authoritative for
declared intent and stable IDs; generated Go, manifests, caches, and evidence are
derived records. A host change is therefore a producer change, not an automatic
meaning change.

## Stage contract

| Stage | Authoritative verifier | Required evidence | Result in this repository |
| --- | --- | --- | --- |
| Go-hosted baseline | Go verifier and protected CI | Go checks, pinned inputs, source/semantic/provenance digests | Baseline; semantic CLI may be deferred |
| Dual evidence | Go verifier decides; gooo host is shadow | Same fixture results and comparable envelopes | Proposed; not run by default |
| gooo-hosted candidate | Go verifier remains fallback | Reproducible bootstrap, BX/locality, parity, rollback | Proposed; not promoted |
| Promoted gooo host | Independent comparison plus protected policy | Soak, adversarial corpus, recovery rehearsal, approval | Future; not implemented |

Promotion is a state transition in CI, not a field a fixture may set to true.
The promotion job must reject the candidate whenever any required evidence is
deferred or absent.

## Failure classification

Every non-promoting result is assigned one primary class. The raw artifact and
logs remain append-only even when the class is corrected later.

| Class | Examples | Required action |
| --- | --- | --- |
| `schema-invalid` | Unknown schema/producer/stage, empty fixture, duplicate fact ID. | Fail closed; preserve the rejected payload. |
| `semantic-mismatch` | Canonical bundles or normalized facts differ. | Keep Go authoritative; compare the smallest differing fact set. |
| `provenance-mismatch` | Manifest digest, evidence digest, or source binding differs. | Reject promotion; retain both attestations. |
| `deferred-capability` | CLI stub, disabled stage, unavailable candidate host. | Record `deferred`/`not-run`; retain the prior stage. |
| `reproducibility-drift` | Same pinned inputs produce different payload or artifact digests. | Quarantine the candidate and rerun from a clean checkout. |
| `policy-scope` | Wrong PR base, branch owner, changed-path escape, or Go cap. | Fail the policy job; do not reinterpret as semantic evidence. |
| `trust-boundary` | Candidate self-approves, edits the policy, or omits a required rule. | Require the independent Go verifier and a protected review. |
| `rollback-failure` | Last-known-good artifact or manifest cannot be verified. | Stop promotion until recovery is restored. |

## Reusable verification checklist

Before comparing or promoting a host, record:

1. The exact branch, base revision, source/fixture digest, toolchain, policy
   revision, and `GOOO_CONFORMANCE_STAGE`.
2. Format, vet, unit-test, race-test, policy-scope, and Go-cap results.
3. A normalized `EvidenceArtifact`, its derived manifest, and the canonical
   payload SHA-256. `deferred` and `not-run` must remain visible.
4. Independent comparison of canonical bundles, including facts and decisions,
   not only a final boolean.
5. Fallback artifact, rollback verification, and the failure class if any gate
   is missing or mismatched.

Promotion is a CI policy transition. It cannot be enabled by changing a fixture's
decision field or by a candidate verifier writing its own manifest.

## Fixtures and execution

The common semantic input is
[`examples/bootstrap/main.gooo`](../examples/bootstrap/main.gooo). The paired
evidence fixtures are:

- [`go-hosted-baseline.json`](../examples/bootstrap/go-hosted-baseline.json),
  which records the current baseline with semantic CLI evidence deferred;
- [`gooo-hosted-proposed.json`](../examples/bootstrap/gooo-hosted-proposed.json),
  which records an unimplemented candidate as `not-run`.

Run the repository's current semantic entry point with:

```sh
./scripts/semantic-conformance.sh
```

If the baseline CLI is still a stub, the script's explicit deferred result is
expected and is not a self-hosting pass. Once both hosts exist, the same fixture
set must produce two `EvidenceArtifact` values and an independent comparison
before any gooo-hosted result can become authoritative.
