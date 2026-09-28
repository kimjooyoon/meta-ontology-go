# CI proof workflow

Each run is bound to one exact repository, event, base/head revision, workflow,
run, and attempt. The proof contains exactly six canonical status checks:
`gofmt`, `go vet`, `go test`, `go test -race`, `Semantic conformance`, and
`CI policy`.

The `agent/* -> dev` route requires no per-branch path registration. Every PR
runs all six canonical checks against its exact head. Promotion uses a
same-repository `main` PR with either `head=dev` or the source-SHA-named
`agent/main-promotion-snapshot-<dev-sha>` branch. A snapshot is accepted only
when its tree equals the live `dev` tree and its sole parent is live `main`.
The promotion proof re-reads the PR identity and refs, verifies fast-forward
topology, and binds the result to all six successful checks and current
artifacts. It emits a digest-bound `promotion_authorization`; it does not write
branches or branch rules.

CI decisions do not use reviewer identities, approval actors, last-push
approval, Guardian outputs, a special one-time route, or a second branch-rule
observer. GitHub branch rules enforce the six status contexts outside the proof
producer. An unavailable or mismatched source, job, artifact, scope, or
provenance record is reported as unknown or rejected and cannot become PASS.

## Scope and failures

The active PR route derives its subject from the exact base/head tuple and does
not consult branch-to-path ownership tables. The legacy ownership records are
not required to create a feature branch and cannot stop its CI run.

The `gooo/ci-failure/v2` report binds terminal failures to the exact run and
attempt. It records the complete failure-code set, proof rejections, artifact
status, ordered terminal failures, missing reasons, evidence references, and a
closed machine `next_operation`. It contains no human handoff or branch-owner
assignment. An unclassified terminal job fails closed. A successful health
closure is not promotion authorization; only an exact promotion proof can emit
that record.

Local workflow checks are listed in `AGENTS.md`. CI itself runs the canonical
six jobs and the proof-producing steps inside the policy job.
