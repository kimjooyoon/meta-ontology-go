## Summary

<!-- State the authoritative view changed and the semantic contract it preserves. -->

## Validation

- [ ] Feature PR targets `dev`; promotion targets `main` from either the exact
      `dev` head or an exact-`dev`-tree snapshot with a direct `main` parent.
- [ ] `gofmt -l .` is empty.
- [ ] `go vet ./...` passes.
- [ ] `go test ./...` passes.
- [ ] `go test -race ./...` passes.
- [ ] Semantic round-trip, evidence, scope, and generated-freshness checks pass.
- [ ] The staged verifier level is identified; no level is promoted without
      parity, independent-evidence, reproducible-build, and rollback gates.
- [ ] No generated region was hand-edited.

## Deterministic CI boundary

- [ ] Changes stay within the declared ownership scope.
- [ ] Any semantic identity, provenance, or policy impact is described above.
- [ ] Acceptance uses exact-head CI evidence for the current PR tuple; human reviews are not CI proof inputs.
