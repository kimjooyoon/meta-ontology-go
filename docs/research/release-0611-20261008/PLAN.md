# Gooo 0.6.11 release regression plan

Freeze this plan with the release implementation before executing the clean
candidate. The earlier source-metadata feature is merged as
`d3552fa2ccee215ff03c98de25263ab698d0a80f`.

1. Build the exact clean 0.6.11 candidate with Go 1.27.1 and retain its build identity.
2. Run the native platform witness, including the new paired explicit/source
   metadata construction and replay. Its diagnostic has four input rows and
   two named expectations per row. Retain raw receipts and independent expected
   values; require unchanged generated code and zero runtime/replay predictions.
3. Reuse standard-library source `27c0f6a0610f44fccd76758640933e1a8b0b99f8`,
   its preview consumer, ten evaluation inputs and three input-only inputs.
   Copy the manifest twice and remove only names/imports from the second copy.
4. Use the unchanged graph QAT model from workbench source
   `19485bb64278ab8ac219e042f92af677ff3f5051`, at
   `models/graph-chooser-20261008/all-data-demonstration/qat_ternary/model.json`.
   Metadata SHA-256 is
   `3c68205a660695103712e2115ce90accd298b406dced8ad5ad5c9b84bab0a202`;
   weights SHA-256 is
   `76f68845a03ed8c8bc261a57c919e96dcdd352d3c35bcdaeb3d9b51a88975c6f`.
5. Compare explicit/fixed, source/fixed, source/model and saved model replay.
   Require equal normalized graphs, native traces, generated program and ten
   expected results. Keep original manifest digests distinct. Record actual
   construction attempts and model predictions without assuming a model benefit.
6. Replay the selected source-derived program on the three input-only rows;
   require OBSERVED, no finite score, actual output values and no new inference.
7. Retain raw observations and exact-number Go comparison, excluding binaries
   and model weights. Keep candidate, public release and installed binary
   identities separate. Existing public 0.6.10 remains installed until a new
   published archive is verified.

These are previously observed programs with fixed sources, cases and weights.
No new training or generalization claim follows from these release regressions.
One local inference timing is descriptive; host CPU utilization and the model's
incremental process memory are not measured by this plan.
