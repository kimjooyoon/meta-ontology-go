# Gooo controls its next construction step

This example runs a small feedback loop whose policy is itself a Gooo program.
The target declares three record-field choices and an initial budget of two
candidate evaluations. Native execution measures the result; `Decide` in
[`policy.gooo.fixture`](policy.gooo.fixture) chooses whether to keep that program,
stop, incorporate a direct-input counterexample, or increase the next attempt budget. Go applies the returned operation,
revises the selected activity's assembly declaration, and executes the next round.

The supplied policy keeps a round when it matches at least as many feedback
expectations as the retained round. It doubles the budget up to the caller's
limit. New counterexamples can become explicit selection cases in the next source
revision. The policy stops when all feedback expectations match or its bounds
leave no next operation.
Changing those expressions changes the policy without editing the Go loop.

## Run

From the compiler repository root, using a newly built `gooo`:

```sh
gooo body-refine \
  --source examples/assembly-feedback/target.gooo.fixture --activity Select \
  --feedback-cases examples/assembly-feedback/feedback-cases.json \
  --evaluation-cases examples/assembly-feedback/evaluation-cases.json \
  --policy examples/assembly-feedback/policy.gooo.fixture \
  --max-attempts 8 --max-rounds 4 --out /tmp/gooo-feedback-run
```

Pass `--go-bin /path/to/go1.27.1` when the default Go is a different version.
Add `--model /path/to/model.json` to rank construction choices with a compatible
local model. Omission uses deterministic ordering. Each construction round loads
its model separately and records its own calls; policy execution is deterministic.

The deterministic regression observes budgets 2, 4 and 8, ending with 14/14
named activity outputs across seven feedback inputs. The final evaluation has
three separate inputs and six named output expectations. Actual results, including
partial outcomes, are saved rather than replaced with a success-only summary.

## What the source controls

The policy receives `matched`, `total`, `best`, `attempts`, `limit`, `round`,
`round_limit` and `counterexamples`. The last field counts new promotable cases
from failed target outputs whose inputs came directly from the caller. `best` is
the matched count of the currently retained round; a
custom policy may retain a lower-scoring round. `Decide` returns:

| Field | Meaning |
| --- | --- |
| `action` | `STOP`, `CONTINUE` or `INCORPORATE` |
| `next_attempts` | Current budget for STOP; a larger budget for CONTINUE; current or larger for INCORPORATE |
| `retain` | Whether this observed program becomes the retained result |
| `reason` | A source-owned explanation |

The adapter bounds requests to eight rounds and 64 attempts per round, honors
cancellation, and rejects a nonadvancing operation. Source IR search also retains
its declared candidate cap. Only the named activity's `assembling` block is
reformatted to change its budget or add explicit feedback cases. Bodies, declared
choices, grammar and semantic IDs are preserved. Matching source holdouts promoted
into selection are removed from the holdout list; conflicting expected answers
produce an error. Upstream computed values cannot become new training inputs.
The original input file is left available;
each revised source is saved as a new round artifact. Ordinary activities in the
same graph run with each candidate program.

## Read and reuse the result

- `refinement.json` retains every round, policy input/output, source, generated
  composition, native trace, model observation and optional final evaluation.
  `feedback_added` and `revised_source` identify the actual contract revision.
- `round-01/`, etc. contain independently replayable composition artifacts.
- `selected/` contains the retained program, generated Go, and its feedback cases.
- `selected_round` is zero-based; directory numbers start at one.

```sh
gooo body-compose \
  --source /tmp/gooo-feedback-run/selected/original.gooo \
  --cases /tmp/gooo-feedback-run/selected/cases.json \
  --composition /tmp/gooo-feedback-run/selected/composition.json
```

Saved replay requires no model. `feedback_status` reports adaptive feedback;
`evaluation_status` is UNKNOWN until separate final cases are run. A failing final
expectation leaves the overall result PROGRESS even when feedback passed. The
command completes successfully for a measured partial result; malformed inputs,
invalid transitions and execution failures return an error with retained evidence.

Feedback inputs affect when to continue and which round to keep, so their success
is an adaptive measurement. The optional final evaluation is executed once after
the retained round is fixed and never reaches the policy. Input separation from
construction cases is still recorded, and model-training exposure remains unknown.
Callers can supply overlapping final inputs; this API does not establish that a
dataset is an independent holdout. The policy's own input-only execution has zero
correctness expectations and is recorded as an observation.

The loop uses already declared choices or expression grammars. A search grammar
can derive new candidates from promoted cases; the expected answers still come
from the caller's feedback document. The loop does not invent expected answers,
adjust model weights, or rewrite arbitrary application bodies. If the declared
grammar cannot express a needed body, the result remains incomplete after its
bounded attempts. Grammar expansion is a separate language operation.
