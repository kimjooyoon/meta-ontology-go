# Rejected assignments remain visible while construction continues

The frozen plan and added assignments preceded implementation in commit
`f908a066`. Published 0.6.14 / clean `2162809f` stopped at `bad_cap_type` before
trying the remaining valid assignments. `baseline.*` retains that failure.

The observed implementation is clean `7c9d8d911b7b9f6c95b7f417fd3715fb92f9f051`,
built with Go 1.27.1. This source experiment is newer than the installed 0.6.14
release even though its development version string has not yet changed.

| Fresh construction | Attempt limit | Attempts | Rejected | Native combinations | Final named outputs |
| --- | ---: | ---: | ---: | ---: | ---: |
| Budget, partial | 3 | 3 | 2 | 1 | 1/4 |
| Budget, complete | 5 | 5 | 2 | 3 | 4/4 |
| Mixed, deterministic | 40 | 40 | 16 | 24 | 4/4 |
| Mixed, own graph model | 40 | 5 | 2 | 3 | 4/4 |

Each completed local fill has three scored assignments and two rejected ones.
Rejections are a Boolean in an Integer position and training-time division by
zero. An unscored rejection contributes neither a case denominator nor a native
execution. The original five assignments still define the whole-program space;
three binary record choices make the mixed space 40.

The unchanged workbench `all-data-demonstration/qat_ternary` graph model made one
prediction to order the record choices. Fill selection stayed deterministic.
Its measured prediction duration was 29,958 ns in this run. The complete mixed
command took 4.81 seconds, including compilation, replay and case execution;
the retained `.time` file reports 0.60 seconds user CPU, 0.41 seconds system CPU,
and 86,933,504 bytes maximum RSS. Other verification ran concurrently. These are
single-run process observations, not whole-machine utilization or a controlled
speed benchmark. There was no new training or model download.

Both new saved histories replayed with zero new predictions and 4/4 final
outputs. The published 0.6.14 budget and mixed histories also replayed unchanged
with zero new predictions and 4/4 outputs. Exact original expectations, including
the integer above 2^53, are retained. Local holdouts remain separate from ranking.

The workbench Gooo feedback loop consumed the original boundary counterexample,
then used budgets 1, 1, 2, 4 and 8. Its five rounds consumed 13 total attempts,
including repeated work, and its final separate evaluation matched 4/4 outputs.
It preserved the source and original counterexample. `workbench-loop.json`
retains the round accounting; the workbench repository retains the raw loop and
the Go recount program using exact JSON numbers.

## Boundaries and verification

Core fill, composition and decision-route packages passed the targeted local
suite before CLI experiments. Regression cases include scalar/record source and
external plans, all-invalid and singleton sets, no-inference singleton selection,
optional model filtering, separate holdout errors, cancellation, rejected-prefix
accounting and altered-history replay. Original valid histories retain v4;
rejected fill histories use v5.

A full local `go test ./...` also exposed a CLI test expecting the old terminal
failure; that test was extended to check both singleton success and all-invalid
failure. That full run did not pass: macOS namespace replacement, callback-preview,
source-splitter and bounded native-child failures were also reported. Changed-area
reruns and Linux CI are recorded separately; no full local-suite pass is claimed.

Malformed contracts, cancellation, holdout-evaluation and native execution errors
remain terminal. Two related programs and their finite cases do not establish
general model accuracy or language-wide correctness. The all-data demonstration
model is not an independent holdout model.
