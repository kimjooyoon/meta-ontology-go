# When all permitted implementations share a wrong answer

This small Gooo program asks for the negative of a refund amount. Its two
permitted expressions, `input` and `input+0`, both preserve the amount. At zero
they satisfy the condition. At every supplied nonzero input they fail together.

The source declares the condition as a fixed Gooo activity:

```gooo
activity Negative(Integer,Box) -> Boolean computes "return input1.value == 0-input0"
```

Inside the assembling activity, it chooses how additional inputs are acquired:

```gooo
criteria { predicate "negative" activity "Negative" input "[0]" holdout_input "[19]" }
observe {
 criteria_strategy "distinguish_then_check"
 criteria_input "[7]"
 criteria_input "[-5]"
 max_candidates "2"
 max_rounds "1"
 reuse_outputs "true"
 resolve_unique "true"
}
```

First Gooo looks for an input that separates the surviving candidates' outputs.
If they agree, it acquires the first remaining source question with an available
output and checks that input against the condition. Only acquired questions are
predicate-scored. The reason for this acquisition is `CHECK_CRITERIA`.

## Recorded observation

| Observation | Distinguishing only | Distinguish, then check |
| --- | --- | --- |
| Initial criterion | 1/1 | 1/1 |
| Acquired input | none | `[7]` |
| Selected candidate's criteria | 1/1 | 1/2 |
| Question outcome | `NO_DISTINGUISHING_INPUT` | `NO_SURVIVING_CANDIDATE` |
| Selection status | `COMPLETE_FINITE` on its one initial obligation | `PARTIAL_FINITE` |
| Separate held criterion | 0/1 | 0/1 |
| Additional native outputs | not run | 0/2 |
| Model calls | 0 | 0 |

The previous distinguishing-only path already reported the held-input failure.
The additional behavior is to acquire a counterexample during construction and
retain it as a local obligation for continuation and caller construction. The
candidate set in this example remains insufficient to implement negation.

The new question loop performed six candidate-output evaluations and four
predicate evaluations. These counts cover the question loop alone. The unchosen
`[-5]` output was inspected, while its predicate was not scored. Selected-body,
held and native evaluations have their own counts.

Native execution completed and its two expected negative outputs both failed.
The command returned exit code zero, meaning the experiment ran to completion;
the recorded `finite_passed: 0` and `finite_total: 2` describe functional success.

## Files and reproduction

- `distinguish.gooo.fixture` contains the source without the explicit strategy.
- `main.gooo.fixture` contains the additional strategy declaration.
- `evaluation.json` supplies two additional native cases, with inputs 23 and -31.
- `observation.json` binds the recorded source, compiler and output digests.

From a checkout containing this implementation:

```sh
go run ./cmd/gooo body-compose \
  --source examples/source-criteria-challenges/main.gooo.fixture \
  --entry Select \
  --cases examples/source-criteria-challenges/evaluation.json \
  --out /tmp/gooo-shared-answer-example
```

These are finite observations of one deliberately insufficient candidate set.
Source formatting, IR round-trip, package lowering, caller retention, changed
predicate dependencies and saved-result replay have focused regression coverage.
The feature uses source-owned conditions and deterministic question selection.
The small decision model's existing input format remains unchanged.
