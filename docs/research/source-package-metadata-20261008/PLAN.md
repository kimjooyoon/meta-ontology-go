# Source-owned workspace metadata: usage plan

Frozen before the candidate compiler is built or connected to the model.

The public standard library exposed a repeated declaration: Gooo package names
and imports had to be restated in the workspace manifest. The development reader
will fill only omitted metadata from parsed Gooo sources. Package paths, source
locations and executable entry remain explicit. Explicit empty imports remain
an assertion and existing manifest-only imports remain supported.

## Checks

1. Compare the entire normalized resolution result with an equivalent fully
   specified manifest; retain different manifest byte digests.
2. Check multiple source files, duplicate imports across files, inconsistent
   headers, invalid source, missing dependencies and cycles. Preserve explicit
   mismatch checks and earlier manifest-only workspaces.
3. Execute imported calls and saved replay after workspace relocation. Continue
   a partially assembled helper with a Gooo policy whose own metadata is omitted.
4. Run the existing library and diagnostic starter checks with the new minimal
   manifests, and the existing resolution regressions.
5. Build a clean candidate and use the previously published standard library at
   `27c0f6a0610f44fccd76758640933e1a8b0b99f8`. Copy the source files unchanged;
   create a second manifest by removing only each package's `name` and `imports`.
6. Compare declared and source-derived resolution, deterministic full-budget
   consumer construction, and the unchanged Workbench graph QAT model from
   `19485bb64278ab8ac219e042f92af677ff3f5051`. Reuse the ten frozen consumer inputs
   as a compiler regression; this is not a new model-generalization study.
7. Replay the source-derived model construction without a model and run its
   three input-only requests. Keep actual values, source/code digests, model
   identity, finite counts and new inference calls. Do not train or change Gooo
   choices/expectations in response to results.

## Current scope

The original source-metadata resolution test failed on
`PACKAGE_IDENTITY_UNKNOWN: path and name are required` before implementation.
The reader change is shared by resolution, execution, saved replay, continuation
and Gooo assembly-policy workspaces. Library/diagnostic starters adopt it.
It does not fetch dependencies or infer the package path from a display name.
The 0.6.10 published binary remains the comparison baseline; the new minimal
manifest needs the candidate compiler until a later public release contains it.
