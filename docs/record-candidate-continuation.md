# Continue after a record combination fails type checking

Gooo records a combination's type-check failure and continues to the next
candidate within the declared budget. A candidate that leaves a local unread
now remains valid: the compiler preserves the local's initialization and writes.

The [runnable example](../examples/body-codegen/record-candidate-continuation.gooo.fixture)
uses `saved.title` and `saved.state` as two baseline expressions. Replacing both
leaves `saved` unread. That combination is evaluated normally. A later
combination retains the title expression and
produces all fifteen expected selection fields.

```sh
gooo body-codegen --json --activity Select \
  examples/body-codegen/record-candidate-continuation.gooo.fixture
gooo body-compose \
  --source examples/body-codegen/record-candidate-continuation.gooo.fixture \
  --cases examples/body-codegen/record-field-updates-cases.json \
  --out /tmp/gooo-record-continuation
```

The deterministic example attempts seven combinations and selects mask6.
Mask3 now matches two of five cases and twelve of fifteen fields.
The selected candidate separately records fifteen of fifteen selection fields.
With `attempts "4"`, that combination consumes the fourth attempt and the
best valid result remains partial at twelve of fifteen fields.

The published 0.6.7-dev compiler treated mask3 as `TYPECHECK_FAILED` because Go
rejected its unread binding. Those historical receipts retain that observation.
The newer source semantics can reject replay of an old attempt history whose
outcome differs; use the original compiler for historical replay or create a new
receipt. See [candidate locals](../examples/candidate-locals/README.md).

Every new record assembly receipt includes `attempt_budget`, copied from the
source `attempts` declaration. It is independent of the number of ranked
combinations and the number actually attempted: eight available combinations
can have a budget of one, three, eight or sixteen. This observation is present
even when no assembly policy or model is connected. Tools can therefore use the
recorded source limit when deciding whether construction can continue.

Replay checks a present budget against the source declaration. Earlier receipts
that omit the field still replay using the source's limit; resuming one writes
the explicit limit in the new receipt and preserves the predecessor unchanged.
As with the other construction observations, the recorded limit is bound to the
receipt digest. A budget alone does not prove that a candidate remains: tools
must also inspect the attempt history, ranking and observed results.

An optional model still predicts once to order the candidates. Failed candidates
do not trigger another prediction. Saved selection replay reconstructs the same
attempts with zero new predictions. Historical prediction counts stay attached
to the original generation observation.

This continuation handles Go type-check errors arising from a combination of
preflighted alternatives. An invalid source or individual alternative still
fails during preflight. Cancellation and other generator errors stop the call.
If the attempt budget contains no valid typed candidate, the command explains
that no implementation was selected.

[Sequential field updates](record-field-updates.md)
· [Source value flow](record-value-flow.md)
