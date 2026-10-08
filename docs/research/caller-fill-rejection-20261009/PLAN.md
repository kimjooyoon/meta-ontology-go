# Continue after rejected complete assignments

Frozen before implementation. Baseline is published 0.6.14 / clean source
2162809f3c49637f0db5d1003873db5e8189d16e. The five-assignment budget fixture retains
all three earlier valid assignments, adding one incompatible value type and one
training-time division by zero. Original construction and evaluation inputs and
expectations are copied unchanged. The mixed program retains its record choices.

- Reproduce the public CLI's terminal failure before implementation.
- Check all declared assignments, retain deterministic type/training-evaluation
  rejections separately from scored candidates, and choose among valid assignments.
  Structural contract/source errors, cancellation and native execution failures
  remain terminal. No fabricated accuracy for unscored rejected candidates.
- Handle zero valid assignments with a complete rejection diagnostic, one valid
  assignment with a recorded deterministic choice and no prediction, and several
  valid assignments with the existing model/default chooser. Preserve source IDs,
  plan binding, all expectations and separate holdouts. Holdouts cannot rank choices.
- Whole-program search retains the original five-assignment space and charges
  rejected assignments to the same attempt limit. Preserve any already scored
  prefix. Replay checks exact candidate identity, rejection reason and stopping.
- Exercise integer, record, source and external plans; malformed declarations,
  all-invalid/single-valid sets, cancellation, optional model use, saved replay,
  altered rejection records, mixed search/record/fill and original valid histories.
- Dogfood the unchanged public own graph model on the mixed program. Compare
  candidate attempts, rejected combinations, actual native executions and exact
  final values. Do not call repeated platform executions independent experiments.
- Connect the new rejection records to the workbench's Gooo diagnostic/feedback
  reader and preserve the original counterexample during an actual automatic loop.
- Publish implementation, fixtures, baseline failure and observations through
  normal CI-checked public PRs. Document supported scope and remaining boundaries.

No model downloads, additional training or new worktrees. No FSM illustrations.
