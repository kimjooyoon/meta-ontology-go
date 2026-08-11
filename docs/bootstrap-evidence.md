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

The semantic evidence model also distinguishes `compiler-run`, `verification`,
and `comparison` records. Evidence is append-only. Producer identity, source
span, and audit metadata may differ between hosts; the normalized semantic claim
must still be comparable.

## Comparable evidence envelope

Each host emits one envelope for the same pinned source, toolchain, fixture set,
and policy revision. The following is a contract shape, not a claim that the
current CLI already emits JSON:

```json
{
  "schema": "gooo/bootstrap-evidence/v1",
  "stage": "go-hosted-baseline",
  "producer": "gooo://host/compiler/go",
  "verifier": "gooo://host/verifier/go",
  "source_digest": null,
  "semantic_digest": null,
  "provenance_digest": null,
  "decision": "deferred",
  "evidence_status": "deferred",
  "promotion_eligible": false,
  "checks": {
    "format": "pass",
    "vet": "pass",
    "test": "pass",
    "race": "pass",
    "semantic_cli": "deferred",
    "bootstrap_compare": "not-run"
  }
}
```

`null`, `deferred`, `not-run`, `candidate`, or `mismatch` is never equivalent to
`pass` for promotion. The Go-hosted fixture records the current baseline shape:
ordinary Go checks can pass while the semantic CLI remains explicitly deferred.
The gooo-hosted fixture uses `not-run` and `promotion_eligible: false` because
that stage is not implemented here.

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
set must produce two envelopes and an independent comparison before any
gooo-hosted result can become authoritative.
