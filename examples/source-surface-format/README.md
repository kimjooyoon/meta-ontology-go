# Retain imports, computations and assembly during formatting

This example contains two distinct source observations. Neither evaluates an
activity, model or finite case, and neither writes its input file.

## Earlier nominal path

`main.gooo.fixture` has one import, two computation bodies and one binding. The
[saved earlier output](baseline-format.json) reported success while omitting
those elements. [Its observation](baseline-observation.json) retains that result
and the actual earlier binary identity. That fixture has not been replayed.

## Complete source path

`retained.gooo.fixture` is a fresh, richer source with a Korean import alias,
qualified signatures, four computation bodies, one binding and one assembly
specification with fixed criteria and separate source inputs.

```sh
gooo format --json retained.gooo.fixture
```

The command needs development source containing this change.
[The proposal](format.json) and [formatted source](formatted.gooo.fixture) retain
the import, all four bodies, binding, assembly and all four qualified signature
references. `[9]` and `[13]` remain source inputs; their results were not evaluated.
[The observation](observation.json) records the producer and output digests.
The whole parsed carrier matched before and after in this source-format path.

Program executions, model predictions and training runs are zero. Native
equivalence, imported type resolution and model quality are outside this format
observation. The declaring-file type resolution has its
[own distinct source inspection](../qualified-signature-types/README.md).

See [the source preservation guide](../../docs/source-preserving-format.ko.md).
