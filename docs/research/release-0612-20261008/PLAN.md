# Gooo 0.6.12 candidate regression plan

Freeze this plan with the release implementation before building and running
the clean candidate. Caller-guided construction is submitted as PR 1384, source
`04aa6b9613dadb73037e64ea63a2da7437dbcf7a`; its model observation is already public.

1. Build the clean candidate with Go 1.27.1 and retain its exact source identity.
2. Run the native platform witness, including the existing division, retry,
   filename, package and graph checks and the added caller construction.
3. In that regression, require two whole-program attempts; local 1/1 in each;
   caller 0/1 followed by 1/1; four evaluation rows with exact integer values;
   one consumed caller input and three other root tuples. Retain fresh and
   saved command outputs and require equal selected programs and no new replay
   predictions. Keep existing structural release-case denominators unchanged.
4. Reuse the committed three-choice source and evaluation files in
   `examples/caller-guided-construction`. Run fixed and own-model ordering with
   whole-program budgets 1 and 8. Retain partial outcomes and actual values.
   This repeats a known demonstration as a release regression.
5. Use unchanged workbench graph model
   `models/graph-chooser-20261008/all-data-demonstration/qat_ternary/model.json`.
   Metadata SHA-256:
   `3c68205a660695103712e2115ce90accd298b406dced8ad5ad5c9b84bab0a202`.
   Weights SHA-256:
   `76f68845a03ed8c8bc261a57c919e96dcdd352d3c35bcdaeb3d9b51a88975c6f`.
   Delete only the temporary model copy before saved replay, retaining the
   original workbench weights. No training or weight update is part of this run.
6. Recount actual local/caller/evaluation values with exact JSON integers, compare
   completed fixed/model source and replay results, and distinguish the consumed
   caller tuple and helper's local zero example from other evaluation roots.
7. Retain compressed raw observations, independent case files, build identity and
   the Go comparison program, without compiler binaries or duplicate weights.
8. After CI and merge, rerun native readiness from the merged source and publish
   a new immutable experimental tag. Verify the published archive before replacing
   the installed 0.6.11 binary. Execute and replay the new command from that archive.

Historical and candidate observations stay separately identified. These checks
do not measure new training, unseen-program transfer, host CPU utilization or
general performance. The public graph weights remain on GitHub; this release
does not establish a new Hugging Face upload.
