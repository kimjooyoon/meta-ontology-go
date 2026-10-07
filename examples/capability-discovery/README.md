# Scoped capability discovery

The example asks what Gooo can do for a billing source and measures declared
scope against a separate Gooo domain contract. The contract expects four
declarations; the current source covers the invoice entity, receipt entity and
typed `Issue` activity while omitting `Reviewer`.

```sh
go run ./cmd/gooo discover --json \
  --query "How can I issue a billing receipt?" \
  --domain-contract examples/capability-discovery/domain-contract.gooo.fixture \
  examples/capability-discovery/current.gooo.fixture
```

`declaration_coverage` should report `PROGRESS`, 3/4, and the uncovered stable
ID `billing://reviewer`. The example renames `Invoice` to `Bill` in the target
while keeping the same stable ID and type, so presentation changes do not lower
coverage. The measure does not assess activity-body correctness, generation,
runtime behavior, or philosophical completeness.

## From discovery to generated code

`generation-domain.gooo.fixture` declares two expected integer activities.
Attach the saved `ClampNegativeToZero` result from the source-owned search
example to observe generation coverage **1/2**, with `Pending` still uncovered.
The [discovery guide](../../docs/language/capability-discovery.md#connect-an-existing-generation)
contains both commands. Replaying the projection uses no model or native build;
the separate `body-execute` command records native execution evidence.

`runtime-cases.json` supplies six new inputs for the same generated activity.
Adding `--execute-cases runtime-cases.json` to discovery (with the appropriate
path from the repository root) performs a fresh native build and two executions,
then links their outputs to the discovery receipt. Expected results are
generation **1/2**, independent inputs **6/6**, and reverse observation **6/6**.
See [the full command](../../docs/language/capability-discovery.md#observe-the-generated-program).
