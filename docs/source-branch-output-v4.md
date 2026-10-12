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
GOWORK=off GOTOOLCHAIN=go1.27.2 go run ./cmd/gooo body-codegen \
  --json --activity Charge examples/body-codegen/branch-owned-output-v4.gooo.fixture
```

A separately trained V4 artifact can be inspected and used with the same source:

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

No fitted V4 weights or V4 quality measurements are published with this change.
The [retained diagnosis](https://github.com/kimjooyoon/gooo-decision-runtime/tree/a7cf509bb7e6c54b5b9165817dfe9a51dd5e27a1/studies/factorized-input-diagnosis-20261012)
found a branch-layout gap in the earlier neutral model. It motivated this explicit
ownership channel. A new study must publish fresh tasks, splits and fit settings
before learning or prediction. The [public model history](https://huggingface.co/asketeddy/gooo-canonical-decision-tiny-v1)
keeps earlier outcomes and failures available for comparison.

## Existing Go and compatibility certificates

Ordinary `generate --previous-go` continues to preserve caller-owned Go slots.
The current implementation-successor certificate does not bind previous Go
content. Combining it with `--previous-go` now gives an actionable usage error
before writing output; omit the certificate to regenerate with the existing Go.
This prevents a successful compatibility report from silently ignoring that input.
