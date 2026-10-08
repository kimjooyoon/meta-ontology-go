# Gooo 0.6.14 candidate regression plan

Frozen before release implementation or candidate execution. Caller-guided
source fill is merged at `e37a34cff9998d4b69c464e8647958a7662bf5b4` (PR 1391).
The feature's six direct observations came from clean `e115bb77`; its own-model
weights and original cases remain unchanged. Workbench main `e2706442` consumes
v4 and its Gooo feedback loop has completed four rounds/seven combinations.

1. Bind a clean Go 1.27.1 candidate to its exact commit and 0.6.14-dev version,
   retaining decision runtime v0.2.26-experimental. Preserve old release witnesses.
2. Add native multi-hole release checks on all four platforms: the frozen budget
   program at two attempts remains partial (final 1/4), at three attempts chooses
   the complete assignment (final 4/4), and saved replay reproduces all attempted
   assignments and final values with no inference. Check original hole contents,
   local training/holdout values, exact large integers, source identity, candidate
   order and native outputs. Keep holdout counts outside the selection denominator.
3. Retain three raw CLI outputs per platform. Regression fixture tests use the
   previously observed e115bb77 results, not invented successful outputs. Altered
   counts, values, candidate/hole identities, selection and replay flags must fail.
4. Run native macOS arm64 readiness, including deterministic builds and archives.
   Re-run the same two source-fill programs: small-budget partial, complete,
   saved replay, mixed fixed order, mixed unchanged own-model order and its saved
   replay. Compare exact values and selected sources with the feature observation.
   The graph model orders record choices; fills remain deterministic in this run.
5. Use the retained graph QAT model: metadata SHA-256
   `3c68205a660695103712e2115ce90accd298b406dced8ad5ad5c9b84bab0a202`, weights
   `76f68845a03ed8c8bc261a57c919e96dcdd352d3c35bcdaeb3d9b51a88975c6f`.
   No training or model downloads. Preserve all original cases and partial results.
6. Check final candidate source in authoritative PR CI and four native platforms.
   Merge normally using the expected head. Run readiness against the merged dev
   source, publish a new immutable prerelease, verify all assets/checksums, install
   the actual public macOS arm64 archive, and exercise the new path and saved replay.
7. Promote the exact current dev tree to main using existing six-check CI proof
   and normal expected-head merge. Update workbench, wiki and release/source links.

Report finite programs, fresh constructions, saved replays and ordinary source
reuse separately. No host CPU benchmark, new-program accuracy or all-domain
correctness claim. Source-fill preflight still rejects invalid assignments;
native errors remain terminal. No new mandatory reviews or Guardian checks.
