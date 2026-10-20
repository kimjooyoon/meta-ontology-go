# V4: keep the parts of each return branch together

This compiler checkout connects SDK 0.2.46's optional V4 decision profile to
`body-context`, `body-codegen` and the shared body-construction session. The public
compiler 0.6.34 still uses SDK 0.2.44; these commands need this newer compiler
source until a compiler release includes the change.

Suppose an activity should divide a small input by 9 and add 23 to a larger input.
The source provides both expressions, a threshold, possible changes to operand
order and branch placement, and examples of the required result. V4 gives the
model a separate view of the literals in each return branch. It is like keeping
the labels on two trays of parts before assembling a function.

The model receives source operators and operand order, every declared integer
output case, and every Boolean condition case. Its new array keeps each branch's
literal relationships separate across the three predicate regions. Preparing
those arrays executes no candidate body and reads no held-out evaluation result.
Intent text remains in the authored source and explicit context export; its model
input cells are zero in this profile.

## Use the source-owned example

The [example](../examples/body-codegen/branch-owned-output-v4.gooo.fixture) declares
six output cases, three Boolean cases and four possible combinations of two
binary choices. From this checkout, deterministic construction is:

```sh
GOWORK=off GOTOOLCHAIN=go1.27.3 go run ./cmd/gooo body-codegen \
  --json --activity Charge examples/body-codegen/branch-owned-output-v4.gooo.fixture
```

The [published research artifact](https://huggingface.co/asketeddy/gooo-canonical-decision-tiny-v1/blob/8d5c1c2310017ec5d147b77b891bce894aaa03e4/research/branch-owned-learning-20261012/result/v4-branch-owned/model.json)
can be downloaded and explicitly selected with this newer compiler. Save that
file as `/path/to/v4-model.json` in the commands below:

```sh
gooo body-context --activity Charge --model /path/to/v4-model.json \
  examples/body-codegen/branch-owned-output-v4.gooo.fixture
gooo body-codegen --json --activity Charge --path-model /path/to/v4-model.json \
  --path-step-attempts 1 examples/body-codegen/branch-owned-output-v4.gooo.fixture
```

The artifact schema is `gooo/source-relative-branch-output-model/v4`. The context
report carries `gooo/compiler-source-relative-contract-model-context/v4`, exact
input digests, the pooled output view, and `branch_outputs` with each branch's
literal set, exact counts and 96 array values. An incompatible explicit input
version is rejected before selection. Existing model formats keep their encoders.

The session makes one prediction before checking any candidate. Each subsequent
step checks another source-owned candidate without another prediction. Omitting
the model keeps deterministic construction. Unsupported source representations
retain an explicit decline reason and the deterministic route. A malformed
supplied artifact produces an error.

## What is known so far

The input remains 528 float32 values per choice. The model has 19,034 FP32
parameters and 76,136 raw tensor bytes. Process memory, CPU use and latency also
include the compiler and are not established by those array sizes.

The compiler regression uses new authored cases and zero weights. It checks the
two branch literal sets, all declared rows, zero-call inspection, actual input
digest agreement, one prediction followed by finite construction, deterministic
omission and format mismatch handling. Completion covers the six declared output
cases and three Boolean cases; it is not a general correctness percentage.

## The first fitted comparison regressed

A [preregistered paired study](https://github.com/kimjooyoon/gooo-decision-runtime/tree/1f2ad49255d825b61dc696b668af5a4b58954984/studies/branch-owned-learning-20261012)
now publishes both fitted artifacts, training histories and original per-row
outcomes. It prepared 1,536 fresh Gooo documents from 96 goal contracts in 16
operator-pair families. Each profile fitted once on the same 768 training rows;
384 evaluation rows represent 24 contracts in four held-out operator families.

| Evaluation observation | V3 prose-neutral | V4 branch-owned |
| --- | ---: | ---: |
| First candidate meets all 8 output and 3 Boolean requirements | 112/384 (29.17%) | 60/384 (15.63%) |
| Candidate checks until finite completion | 960 | 1,016 |
| Complete after checking the four-candidate palette | 384/384 | 384/384 |
| Contracts complete on the first choice across all 16 sibling variants | 1/24 | 0/24 |

A fixed first candidate completes 96/384 rows. V4 therefore underperforms that
simple baseline on this cohort. The separate probes after checked completion
pass 1,536/1,536 observations, covering 96 unique contract/input pairs. Sibling
layouts, forms and languages repeat contracts; the row totals are not counts of
independent tasks or experimental methods. Both profiles leave intent text cells
zero, so Korean/English agreement measures identical inputs rather than language
understanding.

The array design preserves more explicit branch ownership, but this study does
not show a quality improvement. Both profiles also have low first completion on
their training rows (240/768 and 196/768). These observations leave the source and
requirement interaction at the decision head as a question for a fresh study;
they do not identify a single cause. The default ordering remains deterministic.

The study used the public compiler library and SDK 0.2.46; it did not measure a
released compiler executable with V4, native behavior, latency, CPU use, process
memory or quantization. Its fitted-task collection and predictions ran once. The
repository's retained-results check reads the saved corpus and observations and
performs no fitting, prediction or candidate execution.

The [retained diagnosis](https://github.com/kimjooyoon/gooo-decision-runtime/tree/a7cf509bb7e6c54b5b9165817dfe9a51dd5e27a1/studies/factorized-input-diagnosis-20261012)
found a branch-layout gap in the earlier neutral model. It motivated this explicit
ownership channel. Subsequent learning studies must publish fresh tasks, splits
and fit settings before learning or prediction. The
[public model history](https://huggingface.co/asketeddy/gooo-canonical-decision-tiny-v1)
keeps the earlier outcomes and this regression available for comparison.

## Existing Go and compatibility certificates

Ordinary `generate --previous-go` continues to preserve caller-owned Go slots.
The current implementation-successor certificate does not bind previous Go
content. Combining it with `--previous-go` now gives an actionable usage error
before writing output; omit the certificate to regenerate with the existing Go.
This prevents a successful compatibility report from silently ignoring that input.
