# Release documents come from Gooo source

The published `v0.6.36-dev` archives are historical one-file archives. Their
checksums and tag stay unchanged. The `v0.6.37-dev` candidate uses
[`payload.gooo`](../../internal/meta/languagereadiness/toolchainrelease/payload.gooo)
to choose the archive member paths, direct-module versions and expected license
text hashes. This is a fixed Gooo activity; it makes no model call.

For each of Linux amd64, macOS amd64, macOS arm64 and Windows amd64, the release
builder now proposes five files: `gooo` or `gooo.exe`, `LICENSE`, the two direct
module licenses under `licenses/`, and `NOTICE`. `NOTICE` maps each license to its
module, exact version or commit, upstream source URL and SHA-256. Its bytes and
the policy source/projection digests are included in the v2 platform receipt.
The receipt also records each archive member's role, path, size, mode, digest
and source. The four platform archive checksums still enter the release bundle's
`SHA256SUMS`.

The direct modules in this source revision are:

<!-- RELEASE-PAYLOAD-MODULES:BEGIN -->
| Module | Pinned source | License text SHA-256 |
| --- | --- | --- |
| `gooo-decision-runtime` | `v0.2.51-experimental` / `a1e755ba422faefda757d61c2fcd45e1d56453a9` | `3ad2cd8fe84a937a0005a2934e377432f2f86fe10ff86c4242cc48f49ebf4947` |
| `gooo-jev` | `v0.0.0-20261019011127-aa4cc4ed706a` / `aa4cc4ed706a1ba4866fc5e6a8177b76bd9a7927` | `7e6da5f9095a021486d2c867ce56160223b414c31c5bf7c3f29212c429b4ac05` |
<!-- RELEASE-PAYLOAD-MODULES:END -->

The table is projected from `payload.gooo` with
`go run ./scripts/release-payload-document -mode write`. The corresponding
`-mode check` verifies that the committed table still matches the Gooo source.

The release builder compares the policy with the checkout's direct `go.mod`
requirements and the versions linked into the actual compiler binary. It reads
the committed license texts and refuses missing or changed bytes. Adding a new
direct module therefore requires an explicit policy and archive update.
After producing two byte-equal archives, a separate reader opens each tar.gz or
zip file and checks all five names, modes, lengths and content digests against
the policy and receipt. A recomputed whole-archive checksum cannot hide a
changed or missing license member.

This is release artifact provenance. It does not turn a license text or source
URL into a claim about the behavior of a dependency. The candidate is public
only after its exact-head readiness and publication runs succeed.
