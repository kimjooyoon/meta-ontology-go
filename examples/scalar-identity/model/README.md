# Own graph chooser fixture

These are unchanged metadata and packed weights from the public
[ecosystem workbench](https://github.com/kimjooyoon/gooo-ecosystem-workbench/tree/59bc4118a5c4b50843412ed793d61373b065183f/models/graph-chooser-20261008/all-data-demonstration/qat_ternary).
The model was trained for source value graphs and Korean/English intents over
three binary record-field choices. This copy makes the scalar example reproducible
from one compiler checkout. It adds no training or model service.

- Metadata SHA256: `3c68205a660695103712e2115ce90accd298b406dced8ad5ad5c9b84bab0a202`.
- Packed weights SHA256: `76f68845a03ed8c8bc261a57c919e96dcdd352d3c35bcdaeb3d9b51a88975c6f`.
- Weights file: 446 bytes; the existing runtime reports 2,096 resident tensor bytes.
- Arithmetic: `float32_separate_v1`; ternary weights use five base-3 values per byte
  in storage, alongside float32 biases. Packed file size and runtime RSS measure
  different things.

The initial scalar example observation reached 12/12 authored field checks after
six candidates, from 7/12 on the first candidate. Fixed order tried eight. Both
selected identical native code and passed four different input cases; saved replay
made zero new inference calls. These counts describe the supplied example.
Model-training exposure of those inputs is unknown. A one-field example declined
to deterministic selection because this model requires three field choices.

The source graph, actual model call and exact native output are checked again
by the release profile. Release binaries contain the compiler; this model fixture
is available with the source example.
