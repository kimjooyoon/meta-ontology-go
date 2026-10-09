# Preserve operators and ordered values in model input

The explicit `triple_record_value_graph_v3_shared_v1` contract sends a bounded
source value graph to the small field-choice model. A filename classifier had
eight arrangements of three choices, but the original expression features kept
only two different arrays. Following helper values with the v2 ancestor counts
raised that to four; AND and OR still shared an input. The graph contract keeps
those source distinctions before any prediction or finite case execution.

```sh
go run ./cmd/gooo body-context --value-flow --activity Classify \
  --feature-version triple_record_value_graph_v3_shared_v1 \
  examples/text-operations/source.gooo.fixture
```

The [0.6.9 development source](releases/0.6.9-dev.md) requires SDK v0.2.26-experimental.
The earlier Gooo 0.6.8 binary does not contain this option. The context alone performs zero
model predictions and zero candidate tests. The original full value-flow graph
is shown only with `--value-flow`; its digest is retained in the model receipt.

## What reaches the model

The compiler derives root parameter positions and primitive types, stable field
IDs, ordered operands, canonical source literals, conditional joins and execution
guards. It follows same-source helpers, including helpers called only by a
permitted alternative. Literal spans are resolved in their own root body,
alternative expression or helper body before canonicalization. Integer literals
retain exact values, and string literals retain their decoded contents.

The SDK graph omits local names, source offsets and callee display names from
features. The compiler's source receipt still retains those identities and
provenance. Copies and reads retain their underlying value fingerprint; guards
remain represented. This lets a helper rename preserve its model array while a
changed returned literal can change it. Input cases, expected outputs and attempt
budgets do not enter that array.

Each of the three fields uses 256 floats: two ordered alternatives with 96 slots
each, plus 64 intent slots. An alternative has 32 explicit root-operation slots,
32 reachable-operation counts and 32 hashed ordered-relation slots. All seven
channels normalize independently. The model layout stays 256/8/2, with a shared
judge for the three fields. Boolean, integer and string input types are explicit.
Optional values have no graph-model representation yet and cause a recorded
decline to deterministic candidate order.

The complete input is bounded to 512 nodes and 64 KiB; each string has a 1,024-byte
UTF-8 bound. The filename example exports 115 nodes and a 6,299-byte input on the
current fixture. These bounds describe source preparation. Prepared prediction
uses the SDK's existing 768-float array and caller-owned inference workspace.

## Construction, replay and limits

An explicitly trained graph-model file can be passed to `body-codegen
--path-model` or `body-compose --model`. The compiler makes one initial prediction
from source context, then evaluates candidates against the separately declared
cases. Unsupported or oversized context records a reason and uses the ordinary
deterministic order. Saved construction replay rebuilds and checks the context
without another prediction. Existing v1/v2 model files keep their original paths.

Integration tests use synthetic FP32/PTQ/QAT weights to check dispatch, finite
continuation and replay. A native composition test uses all-zero weights: the
first candidate fails, all eight are considered, and the selected code produces
14 expected activity outputs across seven input cases. Executing the saved
composition repeats those outputs with zero new model calls. Those tests verify
the compiler/runtime path; they do not measure learned quality.

All eight arrangements of the filename fixture produce distinct complete arrays
in the source export regression. The compact relation and intent channels still
use hashes and can collide on other programs. The intent channel is smaller
than v2's, so Korean/English wording and separate program families need new
measurements. The judge still scores fields independently. Those integration
tests used synthetic weights. The current [scalar example](language/scalar-identity.md)
includes the existing own QAT graph chooser as a separate reproducible fixture.
Its provenance and finite limitations are documented alongside the model.
The published v1 model keeps its own contract.

For a local record-field model, [check source compatibility](record-model-preflight.md)
with `body-context --model` before assembly. The compiler picks the verified
model's feature and reports representation declines with zero predictions.

[Value-flow provenance and helper scope](record-value-flow.md)
· [Field assembly and finite completeness](record-field-assembly.md)
· [SDK v0.2.26](https://github.com/kimjooyoon/gooo-decision-runtime/releases/tag/v0.2.26-experimental)
