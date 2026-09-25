# Bootstrap comparison fixture

`main.gooo` is the shared semantic input for Go-hosted and future gooo-hosted
bootstrap runs. The paired JSON files are evidence-shape fixtures, not recorded
successes:

- `go-hosted-baseline.json` records ordinary Go checks as passing while the
  semantic CLI remains explicitly deferred;
- `gooo-hosted-proposed.json` records the future host as not run and not eligible
  for promotion.

Use [docs/bootstrap-evidence.md](../../docs/bootstrap-evidence.md) for the
comparison rules. Do not replace `null`, `deferred`, or `not-run` with a pass
value until an independent verifier has produced and compared the evidence.
