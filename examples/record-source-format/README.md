# Record source formatting

This source example keeps a V4 record declaration beside an import, two activity
bodies and a binding. The formatter uses the same explicit scalar-field grammar
as package source execution and compares the complete parsed source carrier.

```sh
gooo format --json main.gooo.fixture
```

The command returns formatted source without writing the input. It does not
resolve the imported package, lower field types, execute record values or invoke
a decision model. Field identities, spelling, type references, presence,
cardinality and ordering are compared before a change proposal is returned.

The development source and CLI compiled. One fresh invocation preserved all3
field carriers (names, IDs, type spelling, presence, cardinality and order),
1import,2compute bodies and1binding. The input was unchanged; direct writes0,
program executions0, model predictions0, training0. The retained producer binds
the actual base14d3b100 and19 source overlays to the new CLI binary. It is not
rebound to a future accepted commit. Earlier examples were not replayed.

[observation.json](observation.json) records the source scope and exact producer.
[format.json](format.json) and [formatted.gooo.fixture](formatted.gooo.fixture)
contain the observed command output.
