# Split one Gooo source used by two workspaces

`main.gooo` adds seven in `Credit`, then subtracts three in `Run`. Two workspace
definitions read this file, with different entry activities. Moving `Run` into
another file must update both source lists, even though `Credit` can still run
without that declaration. The complete package meaning matters beyond its entry.

From the compiler checkout, run this example in a small temporary copy. These
commands require the development compiler containing `--consumer-workspace`:

```sh
split_example=$(mktemp -d)
cp examples/shared-source-consumers/* "$split_example/"
go run ./cmd/gooo package split-source \
  --package shared/app --source main.gooo --before urn:shared:run --into run.gooo \
  --consumer-workspace "$split_example/credit.workspace.json" \
  --json "$split_example/run.workspace.json" > "$split_example/proposal.json"
go run ./cmd/gooo package split-source --apply "$split_example/proposal.json" \
  --json "$split_example/run.workspace.json" > "$split_example/applied.json"
go run ./cmd/gooo package execute --json --cases "$split_example/run-cases.json" \
  "$split_example/run.workspace.json" > "$split_example/run-result.json"
go run ./cmd/gooo package execute --json --cases "$split_example/credit-cases.json" \
  "$split_example/credit.workspace.json" > "$split_example/credit-result.json"
```

The proposal writes nothing. Its `consumers` array retains each additional
manifest's exact digest, updated definition, source origins and independent
package image. Repeat `--consumer-workspace` for further declared consumers,
up to 64 workspaces including the primary. Source and destination addresses are
derived relative to each manifest; every supplied workspace must declare that
same source occurrence within its root. Equal bytes at another address do not
make it a consumer of this change.

Apply reads the scope from the saved proposal. It checks all current inputs,
stages the source pair and both manifests, then checks every projected package
through the real loader before publishing. The source pair is written once.
A staging/publication failure restores prior outputs where the existing writer
can recover; failed recovery remains an error with its actual state. After a
successful write, all supplied workspace images are read again. The receipt's
`OUTPUT_VERIFIED` state covers this declared group. Applying the old proposal
again fails because the source and source lists have changed.

The two execution files report four and two expected activity outputs,
respectively, and native replay. The split itself does not execute activities or
use a model. The example's finite outputs are an independent check of these two
entry contexts, not a proof for every input.

This command does not discover omitted workspaces, embedded Go readers or syntax
registries. Additional paths are explicit, and the saved proposal retains those
addresses. The filesystem checks are sequential; multi-file publication is
rollback-capable, not a crash-atomic or concurrent-writer transaction. The wider
repository executor still has to connect its discovered consumers and obligations.
