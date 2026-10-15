# Read implementation choices and Gooo requirements together

`main.gooo.fixture` declares two subtraction orders, a fixed Boolean condition,
and the fixed `Delta` helper used by that condition. The optional source export
keeps their typed structure together before any prediction or candidate execution.

```sh
go run ./cmd/gooo body-context --value-flow --activity Fee \
  examples/criteria-source-flow/main.gooo.fixture
```

`context.json` records this command's source export with a development compiler.
`observation.json` identifies the source, binary, production overlay and output.
The export has zero model predictions and zero candidate tests. Finite criteria
inputs stay in the original Gooo source; `--include-plan` would export their
separate declarations too.

The existing scalar model route still reports
`SCALAR_CRITERIA_FEATURE_CONTRACT_UNAVAILABLE`. The new graph is a source data
contract; learned weights and a model feature contract need a separate update.

See [the detailed guide, 한국어](../../docs/criteria-source-flow.ko.md).
