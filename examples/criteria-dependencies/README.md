# See a criterion's source dependencies

`Fee` permits two operand orders and declares the fixed `ceiling` criterion.
`Allowed` delegates to `Within`, which reads `Ceiling`. The source dependency view
connects those activities by semantic ID, preserving which criterion owns the path.

```sh
gooo body-context --value-flow --activity Fee main.gooo.fixture
```

The declaration includes separate selection and held inputs. Reading dependencies
does not evaluate either input. A source path says which rule depends on which
helper; actual pass/fail outcomes need their own execution observation.

See [the scope and output fields](../../docs/source-criteria-dependencies.md).

## Recorded source inspection

The command was run once on this fresh example with an empty PATH. The saved
[context](context.json) contains three dependency edges, two helper activities and
a call budget of two. The fixed predicate's symbolic graph contains 14 nodes.
The source file remained unchanged.

[The observation](observation.json) retains the actual development binary and
source overlays. It records one source inspection and zero program executions,
candidate tests, model predictions and training runs. The authored selection and
held inputs were not evaluated. Earlier source examples were not replayed.
