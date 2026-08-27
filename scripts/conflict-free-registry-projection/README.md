# Conflict-free registry projection

Run from the repository root with Go 1.27.0:

```sh
go run ./scripts/conflict-free-registry-projection generate
go run ./scripts/conflict-free-registry-projection check
go run ./scripts/conflict-free-registry-projection prove
```

`generate` writes only the generated projection directory. `check` never writes
and rejects stale output. `prove` runs the bounded vertical-slice evidence and
reports raw versus semantic digests, integration conflict metrics, failure
diagnostics, claim transitions, repository net state, and generated output
metadata. The independent consumer reconstructs the projection from raw local
manifests:

```sh
go run ./scripts/conflict-free-registry-projection-consumer
```

The producer and consumer intentionally do not import one another.
