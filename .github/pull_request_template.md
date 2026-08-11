## Summary

<!-- State the authoritative view changed and the semantic contract it preserves. -->

## Validation

- [ ] PR targets `integration` (unless this is the integration-to-main promotion).
- [ ] `gofmt -l .` is empty.
- [ ] `go vet ./...` passes.
- [ ] `go test ./...` passes.
- [ ] `go test -race ./...` passes.
- [ ] Semantic round-trip, evidence, scope, and generated-freshness checks pass.
- [ ] The staged verifier level is identified; no level is promoted without
      parity, independent-evidence, reproducible-build, and rollback gates.
- [ ] No generated region was hand-edited.

## Review boundary

- [ ] Changes stay within the declared ownership scope.
- [ ] Any semantic identity, provenance, or policy impact is described above.
