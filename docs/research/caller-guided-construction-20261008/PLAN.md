# Caller-guided construction: frozen local observation

This plan and the example sources are committed before calling the independent
graph model. The compiler remains a development candidate until CI, merge and
a subsequent release establish those separate stages.

## Question

Can actual caller failures reopen a locally complete record-choice helper,
while preserving its original finite obligations and bounded alternatives?
Can saved construction replay without a model and distinguish consumed caller
inputs from subsequent root tuples?

The simple example reproduces an existing known gap, so it is a regression.
The three-choice example is one newly authored demonstration. It is not an
independent task corpus or a training/generalization study.

## Fixed inputs and model

- `examples/caller-guided-construction/main.gooo.fixture`: one helper, two
  alternatives, one local zero example, caller `3 -> 6`, four evaluation rows.
- `model.gooo.fixture` in the same directory: one helper with three binary
  fields; local zero example; caller `3 -> 15`; seven evaluation rows.
- The files named `construction-cases.json`, `evaluation-cases.json`,
  `model-construction-cases.json`, and `model-evaluation-cases.json` are frozen
  alongside the source before model use.
- Existing workbench model `graph-chooser-20261008/all-data-demonstration/qat_ternary`,
  unchanged from public workbench source `19485bb64278ab8ac219e042f92af677ff3f5051`.
  Model metadata SHA-256: `3c68205a660695103712e2115ce90accd298b406dced8ad5ad5c9b84bab0a202`.
  Weights SHA-256: `76f68845a03ed8c8bc261a57c919e96dcdd352d3c35bcdaeb3d9b51a88975c6f`.

## Paired procedure

Run the three-choice example at program budgets 1, 2, 4 and 8 under deterministic
and model ordering, using the same clean compiler, Go 1.27.1, source and cases.
These are eight runs of one program, not eight programs. Retain all attempts,
local values, caller outputs, initial model context/prediction and resource
observations. Replay the budget-8 model construction after deleting the temporary
model copy. No model is retrained or updated.

Separately run the simple two-choice regression and replay it. Keep its results
separate from the three-choice demonstration.

## Required readings

Report initial local candidate checks, whole-program attempts and saved replay
as different work. Record construction and evaluation numerators/denominators
separately. The three-choice evaluation has one consumed caller root tuple and
six other tuples; zero is also a local helper example. Do not label all seven
as unseen. Input count is not probability of matching arbitrary intent.

Report whether either ordering reaches complete local and caller obligations
within each budget. Both failure and contradiction tests must keep partial
observations. Replay must preserve the selected source and actual values with
zero new inference. The model only ranks initial local choices in this route;
whole-program feedback advances the retained bounded order without new inference.

Use the retained prediction/setup times and native process observations. Host
CPU utilization is not sampled in this study. No general speed or memory
advantage follows from this small, cache-sensitive local observation.
