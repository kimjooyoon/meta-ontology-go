# Helper checkpoint release fixtures

The two compressed JSON files preserve the original constructor/evaluation and
saved replay from clean development compiler `722161a1ac3f7d82d3fedc4aeb3266ee1b0c81ee`.
They were copied from the already completed synthetic observation; no model fit,
candidate construction or original observation was repeated to make these files.
The source and independent case files are in `examples/caller-search-checkpoint`.

These fixtures test the release reader, including changed-source, changed-pool,
changed-helper-input and changed-output counterexamples. Actual release checks
run the candidate binary on each native platform and preserve two new raw command
outputs. Fixture success alone does not establish that a release binary works.
