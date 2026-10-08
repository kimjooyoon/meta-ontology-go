# Assemble a result from prepared local values

Build the current development compiler with Go 1.27.1. This example needs the
unread-local semantics added after the published 0.6.7-dev binary.

`PlanRetry` computes a retry decision, a capped delay and a reason. Its baseline
returns `false`, `0` and `pending`; three declared alternatives can connect the
prepared locals to those output fields. All eight combinations are valid typed
candidates, including combinations that leave some locals unread. Initialization
and later assignments remain part of execution.

The policy doubles a nonnegative delay up to its cap, uses one millisecond when
starting from zero with a positive cap, and stops on success, permanent failure
or an exhausted attempt count. Negative numeric input produces `invalid-input`.
This is a calculation example; it performs no network calls or sleeping.

From the repository root:

```sh
go build -o /tmp/gooo-candidate-locals ./cmd/gooo
/tmp/gooo-candidate-locals body-compose \
  --source examples/candidate-locals/retry.gooo.fixture \
  --cases examples/candidate-locals/cases.json --out /tmp/gooo-retry-locals
```

Use a new output directory. An optional `--model /path/to/shared-qat/model.json`
lets the compatible compact shared record judge rank the same choices. Without
a model the candidate order is deterministic. The five source cases contain
fifteen field expectations; the twelve native cases separately cover completion,
permanent failure, attempt limits, invalid counts, zero delays and int64 caps.
Keep their denominators separate when reading a result.

Replay the saved selection without new model calls:

```sh
/tmp/gooo-candidate-locals body-compose \
  --source examples/candidate-locals/retry.gooo.fixture \
  --cases examples/candidate-locals/cases.json \
  --composition /tmp/gooo-retry-locals/composition.json
```

Go output contains `_ = local` only when a particular binding is unread. The
selected Gooo body keeps its declarations and assignments. A division by zero in
an executed initializer still fails; an initializer inside an unentered branch
is not evaluated. Closed syntax, typing and candidate budgets continue to apply.

Receipts created by older compilers may record an unused local as
`TYPECHECK_FAILED`. Replaying that history with the new semantics can produce a
different attempt outcome and reject the old receipt. Preserve the original
compiler for historical replay; generate a new receipt for this source version.
The original workbench retry experiment and its published observations remain
separate from this new example.
