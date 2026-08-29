# OpenTofu released-CLI observation

This experiment observes only the immutable OpenTofu v1.12.6 Linux AMD64
release through `version -json`, a providerless `init`/`plan`/`show -json`
path, and a plan-only `test -json` path. The fixture has no provider, backend,
credential, apply, deployment, or remote mutation capability.

The three user paths are `P1 RELEASE_IDENTITY`, `P2 PLAN_JSON`, and
`P3 TEST_JSON`. The raw Terraform-compatible JSON field names emitted by the
released CLI remain at the boundary; the observer records their schema rather
than rewriting them. All command output and generated artifacts live in a
caller-owned CI temporary directory.

The fixed contract has twelve cells and twelve corresponding Gooo activities:
four FOUNDATION, four COHERENCE, and four REGRESSION proofs, with four DRIVER,
four OUTCOME, and four GUARDRAIL indicators. A second identical invocation is
counted as deterministic replay, never as semantic cache reuse. This contract
does not infer the Go version used to build the released OpenTofu binary.
