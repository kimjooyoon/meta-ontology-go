# Gooo 0.6.15 candidate regression plan

Frozen before release implementation and candidate execution. The rejected-fill
feature is merged at `c38201f2bebe401736ac1cb759c002480b7589c3`. The fixtures below
retain actual clean `7c9d8d91` observations from the original feature experiment.

1. Keep the existing four-platform witnesses and their original cases. Add native
   checks for `examples/caller-fill-rejection`: three attempts retain a partial
   result (one executed combination and two rejected assignments, final 1/4);
   five attempts include three executed combinations, two rejections and final
   4/4. Saved replay preserves the complete history with zero new inference.
2. Check rejected IDs, stages, reasons and every hole expression. Retain the
   source/plan identities and all five assignments in the budget. Rejected rows
   must carry no local score or native execution. Initial preparation must retain
   three scored candidates and two rejections; holdouts stay outside selection.
3. Use the original construction and evaluation cases, including the integer
   above 2^53. Mutations of rejection, counts, source, plan, selected code, actual
   values and replay evidence must fail the release witness. Freeze the raw
   partial/complete/replay observations before implementing this witness.
4. Bind version 0.6.15-dev to Go 1.27.1 and decision runtime v0.2.26-experimental.
   Run the new witness locally, then the four existing native CI platforms.
   Exercise the unchanged graph QAT model on the mixed fixture and preserve
   separate deterministic, model-guided, partial and replay observations.
5. Verify canonical exact-head CI, merge normally, and verify the merged dev
   readiness before publishing a new immutable prerelease. Verify the actual
   public archive and install it; repeat the new path and saved replay using it.
   Promote the exact dev tree using the existing six checks and normal merge.

No new training or model download. The graph model orders record choices; fill
assignments remain source-owned. Report finite observations and separate holdouts
without inferring new-program accuracy. Keep prior release records unchanged.

Fixture provenance: `docs/research/caller-fill-rejection-20261009/` contains the
clean build identity, original raw observations, and their checksum manifest.
