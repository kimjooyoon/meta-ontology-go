# Helper-fill release observation

The two compressed JSON files preserve one construction and one saved replay
of `examples/caller-fill-history` by clean development compiler
`87caaa692603c0fc1ff5db44010ab157347cce70`, version `0.6.30-dev`, Go `1.27.3`,
decision runtime `0.2.41-experimental`, on macOS arm64. The source and both case
files were checked before and after the commands. No model was configured.

This producer includes fill preflight capture. It retains the failed `fails`
assignment entering `Seed(2)` and `Seed(0)`, with one completed case out of two
started cases and no completed candidate score. The selected `works` assignment
matches 2/2 local training and 1/1 separate holdout cases. Its preflight, selection
and caller-selected passes each retain their own original helper arguments.

Native evaluation and saved replay each match 3/3 supplied rows. Input exposure
remains two overlapping inputs and one unknown input. These are finite program
regressions, with no model quality or controlled performance result.

The release reader checks the original failure prefix, helper identity, source
checkpoint, selected assignment, separate local scores and native outputs against
the repository's independent case file. Replay must retain the complete original
construction value. Command-wiring tests reuse these bytes and do not execute a
compiler. Native release jobs run the profile with their own accepted build.

These files are a new producer/input release regression. The earlier observation
by `47cb0221` remains unchanged in the example documentation. Original model
studies and the earlier producer's measurements were not repeated.
