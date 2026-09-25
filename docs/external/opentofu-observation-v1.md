# OpenTofu released-CLI observation v1

This experiment observes only the pinned OpenTofu v1.12.6 Linux AMD64 release
through its released CLI. It does not import OpenTofu source or internal Go
packages. The external boundary is the immutable tarball, its SHA256SUMS file,
`tofu version -json`, providerless `init`/`plan`/`show -json`, and a plan-only
`tofu test -json` event stream.

The observer has exactly three user paths: `P1 RELEASE_IDENTITY`, `P2
PLAN_JSON`, and `P3 TEST_JSON`. It performs no apply, infrastructure action,
credential access, deployment, or remote-backend operation. Every OpenTofu
working directory, state, plan, event stream, and report is created below a
caller-owned CI temporary directory. The checked-out repository is observed
read-only and must report zero writes.

The contract contains twelve cells and the Gooo program contains exactly twelve
activities, one for each cell. The proof partition is four FOUNDATION, four
COHERENCE, and four REGRESSION cells. The indicator partition is four DRIVER,
four OUTCOME, and four GUARDRAIL cells. A second execution is deterministic
replay, not semantic cache reuse. The first execution records
`discovered=1, executed=1, reused=0, skipped=0, prior_candidates=0,
invalidated=0`; without an exact prior receipt, reuse is not claimed.

The observer preserves the released CLI's Terraform-compatible JSON field names
at the raw boundary. It records normalized event and plan digests only for
replay comparison and does not infer the Go toolchain used to build OpenTofu.
The Go 1.27 toolchain is the observer's toolchain only. The version command's
`terraform_version` and `platform` fields are retained as observed data.

The relevant released CLI contracts are the OpenTofu documentation for
[`version -json`](https://opentofu.org/docs/cli/commands/version/),
[`test -json`](https://opentofu.org/docs/cli/commands/test/), and
[`show -json`](https://opentofu.org/docs/cli/commands/show/).
