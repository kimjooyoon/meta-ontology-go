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
