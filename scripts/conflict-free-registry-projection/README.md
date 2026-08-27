# Manual-source-registration-edit-free registry projection

Run from the repository root with Go 1.27.0:

```sh
go run ./scripts/conflict-free-registry-projection generate
go run ./scripts/conflict-free-registry-projection check
go run ./scripts/conflict-free-registry-projection prove
```

`generate` writes only the generated projection directory. `check` is read-only
and rejects stale output. `prove` reports three separate integration surfaces:
human-edited existing shared source, generator-changed shared output, and
production consumer adoption. Its proof also records the measured 12 source
digests, all 8 generated output paths/digests/bytes, raw-vs-semantic digest
interventions, source denominator reconciliation, fail-closed diagnostics, and
claim transitions. The independent consumer reconstructs the projection from
raw local manifests without importing the producer:

```sh
go run ./scripts/conflict-free-registry-projection-consumer -check-generated
```

The producer and consumer intentionally do not import one another.
