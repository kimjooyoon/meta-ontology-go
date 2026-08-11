# Bootstrap comparison fixture

`main.gooo` is the shared semantic input for Go-hosted and future gooo-hosted
bootstrap runs. The paired JSON files are evidence-shape fixtures, not recorded
successes:

- `go-hosted-baseline.json` is a valid `gooo/evidence/v1` Stage 0 artifact whose
  decision is explicitly deferred;
- `gooo-hosted-proposed.json` is a valid Stage 1 candidate artifact whose
  decision is `not-run` and therefore cannot be promoted.

Use [docs/bootstrap-evidence.md](../../docs/bootstrap-evidence.md) for the
comparison rules. The JSON shape is an `EvidenceArtifact`; its manifest is
derived from the canonical bundle and is not fabricated in this fixture. Do not
replace `deferred` or `not-run` with a pass value until an independent verifier
has produced and compared the evidence.
