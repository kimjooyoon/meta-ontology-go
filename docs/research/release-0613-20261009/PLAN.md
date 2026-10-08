# Gooo 0.6.13 candidate regression plan

Freeze this plan and release implementation before building the candidate.
The IR-search feature is merged at `934aed807ee6918682b51d0de74bf583e4e61ecf`.
Rejected-expression continuation is submitted as PR 1388, and its original
failure and unchanged-model observations are already published.

1. Bind a clean Go 1.27.1 candidate to its exact source commit and development
   version 0.6.13, retaining decision runtime v0.2.26-experimental.
2. Preserve all existing native release witnesses. Add the frozen scalar
   rejection example: three combination attempts, one unscored local rejection,
   two natively executed programs, original local zero case, four exact evaluation
   rows, then saved-history replay with no inference and the same generated code.
   Keep structural release denominators separate from these language regressions.
3. Run native macOS arm64 readiness, including deterministic builds/archives and
   all prior language/package/source-graph checks. Retain both raw search outputs.
4. Reuse the frozen mixed rejection example with fixed and own-model ordering.
   Preserve its construction and independent evaluation cases. Also retain the
   two-attempt scalar partial outcome. Compare exact actual values and final
   selected source with the earlier clean ae3890c2 observations.
5. Reuse unchanged workbench graph QAT metadata SHA-256
   `3c68205a660695103712e2115ce90accd298b406dced8ad5ad5c9b84bab0a202`
   and weights SHA-256
   `76f68845a03ed8c8bc261a57c919e96dcdd352d3c35bcdaeb3d9b51a88975c6f`.
   Re-run the saved model history without a model option. This does not retrain
   weights or establish new-program generalization.
6. Preserve native original outputs, comparison code, partial outcomes and build
   identity. Do not store binaries or duplicate model weights in the repository.
7. Check the exact final PR source in CI, including all four native platforms.
   After merge, rerun readiness at the merged dev source, publish a new immutable
   prerelease, verify the downloaded archive and install it. Exercise the new
   command and saved replay from the actual public archive before claiming release.
8. Promote the exact current dev tree to main using the existing CI proof and
   normal expected-head PR merge. Update ecosystem/wiki source and release links.

No host CPU utilization or controlled performance benchmark is claimed. Local
source-search rejection continuation does not cover native panic recovery,
all-invalid initial local construction or multi-hole joint search.
